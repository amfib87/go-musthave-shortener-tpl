// Package service предназначен для реализации сервисных функций
package service

import (
	"bufio"
	"context"
	crrand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/model"
	"github.com/amfib87/go-musthave-shortener-tpl/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const maxRetries = 5
const CookieName = "user_auth"

func InitMap(st URLStorage) (*model.StringMap, error) {
	stringMap := &model.StringMap{
		Data: make(model.TData),
	}

	if st.File != nil {

		reader := bufio.NewReader(st.File)
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed read data from file: %w", err)
		}

		if len(data) != 0 {
			if err := json.Unmarshal(data, &stringMap.Data); err != nil {
				return nil, fmt.Errorf("failed json Unmarshal: %v", err)
			}
		}

	} else if st.DB != nil {
		dataDB, err := model.ReadDB(st.DB)
		if err != nil {
			return nil, fmt.Errorf("failed read DB: %v", err)
		}
		stringMap.Data = dataDB
	}

	return stringMap, nil
}

func GetShortURL(ctx context.Context, data model.DataRow, m *model.StringMap, st URLStorage, lg *logger.TLog) (string, error) {
	const maxRetries = 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortURL := generateShortID()

		shortURLExist, err := m.InsertShortURL(ctx, data, shortURL, st.File, st.DB)
		if err == nil {
			return shortURL, nil
		}

		if errors.Is(err, model.ErrOriginalURLExist) {
			return shortURLExist, err
		} else if errors.Is(err, model.ErrKeyExists) {
			continue
		}

		return "", fmt.Errorf("unexpected error on insert")
	}

	return "", fmt.Errorf("failed to compose unique short URL after %d attempts", maxRetries)
}

func generateShortID() string {
	const Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 8)
	for i := range b {
		b[i] = Letters[rand.Intn(len(Letters))]
	}
	return string(b)
}

func InitFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func FileClose(file *os.File, lg *logger.TLog) {
	if err := file.Close(); err != nil {
		lg.Lg.Sugar().Infoln("failed close file: %v", err)
	}
}

func GetShortURLMass(ctx context.Context, values []model.DataRequestMass, m *model.StringMap, st URLStorage, userID string) ([]model.DataAnswerMass, error) {
	export := []model.DataAnswerMass{}
	shortKeys := make(model.TData)

	for _, lineData := range values {
		for attempt := 0; attempt < maxRetries; attempt++ {
			shortURL := generateShortID()

			// проверяем на наличие сгенерированного shorturl
			count, err := model.CheckExistShortURL(ctx, shortURL, st.DB)
			if err != nil {
				return nil, fmt.Errorf("CheckExistShortURL, err: %w", err)
			}
			if count != 0 {
				continue
			}

			shortKeys[shortURL] = model.DataRow{
				URL:    lineData.OriginalURL,
				UserID: userID}
			export = append(export, model.DataAnswerMass{CorrelationID: lineData.CorrelationID, ShortURL: shortURL})
			break
		}
	}

	if len(export) == 0 {
		return nil, fmt.Errorf("failed to compose unique short URL after %d attempts", maxRetries)
	}

	if err := m.InsertShortURLMass(ctx, shortKeys, st.DB, st.File); err != nil {
		return nil, fmt.Errorf("failed insertShortURLMass, err: %w", err)
	}

	return export, nil
}

type URLStorage struct {
	DB   *sql.DB
	File *os.File
}

func InitURLStorage(cfg *config.Cnfg, log *logger.TLog) (URLStorage, error) {
	URLstorage := URLStorage{}
	var err error

	if cfg.DataBaseDsn != "" {
		URLstorage.DB, err = repository.InitDB(cfg.DataBaseDsn)
		if err != nil {
			log.Lg.Sugar().Fatalf("failed InitDB: %v", err)
			return URLstorage, err
		}

	} else if cfg.StoragePath != "" {
		URLstorage.File, err = InitFile(cfg.StoragePath)
		if err != nil {
			log.Lg.Sugar().Fatalf("failed to init file: %v", err)
			return URLstorage, err
		}
	}

	return URLstorage, nil
}

func (st URLStorage) Close(log *logger.TLog) {
	if st.DB != nil {
		defer func() { _ = st.DB.Close() }()
	}
	if st.File != nil {
		defer FileClose(st.File, log)
	}

}

func DelShortURLs(shortURLs []model.ShortURL, userID string, st URLStorage, data *model.StringMap, log *logger.TLog) {
	const batchSize = 10
	ch := make(chan []string, batchSize)

	go func() {
		defer close(ch)
		var ids []string
		for _, shortURL := range shortURLs {
			ids = append(ids, string(shortURL)) // предполагается, что поле называется ShortID
		}

		for i := 0; i < len(ids); i += batchSize {
			end := i + batchSize
			if end > len(ids) {
				end = len(ids)
			}
			ch <- ids[i:end]
		}
	}()

	var wg sync.WaitGroup
	for batch := range ch {
		wg.Add(1)
		go func(b []string) {
			defer wg.Done()
			if err := model.MarkAsDeleted(b, userID, st.DB, data); err != nil {
				log.Lg.Error("Failed to mark batch as deleted: %v", zap.Error(err))
			}
		}(batch)
	}
	wg.Wait()
}

func GetUserIDContx(cont context.Context, key model.ContextKey) (userID string) {

	valUserID := cont.Value(key)
	if valUserID != nil {
		userID = valUserID.(string)
	} else {
		userID = "unknown"
	}

	return userID
}

func GenerateTLSCertificate(certFile, keyFile string) error {
	privateKey, err := rsa.GenerateKey(crrand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Заполняем информацию о сертификате
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"My Organization"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
		BasicConstraintsValid: true,
	}

	// Создаём самоподписанный сертификат
	derBytes, err := x509.CreateCertificate(crrand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Записываем сертификат в файл (PEM‑формат)
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %v", err)
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	// Записываем приватный ключ в файл (PEM‑формат)
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %v", err)
	}
	defer keyOut.Close()

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	log.Printf("Successfully created TLS certificate: %s", certFile)
	log.Printf("Successfully created private key: %s", keyFile)
	return nil
}

func IsIPInSubnet(ipStr, subnetStr string) bool {
	if subnetStr == "" {
		return false
	}

	_, subnet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	return subnet.Contains(ip)
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GetUserID(str string) (id string, err error) {

	token, err := jwt.ParseWithClaims(str, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(model.SecretKey), nil
	})

	if err == nil && token.Valid {
		claims, ok := token.Claims.(*Claims)
		if ok {
			return claims.ID, nil
		}

		if claims.UserID == "" { // Кука есть, но id пуст => возвращаем 401 Unauthorized
			return
		}
	}

	return "", err
}
