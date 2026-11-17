GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

.PHONY: all prep deps-init fmt lint test vuln sec build clean

all: clean prep deps-init fmt lint vuln sec test build

prep:
	@echo "🔍 Проверка установленных утилит..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "❌ golangci-lint не установлен." && exit 1)
	@command -v govulncheck >/dev/null 2>&1 || (echo "❌ govulncheck не установлен." && exit 1)
	@command -v gosec >/dev/null 2>&1 || (echo "❌ gosec не установлен." && exit 1)
	@echo "✅ Все утилиты на месте!"

deps-init:
	@echo "📦 Очистка и проверка зависимостей Go..."
	$(GOMOD) tidy
	$(GOMOD) verify

fmt:
	@echo "🎨 Форматирование кода стандартами Go..."
	$(GOFMT) ./...
	@command -v goimports >/dev/null 2>&1 && (echo "🧹 Сортировка импортов через goimports..." && goimports -w ./...) || true

lint:
	@echo "⚙️ Запуск golangci-lint..."
	golangci-lint run ./...

test:
	@echo "🧪 Запуск юнит-тестов..."
	$(GOTEST) -v -race -cover ./...

vuln:
	@echo "🛡️ Анализ зависимостей через govulncheck..."
	govulncheck ./...

sec:
	@echo "🚨 Анализ безопасности кода через gosec..."
	@gosec -fmt json -out results.json -no-fail ./... || true
	@gosec -fmt text ./...

build: 
	@echo "🏗️ Сборка бинарного файла..."
	mkdir -p bin
	$(GOBUILD) ./...
	@echo "🎉 Сборка успешно завершена!"

clean:
	@echo "🧹 Очистка проекта..."
	$(GOCLEAN)
	rm -f results.json
	@echo "✨ Проект очищен!"

