help:
	@echo "=========================================================================\n \
		dev - Run server locally\n \
		full-dev - Run server locally and the rest of infra\n \
		infra [up|down] - Run server and its dependencies (Server, Background services, DB, ES)\n \
		build - Build the server executable in bin folder \
		\n========================================================================="

full-dev:
	@echo "=========================================================================\n \
		Initializing infra.. \
		\n========================================================================="
	make infra up
	@echo "=========================================================================\n \
		Initializing server... \
		\n========================================================================="
	make run

dev: 
	go run main.go

build:
	go build -o bin/server main.go

infra:
	@echo "=========================================================================\n \
		Usage: make infra [up|down] \
		\n========================================================================="

up:
	docker compose up setup server database adminer -d

down:
	docker compose down setup server database adminer


	
