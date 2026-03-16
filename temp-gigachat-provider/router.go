package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func NewRouter(handlers *HTTPHandlers, logger *Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/completions", loggingMiddleware(logger, corsMiddleware(authMiddleware(handlers.HandleCompletions))))

	mux.HandleFunc("/v1/chat/completions", loggingMiddleware(logger, corsMiddleware(authMiddleware(handlers.HandleChatCompletions))))

	mux.HandleFunc("/v1/embeddings", loggingMiddleware(logger, corsMiddleware(authMiddleware(handlers.HandleEmbeddings))))

	mux.HandleFunc("/v1/models", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"GET": handlers.HandleListModels,
	})))))

	mux.HandleFunc("/v1/models/", loggingMiddleware(logger, corsMiddleware(authMiddleware(modelIDHandler(handlers)))))

	mux.HandleFunc("/v1/files", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"GET":  handlers.HandleListFiles,
		"POST": handlers.HandleUploadFile,
	})))))

	mux.HandleFunc("/v1/files/", loggingMiddleware(logger, corsMiddleware(authMiddleware(fileIDHandler(handlers)))))

	mux.HandleFunc("/v1/fine_tuning/jobs", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateFineTuningJob,
		"GET":  handlers.HandleListFineTuningJobs,
	})))))

	mux.HandleFunc("/v1/fine_tuning/jobs/", loggingMiddleware(logger, corsMiddleware(authMiddleware(fineTuningJobHandler(handlers)))))

	mux.HandleFunc("/v1/images/generations", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateImage,
	})))))

	mux.HandleFunc("/v1/images/edits", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateImageEdit,
	})))))

	mux.HandleFunc("/v1/images/variations", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateImageVariation,
	})))))

	mux.HandleFunc("/v1/audio/transcriptions", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateTranscription,
	})))))

	mux.HandleFunc("/v1/audio/translations", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateTranslation,
	})))))

	mux.HandleFunc("/v1/audio/speech", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateSpeech,
	})))))

	mux.HandleFunc("/v1/moderations", loggingMiddleware(logger, corsMiddleware(authMiddleware(methodHandler(map[string]http.HandlerFunc{
		"POST": handlers.HandleCreateModeration,
	})))))

	return mux
}

// responseWriter wrapper для перехвата статуса и тела ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func loggingMiddleware(logger *Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Читаем тело запроса
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Логируем входящий запрос
		logger.Info("→ %s %s | Headers: %v", r.Method, r.URL.Path, r.Header)
		if len(requestBody) > 0 {
			logger.Debug("→ OpenAI Request Body: %s", string(requestBody))
		}

		// Оборачиваем ResponseWriter для перехвата ответа
		rw := newResponseWriter(w)

		// Вызываем следующий обработчик
		next(rw, r)

		// Логируем ответ
		duration := time.Since(start)
		logger.Info("← %s %s | Status: %d | Duration: %v", r.Method, r.URL.Path, rw.statusCode, duration)
		if rw.body.Len() > 0 {
			logger.Debug("← OpenAI Response Body: %s", rw.body.String())
		}
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

type contextKey string

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Authorization header is optional - some clients (like OpenCode Desktop 1.2.24) don't send it
		// but we still accept it if present for compatibility
		if authHeader != "" && !strings.HasPrefix(authHeader, "Bearer ") {
			writeErrorResponse(w, http.StatusUnauthorized, "invalid_authorization", "Authorization header must start with 'Bearer '")
			return
		}

		// Add auth token to context using a custom key type
		if authHeader != "" {
			ctx := context.WithValue(r.Context(), contextKey("auth_token"), authHeader)
			r = r.WithContext(ctx)
		}

		next(w, r)
	}
}

func methodHandler(methods map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler, exists := methods[r.Method]
		if !exists {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
			return
		}
		handler(w, r)
	}
}

func modelIDHandler(handlers *HTTPHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modelID := extractIDFromPath(r.URL.Path, "/v1/models/")
		if modelID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Model ID is required")
			return
		}

		switch r.Method {
		case "GET":
			handlers.HandleRetrieveModel(w, r)
		case "DELETE":
			handlers.HandleDeleteModel(w, r)
		default:
			writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		}
	}
}

func fileIDHandler(handlers *HTTPHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		fileID := extractIDFromPath(path, "/v1/files/")
		if fileID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "File ID is required")
			return
		}

		if strings.HasSuffix(path, "/content") {
			if r.Method == "GET" {
				handlers.HandleRetrieveFileContent(w, r)
			} else {
				writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
			}
			return
		}

		switch r.Method {
		case "GET":
			handlers.HandleRetrieveFile(w, r)
		case "DELETE":
			handlers.HandleDeleteFile(w, r)
		default:
			writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		}
	}
}

func fineTuningJobHandler(handlers *HTTPHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		jobID := extractIDFromPath(path, "/v1/fine_tuning/jobs/")
		if jobID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Job ID is required")
			return
		}

		if strings.HasSuffix(path, "/cancel") {
			if r.Method == "POST" {
				handlers.HandleCancelFineTuningJob(w, r)
			} else {
				writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
			}
			return
		}

		if r.Method == "GET" {
			handlers.HandleRetrieveFineTuningJob(w, r)
		} else {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		}
	}
}

func extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	remaining := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remaining, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return ""
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Simple JSON encoding without external dependencies
	jsonResp := `{"error":{"message":"` + message + `","type":"` + errorType + `"}}`
	w.Write([]byte(jsonResp))
}
