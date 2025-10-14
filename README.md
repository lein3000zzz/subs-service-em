# Subscriptions Service API
This project is a REST API service for managing user subscriptions with auto-generating swagger docs, built with Go, Gin, GORM, [gin-swagger](https://github.com/swaggo/gin-swagger) and PostgreSQL.

### Requirements
```
Go 1.25 or later
PostgreSQL
Docker
```

### Installation
```
git clone https://github.com/lein3000zzz/subs-service-em
cd subs-service-em
go mod download
```
### Running without docker
1. Set up a PostgreSQL database.
2. Fill `local.env` with your own data
3. Use the provided migration file
4. `go run cmd/subs/main.go`
5. The API will be available at `localhost:8080/`.
6. Docs generation is not automatic without docker - use `swag init -g cmd/subs/main.go -o docs` or see [gin-swagger](https://github.com/swaggo/gin-swagger) for more 
### Running with Docker
1. Ensure Docker and Docker Compose are installed.
2. Edit `prod.env` based on your needs.
3. Run the services: `docker-compose -f deployments/docker-compose.yml up --build`
4. The API will be available at http://localhost:8080.
### API Documentation
- Access Swagger UI at `http://localhost:8080/swagger/index.html` after starting the service.
