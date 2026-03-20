# aditum-gateway
Динамический API Gateway на Go с интеграцией Consul.
Обеспечивает единую точку входа для микросервисов с балансировкой нагрузки, rate limiting и защитой от подозрительных запросов.

Возможности
Service Discovery — интеграция с Consul, автоматическое обнаружение здоровых инстансов сервисов.

Балансировка нагрузки — round-robin распределение запросов между экземплярами.

Rate Limiting — ограничение количества запросов с одного IP (настраивается).

Безопасность — детект подозрительных путей (.git, .env, wp-admin и др.), защитные HTTP-заголовки.

Маршрутизация — проксирование запросов по схеме /api/{service_name}/....

Логирование — запись всех запросов с временем выполнения.

Структура проекта
text
aditum-gateway/
├── main.go                 # Точка входа
├── balancer/               # Балансировщик нагрузки
│   └── roundrobin.go
├── discovery/              # Клиент Consul
│   └── consul.go
├── middleware/             # Middleware компоненты
│   ├── logger.go
│   ├── rate_limit.go
│   └── security.go
├── .env                    # Конфигурация
├── go.mod
├── go.sum
└── Dockerfile
Конфигурация (.env)
Переменная	Описание	По умолчанию
SERVER_PORT	Порт, на котором слушает Gateway	8080
CONSUL_ADDR	Адрес Consul (host:port)	consul:8500
RATE_LIMIT	Максимум запросов в минуту с одного IP	100
RATE_WINDOW	Окно для rate limit (секунд)	60
Запуск
Локально (без Docker)
bash
# Установка зависимостей
go mod download

# Запуск
go run main.go
В Docker
bash
# Сборка образа
docker build -t aditum-gateway .

# Запуск
docker run -d \
  -p 8080:8080 \
  -e CONSUL_ADDR=consul:8500 \
  --name gateway \
  aditum-gateway
С docker-compose
yaml
api-gateway:
  build: ./aditum-gateway
  container_name: gateway
  restart: unless-stopped
  ports:
    - "8080:8080"
  environment:
    - CONSUL_ADDR=consul:8500
    - SERVER_PORT=8080
    - RATE_LIMIT=100
    - RATE_WINDOW=60
  depends_on:
    - consul
  networks:
    - backend
Использование
Gateway проксирует запросы по пути /api/{service_name}/..., где {service_name} — имя сервиса, зарегистрированного в Consul.

Пример:

bash
# Запрос к сервису 'auth'
curl http://localhost:8080/api/auth/login

# Запрос к сервису 'chat'
curl http://localhost:8080/api/chat/messages
Health check
bash
curl http://localhost:8080/health
Регистрация сервисов в Consul
Каждый микросервис должен зарегистрироваться в Consul при запуске. Пример на Go:

go
import consul "github.com/hashicorp/consul/api"

registration := &consul.AgentServiceRegistration{
    ID:      "auth-1",
    Name:    "auth",
    Address: "auth",
    Port:    8081,
    Check: &consul.AgentServiceCheck{
        HTTP:     "http://auth:8081/health",
        Interval: "10s",
        Timeout:  "5s",
    },
}
client.Agent().ServiceRegister(registration)
Middleware
Gateway последовательно применяет три middleware:

Logger — записывает в лог IP, метод, путь и время выполнения.

Security — блокирует подозрительные пути, добавляет заголовки безопасности.

RateLimit — ограничивает частоту запросов с одного IP.

Требования
Go 1.20+

Consul 1.15+ (для работы с service discovery)

Лицензия
MIT
