# Koyeb Deployment Guide

## Prerequisites

- Koyeb account
- PostgreSQL database (can be Koyeb Database or external)
- Git repository with your code

## Environment Variables

Set the following environment variables in Koyeb:

### Database Configuration
- `DB_HOST` - PostgreSQL host
- `DB_HOST_UNPOOLED` - PostgreSQL host (unpooled)
- `DB_USER` - PostgreSQL username
- `DB_NAME` - PostgreSQL database name
- `DB_PASSWORD` - PostgreSQL password
- `DB_PORT` - PostgreSQL port (default: 5432)

### Admin User
- `ADMIN_EMAIL` - Default admin email
- `ADMIN_PASSWORD` - Default admin password
- `ADMIN_ROLE` - Default admin role

### Application Config
- `JWT_SECRET` - Secret key for JWT tokens
- `ENV` - Environment (production/development)
- `PORT` - Application port (default: 8080)
- `ALLOWED_ORIGINS` - CORS allowed origins

### Email Configuration
- `EMAIL_HOST` - SMTP host (e.g., smtp.gmail.com)
- `EMAIL_PORT` - SMTP port (e.g., 587)
- `EMAIL_USERNAME` - SMTP username
- `EMAIL_PASSWORD` - SMTP password/app password

## Deployment Methods

### Method 1: Using koyeb.yml (Recommended)

1. Push your code to a Git repository
2. In Koyeb dashboard, create a new service
3. Select your Git repository
4. Koyeb will automatically detect the koyeb.yml configuration
5. Set the environment variables (or use Koyeb Database for auto-population)
6. Deploy

### Method 2: Manual Configuration

1. Push your code to a Git repository
2. In Koyeb dashboard, create a new service
3. Select your Git repository
4. Choose Dockerfile build
5. Set environment variables manually
6. Expose port 8080
7. Deploy

## Building Locally (Optional)

```bash
docker build -t kasirpinter-go .
docker run -p 8080:8080 --env-file .env kasirpinter-go
```

## Health Check

The application will be available at the Koyeb service URL with:
- GraphQL Playground: `/`
- GraphQL API: `/query`
