package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/logger"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

type AuditEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type AuditSubscriber interface {
	Notify(event *AuditEvent) error
	Close() error
}

func NewAuditEvent(action, userID, url string) *AuditEvent {
	return &AuditEvent{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}

type FileAuditSubscriber struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileAuditSubscriber(filePath string) (*FileAuditSubscriber, error) {
	if filePath == "" {
		return nil, fmt.Errorf("filepath is empty")
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileAuditSubscriber{file: file}, nil
}

func (f *FileAuditSubscriber) Notify(event *AuditEvent) error {
	if f == nil {
		return fmt.Errorf("file is nill")
	}

	if f.file == nil {
		return fmt.Errorf("file is nill")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	_, err = f.file.WriteString(string(data) + "\n")
	return err
}

func (f *FileAuditSubscriber) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

type RemoteAuditSubscriber struct {
	client *http.Client
	url    string
}

func NewRemoteAuditSubscriber(url string) (*RemoteAuditSubscriber, error) {
	if url == "" {
		return nil, fmt.Errorf("url is empty")
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10

	standardClient := retryClient.StandardClient() // *http.Client
	standardClient.Timeout = 5 * time.Second

	return &RemoteAuditSubscriber{
			client: standardClient,
			url:    url},
		nil
}

func (r *RemoteAuditSubscriber) Notify(event *AuditEvent) error {
	if r == nil {
		return fmt.Errorf("r is nill")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := r.client.Post(r.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote audit failed: %s", resp.Status)
	}
	return nil
}

func (r *RemoteAuditSubscriber) Close() error {
	return nil
}

type AuditManager struct {
	subscribers []AuditSubscriber
	mu          sync.Mutex
}

func NewAuditManager(filePath, url string) (*AuditManager, error) {
	manager := &AuditManager{}

	// Добавляем файл-приемник, если указан
	if filePath != "" {
		fileSub, err := NewFileAuditSubscriber(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed NewFileAuditSubscriber: %w", err)
		}
		if fileSub != nil {
			manager.addSubscriber(fileSub)
		}
	}

	// Добавляем удаленный приемник, если указан
	if url != "" {
		remoteSub, err := NewRemoteAuditSubscriber(url)
		if err != nil {
			return nil, fmt.Errorf("failed NewRemoteAuditSubscriber: %w", err)
		}
		if remoteSub != nil {
			manager.addSubscriber(remoteSub)
		}
	}

	return manager, nil
}

func (a *AuditManager) NotifyAll(lg logger.TLog, event *AuditEvent) {
	if a == nil { // Проверка т.к. тест из-за паники не проходит
		return
	}

	if a.subscribers == nil {
		return
	}

	for _, sub := range a.subscribers {
		if err := sub.Notify(event); err != nil {
			lg.Lg.Error("audit notification failed for subscriber: %v", zap.Error(err))
		}
	}
}

func (a *AuditManager) Close() error {
	for _, sub := range a.subscribers {
		if err := sub.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (a *AuditManager) addSubscriber(subscr AuditSubscriber) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.subscribers = append(a.subscribers, subscr)
}

func (a *AuditManager) removeSubscriber(subscr AuditSubscriber) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, sub := range a.subscribers {
		if sub == subscr {
			a.subscribers = append(a.subscribers[:i], a.subscribers[i+1:]...)
			return true
		}
	}

	return false
}
