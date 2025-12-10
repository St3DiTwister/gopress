# gopress — Blog API Service

A lightweight blog API written in Go, with PostgreSQL, JWT authentication, and clean layered architecture.

## 🚀 Features

- User registration (with bcrypt password hashing)
- User login (JWT-based, stored in HttpOnly cookies)
- JWT utilities for token generation & validation
- Clean repository pattern for database access
- Modular handlers and router
- PostgreSQL with docker-compose setup
- Ready for extension with middleware, protected routes, article CRUD, etc.

---


## 📦 Project Structure

```

cmd/
app/
main.go          # Application entrypoint

internal/
database/          # PostgreSQL connection (pgxpool)
models/            # Data models (User, Article, etc.)
repository/        # Repositories for DB queries
handler/
http/            # HTTP handlers and router

pkg/
jwt/               # JWT token manager (HS256)
password/          # Password hashing (bcrypt)

migrations/          # goose migrations
docker-compose.yml   # PostgreSQL + tools

````

---

## 🗄️ Database

PostgreSQL connection uses `pgxpool`.  
Run the DB using:

```bash
  docker-compose up -d
````

Apply migrations:

```bash
  goose up
```

---

## 🔐 Authentication Flow

### Registration:

* User sends email, username, password.
* Password is hashed via bcrypt.
* User is stored in the DB.

### Login:

* Credentials are verified.
* JWT token is generated (HS256).
* Token is sent to the client via **HttpOnly cookie**.

This allows secure authentication without exposing tokens to JavaScript.

---

## ⚙️ Environment Variables

Create a `.env` file:

```
DATABASE_URL=postgres://user:pass@localhost:5432/gopress?sslmode=disable
JWT_SECRET=<your_generated_secret>
```

Generate a secure secret:

```bash
  openssl rand -base64 64
```

---

## ▶️ Run the Application

```bash
  go run ./cmd/app
```

Server starts on:

```
http://localhost:8080
```

---
