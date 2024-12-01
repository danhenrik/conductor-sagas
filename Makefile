help:
	@echo "=========================================================================\n \
		full - Run servers locally\n \
		flight - Run flight service locally\n \
		hotel - Run hotel service locally\n \
		infra [up|down] - Run server and its dependencies (Server, Background services, DB, ES)\n \
		build-flight - Build the flight service executable in respective bin folder \
		build-hotel - Build the hotel service executable in respective bin folder \
		\n========================================================================="

s1: 
	cd flight/src && go run main.go

s2: 
	cd hotel/src && go run main.go

build-flight:
	go build -o flight/bin/server flight/main.go

build-hotel:
	go build -o hotel/bin/server hotel/main.go

infra:
	@echo "=========================================================================\n \
		Usage: make infra [up|down] \
		\n========================================================================="

up:
	docker compose up --build -d

down:
	docker compose down