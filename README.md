# Personal Finance DSS Backend

## 🚀 Quick Start

### Development Mode

```bash
# 1. Start PostgreSQL only
docker-compose up -d
````
```bash

# 2. Run backend locally
go run ./cmd/server
````
or 
```bash
just dev
```

Backend will run at: `http://localhost:8080`

### Production Mode

```bash
# Start all services (PostgreSQL + Backend)
docker-compose -f docker-compose.prod.yml up -d
```

## 📝 Development Setup

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- PostgreSQL (via Docker)
- 

### Environment Variables

Create `.env` file (copy from `env.example`):

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=personalfinancedss

SERVER_HOST=localhost
SERVER_PORT=8080

JWT_SECRET=your-secret-key
JWT_EXPIRATION=3600
```

### Start Development

```bash
# Start database
docker-compose up -d

# Run backend
go run ./cmd/server
```

## 📦 Production Setup

### Prerequisites

- Docker & Docker Compose
- `.env` file configured

### Start Production

```bash
# Start all services
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Stop services
docker-compose -f docker-compose.prod.yml down
```

## 🧪 API Testing

### Health Check

```bash
curl http://localhost:8080/health
```

### Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john.doe@example.com",
    "password": "password123",
    "full_name": "John Doe"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "password123"
  }'
```

## 📊 API Endpoints

- Swagger UI: `http://localhost:8080/swagger/index.html`
- Health Check: `http://localhost:8080/health`
- Base API: `http://localhost:8080/api/v1`

## 🔧 Database

### Access Database

```bash
# Via Docker
docker-compose exec postgres psql -U postgres -d personalfinancedss

# Or directly
psql -h localhost -p 5432 -U postgres -d personalfinancedss
```

### Database Reset

```bash
# Stop and remove
docker-compose down -v
```
```bash
# Start fresh
docker-compose up -d
```

## Documentation

- `DOCKER_README.md` - Docker setup guide
- `API_INTEGRATION_README.md` - API integration guide
- `CONFIG_README.md` - Configuration guide
- `FX_README.md` - FX dependency injection guide

## Features

- Authentication & Authorization
- Account Management
- Transaction Management
- Budget Management
- Goal Management
- Investment Tracking
- Category Management
- Summary & Analytics

## Troubleshooting

### Port Already in Use

```bash
lsof -i :8080
```
```bash
lsof -i :5432
```
```bash
kill -9 <PID>
```

### Database Connection

```bash
docker-compose logs postgres
docker-compose restart postgres
```

### Backend Won't Start

```bash
go mod tidy
go run ./cmd/server
```
