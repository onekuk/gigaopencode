package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gitverse.ru/kmpavloff/openai-provider-gigachat/gigachat"
)

type TokenManager struct {
	config      *Config
	httpClient  *http.Client
	accessToken string
	expiresAt   int64
	mutex       sync.RWMutex
	logger      *Logger
}

func NewTokenManager(config *Config, logger *Logger) *TokenManager {
	// Create HTTP client with TLS config that skips certificate verification
	// This is needed for GigaChat API which uses self-signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &TokenManager{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   5 * time.Minute,
		},
	}
}

func (tm *TokenManager) GetAccessToken() (string, error) {
	tm.mutex.RLock()

	// Проверяем, есть ли действующий токен
	now := time.Now().UnixMilli()
	if tm.accessToken != "" && tm.expiresAt > now+60000 { // за минуту до истечения
		expiresIn := time.Duration(tm.expiresAt-now) * time.Millisecond
		tm.logger.Debug("Using cached access token, expires in %v", expiresIn)
		token := tm.accessToken
		tm.mutex.RUnlock()
		return token, nil
	}

	tm.mutex.RUnlock()

	// Нужно обновить токен
	tm.logger.Info("Access token expired or missing, refreshing...")
	return tm.refreshToken()
}

func (tm *TokenManager) refreshToken() (string, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	requestStart := time.Now()

	// Двойная проверка после получения блокировки
	now := time.Now().UnixMilli()
	if tm.accessToken != "" && tm.expiresAt > now+60000 {
		tm.logger.Debug("Token was refreshed by another goroutine")
		return tm.accessToken, nil
	}

	tm.logger.Debug("Requesting new access token from %s", tm.config.OAuthURL)

	// Подготовка данных для запроса
	data := url.Values{}
	data.Set("scope", tm.config.Scope)

	req, err := http.NewRequest("POST", tm.config.OAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		tm.logger.Error("Failed to create OAuth request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Установка заголовков
	rqUID := generateRqUID()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RqUID", rqUID)
	req.Header.Set("Authorization", "Basic "+tm.config.AuthorizationKey)

	tm.logger.Debug("OAuth request: POST %s, RqUID: %s, Scope: %s", tm.config.OAuthURL, rqUID, tm.config.Scope)

	// Выполнение запроса
	resp, err := tm.httpClient.Do(req)
	requestDuration := time.Since(requestStart)

	if err != nil {
		tm.logger.Error("OAuth request failed after %v: %v", requestDuration, err)
		return "", fmt.Errorf("failed to make OAuth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tm.logger.Error("OAuth request failed with status %d after %v", resp.StatusCode, requestDuration)
		return "", fmt.Errorf("OAuth request failed with status %d", resp.StatusCode)
	}

	// Парсинг ответа
	var tokenResp gigachat.Token
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		tm.logger.Error("Failed to decode OAuth response: %v", err)
		return "", fmt.Errorf("failed to decode OAuth response: %w", err)
	}

	// Сохранение токена
	tm.accessToken = tokenResp.AccessToken
	tm.expiresAt = tokenResp.ExpiresAt

	expiresIn := time.Duration(tokenResp.ExpiresAt-time.Now().UnixMilli()) * time.Millisecond
	maskedToken := maskToken(tokenResp.AccessToken)
	tm.logger.Info("Successfully refreshed access token in %v, expires in %v, token: %s", requestDuration, expiresIn, maskedToken)

	return tm.accessToken, nil
}

// Генерация уникального идентификатора запроса
func generateRqUID() string {
	return uuid.New().String()
}
