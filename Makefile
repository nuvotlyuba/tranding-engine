.PHONY: up down restart logs build ps clean \
        app db kafka postgres shell-postgres

COMPOSE_FILE=docker-compose.dev.yml
COMPOSE=docker compose -f $(COMPOSE_FILE)

# -----------------------------------------------------------------------------
# Основные команды
# -----------------------------------------------------------------------------

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart

build:
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

clean:
	$(COMPOSE) down -v --remove-orphans

# -----------------------------------------------------------------------------
# Сервисы
# -----------------------------------------------------------------------------

app:
	$(COMPOSE) up -d app

db:
	$(COMPOSE) up -d postgres kafka

postgres:
	$(COMPOSE) up -d postgres

kafka:
	$(COMPOSE) up -d kafka zookeeper kafka-ui

# -----------------------------------------------------------------------------
# Утилиты
# -----------------------------------------------------------------------------

shell-postgres:
	$(COMPOSE) exec postgres \
		psql -U trading -d trading_engine

app-logs:
	$(COMPOSE) logs -f app

kafka-logs:
	$(COMPOSE) logs -f kafka

postgres-logs:
	$(COMPOSE) logs -f postgres
