package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"aditum-gateway/balancer"
	"aditum-gateway/discovery"
	"aditum-gateway/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	consulAddr := getEnv("CONSUL_ADDR", "consul:8500")
	serverPort := getEnv("SERVER_PORT", "8080")
	rateLimit := getEnvAsInt("RATE_LIMIT", 100)
	rateWindow := getEnvAsInt("RATE_WINDOW", 60)

	// Создаём клиент Consul
	consulClient, err := discovery.NewConsulClient(consulAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Consul: %v", err)
	}

	// Балансировщик
	rr := &balancer.RoundRobin{}

	// Основной обработчик
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем имя сервиса из пути: /api/{service}/...
		pathParts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/"), "/", 2)
		if len(pathParts) < 1 || pathParts[0] == "" {
			http.Error(w, "Invalid service name", http.StatusBadRequest)
			return
		}
		serviceName := pathParts[0]

		// Получаем здоровые инстансы из Consul
		instances, err := consulClient.GetHealthyServices(serviceName)
		if err != nil {
			log.Printf("Error getting instances for %s: %v", serviceName, err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		if len(instances) == 0 {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		// Выбираем инстанс (round-robin)
		target := rr.Next(instances)
		if target == "" {
			http.Error(w, "No healthy instances", http.StatusServiceUnavailable)
			return
		}

		// Проксируем запрос
		targetURL, err := url.Parse("http://" + target)
		if err != nil {
			log.Printf("Invalid target URL: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		// Переписываем путь, убирая /api/{service}
		if len(pathParts) > 1 {
			r.URL.Path = "/" + pathParts[1]
		} else {
			r.URL.Path = "/"
		}
		r.Host = targetURL.Host // важно для корректной работы
		proxy.ServeHTTP(w, r)
	})

	// Health check для Gateway
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	var handler http.Handler = mux

	// Оборачиваем в middleware
	handler = middleware.Logger(handler)                                                      // логирование
	handler = middleware.Security(handler)                                                    // безопасность
	handler = middleware.RateLimit(rateLimit, time.Duration(rateWindow)*time.Second)(handler) // rate limit

	log.Printf("API Gateway starting on port %s", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, handler))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var val int
		fmt.Sscanf(v, "%d", &val)
		return val
	}
	return fallback
}
