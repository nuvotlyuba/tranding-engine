dc dev:
	docker compose -f docker-compose.dev.yml up -d
dc dev v :
	docker compose -f docker-compose.dev.yml down -v


# bash# Запустить только базы данных (без приложения)
# # Удобно когда запускаешь приложение локально из IDE
# docker compose -f docker-compose.dev.yml up -d postgres kafka

# # Посмотреть логи конкретного сервиса в реальном времени
# docker compose -f docker-compose.dev.yml logs -f app

# # Перезапустить только приложение без остановки БД
# docker compose -f docker-compose.dev.yml restart app

# # Подключиться к PostgreSQL из контейнера
# docker compose -f docker-compose.dev.yml exec postgres \
#   psql -U trading -d trading_engine

# # Полный сброс — удалить всё включая данные
# docker compose -f docker-compose.dev.yml down -v --remove-orphans

# # Пересобрать образ приложения после изменения Dockerfile
# docker compose -f docker-compose.dev.yml build app