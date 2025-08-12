package main

import (
	"context"
	"log"

	"github.com/skinkvi/money_managment/internal/config"
	"github.com/skinkvi/money_managment/pkg/logger"
)

/*
🏦 PERSONAL FINANCE TRACKER - ПЛАН РАЗРАБОТКИ
=============================================

Стек: Go (Gin), PostgreSQL, Docker/Docker Compose, Redis
Опционально: CI/CD (GH Actions), Swagger для API docs

📋 ЭТАПЫ РАЗРАБОТКИ
=================

🏗️ ЭТАП 1: БАЗОВАЯ НАСТРОЙКА И АРХИТЕКТУРА
------------------------------------------
TODO: ✅ Настроить базовую структуру проекта
  - ✅ cmd/mm/main.go (entry point)
  - config/ (конфиги для разных env)
  - internal/ (бизнес-логика, handlers, models)
  - migrations/ (SQL миграции)
  - docs/ (документация)
  - scripts/ (скрипты деплоя/утилиты)

TODO: 🐳 Docker окружение
  - docker-compose.yml (PostgreSQL + Redis + приложение)
  - Dockerfile для приложения
  - .dockerignore
  - Настройка volumes для данных

TODO: ⚙️ Конфигурация приложения
  - config/config.go - структуры конфигов
  - config/dev.yaml, config/prod.yaml
  - Парсинг env переменных
  - Валидация конфигов при старте

🗄️ ЭТАП 2: БАЗА ДАННЫХ И МИГРАЦИИ
---------------------------------
TODO: 📊 Дизайн схемы БД
  - users (id, email, password_hash, created_at, updated_at)
  - accounts (id, user_id, name, type, balance, currency, created_at)
  - categories (id, user_id, name, type, color, icon, created_at)
  - transactions (id, account_id, category_id, amount, description, date, type, created_at)
  - budgets (id, user_id, category_id, amount, period, start_date, end_date)
  - goals (id, user_id, name, target_amount, current_amount, deadline)

TODO: 🔄 Миграции
  - Использовать golang-migrate или подобное
  - migrations/000001_init.up.sql и .down.sql
  - migrations/000002_add_indexes.up.sql и .down.sql
  - Скрипт для запуска миграций в docker-compose

TODO: 💾 Database layer
  - internal/storage/postgres.go - подключение к БД
  - internal/models/ - Go structs для всех таблиц
  - internal/repository/ - интерфейсы и реализации для работы с данными
  - Использовать sqlx или GORM (на выбор)

🔐 ЭТАП 3: АВТОРИЗАЦИЯ И БЕЗОПАСНОСТЬ
-----------------------------------
TODO: 🔑 JWT Authentication
  - internal/auth/jwt.go - генерация и валидация токенов
  - internal/middleware/auth.go - middleware для проверки токенов
  - Refresh tokens в Redis
  - Password hashing (bcrypt)

TODO: 👤 User management
  - POST /api/v1/auth/register - регистрация
  - POST /api/v1/auth/login - логин
  - POST /api/v1/auth/refresh - обновление токена
  - POST /api/v1/auth/logout - выход
  - GET /api/v1/user/profile - профиль пользователя

📱 ЭТАП 4: CORE API - ФИНАНСОВЫЕ ОПЕРАЦИИ
----------------------------------------
TODO: 💳 Accounts API
  - GET /api/v1/accounts - список счетов пользователя
  - POST /api/v1/accounts - создать новый счет
  - PUT /api/v1/accounts/{id} - обновить счет
  - DELETE /api/v1/accounts/{id} - удалить счет
  - GET /api/v1/accounts/{id}/balance - баланс счета

TODO: 🏷️ Categories API
  - GET /api/v1/categories - список категорий (доходы/расходы)
  - POST /api/v1/categories - создать категорию
  - PUT /api/v1/categories/{id} - обновить категорию
  - DELETE /api/v1/categories/{id} - удалить категорию

TODO: 💸 Transactions API
  - GET /api/v1/transactions - список транзакций с фильтрами
    * ?account_id=X&category_id=Y&date_from=Z&date_to=W&type=income/expense
  - POST /api/v1/transactions - создать транзакцию
  - PUT /api/v1/transactions/{id} - обновить транзакцию
  - DELETE /api/v1/transactions/{id} - удалить транзакцию
  - GET /api/v1/transactions/{id} - детали транзакции

📊 ЭТАП 5: АНАЛИТИКА И ОТЧЕТЫ
---------------------------
TODO: 📈 Statistics API
  - GET /api/v1/stats/summary - общая статистика (баланс, доходы, расходы за период)
  - GET /api/v1/stats/by-category - статистика по категориям
  - GET /api/v1/stats/by-month - помесячная статистика
  - GET /api/v1/stats/trends - тренды доходов/расходов

TODO: 🎯 Budgets & Goals API
  - GET /api/v1/budgets - список бюджетов
  - POST /api/v1/budgets - создать бюджет
  - PUT /api/v1/budgets/{id} - обновить бюджет
  - GET /api/v1/budgets/{id}/status - статус выполнения бюджета
  - GET /api/v1/goals - список финансовых целей
  - POST /api/v1/goals - создать цель
  - PUT /api/v1/goals/{id} - обновить прогресс цели

⚡ ЭТАП 6: ОПТИМИЗАЦИЯ И КЭШИРОВАНИЕ
----------------------------------
TODO: 🔥 Redis интеграция
  - internal/cache/redis.go - подключение к Redis
  - Кэширование частых запросов (балансы, статистика)
  - Session storage для JWT refresh tokens
  - Rate limiting для API endpoints

TODO: ⚡ Performance optimization
  - Индексы в БД для частых запросов
  - Пагинация для списков транзакций
  - Batch операции для массовых операций
  - Connection pooling для БД

📚 ЭТАП 7: ДОКУМЕНТАЦИЯ И ТЕСТИРОВАНИЕ
------------------------------------
TODO: 📖 Swagger документация
  - swaggo/swag для генерации docs
  - Аннотации в handlers
  - GET /swagger/index.html - UI документации
  - swagger.yaml для экспорта

TODO: 🧪 Тестирование
  - Unit tests для business logic
  - Integration tests для API endpoints
  - Моки для внешних зависимостей (БД, Redis)
  - Test containers для тестов с БД

🚀 ЭТАП 8: ДЕПЛОЙ И CI/CD
------------------------
TODO: 🔄 GitHub Actions
  - .github/workflows/test.yml - запуск тестов
  - .github/workflows/build.yml - сборка Docker образа
  - .github/workflows/deploy.yml - деплой (опционально)
  - Linting с golangci-lint

TODO: 🎯 Production ready
  - Health check endpoints (/health, /ready)
  - Graceful shutdown
  - Логирование (logrus/zap)
  - Метрики (prometheus, опционально)
  - Environment-specific configs

💡 ДОПОЛНИТЕЛЬНЫЕ ФИЧИ (V2.0)
============================
TODO: 📤 Import/Export
  - Импорт из CSV файлов
  - Экспорт данных в различных форматах
  - Интеграция с банковскими API

TODO: 📱 Mobile API готовность
  - API versioning
  - Push notifications (Firebase)
  - Offline sync capabilities

TODO: 🤖 Автоматизация
  - Recurring transactions (подписки, зарплата)
  - Smart categorization (ML для автокатегоризации)
  - Budget alerts и notifications

🛠️ ТЕХНИЧЕСКАЯ АРХИТЕКТУРА
==========================
internal/
├── api/
│   ├── handlers/     # HTTP handlers
│   ├── middleware/   # Auth, logging, cors middleware
│   └── router.go     # Gin router setup
├── auth/            # JWT, password hashing
├── cache/           # Redis operations
├── config/          # Configuration management
├── models/          # Data structures
├── repository/      # Data access layer
├── service/         # Business logic
└── storage/         # Database connections

Принципы:
- Clean Architecture
- Dependency Injection
- Interface segregation
- Тестируемый код
- Graceful error handling

🎯 ROADMAP
=========
Неделя 1: Этапы 1-2 (Базовая настройка + БД)
Неделя 2: Этап 3 (Авторизация)
Неделя 3: Этап 4 (Core API)
Неделя 4: Этап 5-6 (Аналитика + Кэширование)
Неделя 5: Этапы 7-8 (Документация + CI/CD)

Удачи в разработке! 🚀
*/

func main() {
	cfg, err := config.MustLoadConfig("../../config/dev.yaml")
	if err != nil {
		log.Fatal(err)
	}

	log, err := logger.New(&cfg.Logger)
	if err != nil {
		return
	}

	ctx := context.Background()

	log.Info(ctx, "config load", logger.Field{
		Key:   "cfg",
		Value: cfg,
	})
	// TODO: init database connection
	// TODO: run migrations
	// TODO: init Redis cache
	// TODO: setup Gin router
	// TODO: start HTTP server with graceful shutdown

}
