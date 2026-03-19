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
	"go.uber.org/zap"
)

type Audit struct {
	File     *os.File
	CondFile *sync.Cond
}

type AuditEvent struct {
	Ts     int64  `json:"ts"`
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
		Ts:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}

type FileAuditSubscriber struct {
	file *os.File
}

func NewFileAuditSubscriber(filePath string) (*FileAuditSubscriber, error) {
	if filePath == "" {
		return nil, nil
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
		return nil, nil
	}

	return &RemoteAuditSubscriber{client: &http.Client{Timeout: 5 * time.Second},
		url: url}, nil
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
}

func NewAuditManager(filePath, URL string) (*AuditManager, error) {
	manager := &AuditManager{}

	// Добавляем файл-приемник, если указан
	fileSub, err := NewFileAuditSubscriber(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed NewFileAuditSubscriber: %w", err)
	}
	if fileSub != nil {
		manager.subscribers = append(manager.subscribers, fileSub)
	}

	// Добавляем удаленный приемник, если указан
	remoteSub, err := NewRemoteAuditSubscriber(URL)
	if err != nil {
		return nil, fmt.Errorf("failed NewRemoteAuditSubscriber: %w", err)
	}
	if remoteSub != nil {
		manager.subscribers = append(manager.subscribers, remoteSub)
	}

	return manager, nil
}

func (a *AuditManager) NotifyAll(lg logger.TLog, event *AuditEvent) {
	if a == nil {
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
