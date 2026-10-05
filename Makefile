DATABASE_DSN="user=user password=password dbname=pg sslmode=disable host=localhost port=5433"

MIGRATION_NAME := $(word 2, $(MAKECMDGOALS))

.PHONY: migration_up migration_create

migration_status:
	goose postgres $(DATABASE_DSN) -dir ./migrations status

migration_up:
	goose postgres $(DATABASE_DSN) -dir ./migrations up

migration_create:
	@if [ -z "$(MIGRATION_NAME)" ]; then \
		echo "Ошибка: Укажите имя миграции! Пример: make migration_create migration_name"; \
		exit 1; \
	fi
	@echo "Создаем миграцию с именем: $(MIGRATION_NAME)..."

	goose create -dir ./migrations $(MIGRATION_NAME) sql

%:
	@: