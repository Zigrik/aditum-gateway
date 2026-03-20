package middleware

import (
	"log"
	"net/http"
	"strings"
)

// AnomalyLogger может писать в БД или просто в лог
var anomalyLog = log.New(log.Writer(), "ANOMALY: ", log.LstdFlags)

func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.ToLower(r.URL.Path)

		// Чёрный список паттернов
		suspicious := []string{
			".git", ".env", "wp-admin", "phpmyadmin",
			"admin", "backup", "sql", "union select",
			"../", "..\\", "/etc/passwd",
		}

		for _, pattern := range suspicious {
			if strings.Contains(path, pattern) {
				anomalyLog.Printf("Suspicious path from %s: %s", r.RemoteAddr, r.URL.Path)
				// Возвращаем 404, чтобы не раскрывать наличие системы
				http.NotFound(w, r)
				return
			}
		}

		// Добавляем стандартные заголовки безопасности
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, r)
	})
}
