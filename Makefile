GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

FONT_DIR=internal/terrain/fonts
FONT_FILE=$(FONT_DIR)/DejaVuSans.ttf
SYSTEM_FONT=/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf

.PHONY: all prep deps-init fmt lint test vuln sec build clean prepare-font

all: clean prep prepare-font deps-init fmt lint vuln sec test build

prep:
	@echo "🔍 Проверка установленных утилит..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "❌ golangci-lint не установлен." && exit 1)
	@command -v govulncheck >/dev/null 2>&1 || (echo "❌ govulncheck не установлен." && exit 1)
	@command -v gosec >/dev/null 2>&1 || (echo "❌ gosec не установлен." && exit 1)
	@echo "✅ Все утилиты на месте!"

prepare-font:
	@echo "🔤 Подготовка шрифта для легенды..."
	@if [ ! -f "$(SYSTEM_FONT)" ]; then \
		echo "❌ Системный шрифт не найден: $(SYSTEM_FONT)"; \
		echo "Установите пакет fonts-dejavu: sudo apt install fonts-dejavu"; \
		exit 1; \
	fi
	@mkdir -p $(FONT_DIR)
	@if [ ! -f "$(FONT_FILE)" ]; then \
		echo "📥 Копирование шрифта из системы..."; \
		cp $(SYSTEM_FONT) $(FONT_FILE); \
		echo "✅ Шрифт скопирован в $(FONT_FILE)"; \
	else \
		echo "✅ Шрифт уже на месте: $(FONT_FILE)"; \
	fi

deps-init:
	@echo "📦 Очистка и проверка зависимостей Go..."
	@$(GOCMD) get golang.org/x/image/font >/dev/null 2>&1 || true
	@$(GOCMD) get golang.org/x/image/font/opentype >/dev/null 2>&1 || true
	@$(GOCMD) get golang.org/x/image/math/fixed >/dev/null 2>&1 || true
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

build: prepare-font
	@echo "🏗️ Сборка бинарного файла..."
	mkdir -p bin
	$(GOBUILD) -o bin/map_generator ./cmd/map_generator
	@echo "🎉 Сборка успешно завершена!"

clean:
	@echo "🧹 Очистка проекта..."
	$(GOCLEAN)
	rm -f results.json
	@echo "✨ Проект очищен!"