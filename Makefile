.PHONY: web up start down stop restart logs build help

web:
	cd web && npm run dev

build:
	docker compose up -d --build

up:
	docker compose up -d

start:
	docker compose start

down:
	docker compose down

stop: 
	docker compose stop

restart: down up

logs:
	docker compose logs -f

help:
	@echo ""
	@echo "Available commands:"
	@echo "  web           - Run frontend dev server"
	@echo "  build         - Build and start containers"
	@echo "  up            - Start containers"
	@echo "  start         - Start existing stopped containers"
	@echo "  down          - Stop and remove containers and networks"
	@echo "  stop          - Stop containers without removing them"
	@echo "  restart       - Restart containers (down then up)"
	@echo "  logs          - Follow container logs"