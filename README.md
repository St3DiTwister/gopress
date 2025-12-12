# gopress — Blog API Service

A lightweight blog API written in Go, with PostgreSQL, JWT authentication, and clean layered architecture.

## 🚀 Features

- User registration (with bcrypt password hashing)
- User login (JWT-based, stored in HttpOnly cookies)
- JWT utilities for token generation & validation
- Clean repository pattern for database access
- Modular handlers and router
- PostgreSQL with docker-compose setup
- HTTP API and gRPC API
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
  grpc/              # gRPC servers, services, interceptors

pkg/
  jwt/               # JWT token manager (HS256)
  password/          # Password hashing (bcrypt)

migrations/          # goose migrations
docker-compose.yml   # PostgreSQL + tools
```

---

## 🗄️ Database

PostgreSQL connection uses `pgxpool`.
Run the DB using:

```bash
docker-compose up -d
```

Apply migrations:

```bash
goose up
```

---

## 🔐 Authentication Flow

### Registration

* User sends email, username, password.
* Password is hashed via bcrypt.
* User is stored in the database.

### Login

* Credentials are verified.
* JWT token is generated (HS256).
* Token is sent to the client via **HttpOnly cookie** (HTTP API).
* Token is returned in response body (gRPC API).

---

## ⚙️ Environment Variables

Create a `.env` file:

```
DATABASE_URL=postgres://user:pass@localhost:5432/gopress?sslmode=disable
JWT_SECRET=<your_generated_secret>
GRPC_PORT=50051
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

Servers start on:

```
HTTP: http://localhost:8080
gRPC: 127.0.0.1:50051
```

---

## 📡 HTTP API Endpoints

### Authentication

#### POST `/register`

Register a new user.

Request body (JSON):

```
{
  "email": "user@mail.com",
  "username": "user",
  "password": "123456"
}
```

Response (200):

```
{
  "id": "uuid",
  "username": "user",
  "email": "user@mail.com"
}
```

---

#### POST `/login`

Login user and set authentication cookie.

Request body (JSON):

```
{
  "username": "user",
  "password": "123456"
}
```

Response (200):

```
{
  "status": "ok"
}
```

JWT token is stored in an **HttpOnly cookie**.

---

#### GET `/me` 🔒

Get current authenticated user.

Response (200):

```
{
  "id": "uuid",
  "username": "user",
  "email": "user@mail.com"
}
```

---

### Articles (all routes require authentication)

#### GET `/articles` 🔒

Get list of articles.

Query parameters:

* `limit` (optional)
* `offset` (optional)

Response (200):

```
[
  {
    "id": 1,
    "title": "Title",
    "content": "Content",
    "author_id": "uuid",
    "author_username": "user",
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-01T12:00:00Z"
  }
]
```

---

#### POST `/articles` 🔒

Create new article.

Request body (JSON):

```
{
  "title": "My title",
  "content": "My content"
}
```

Response (200):

```
{
  "status": "ok",
  "id": 123
}
```

---

#### GET `/articles/{id}` 🔒

Get article by ID.

Response (200):

```
{
  "id": 123,
  "title": "My title",
  "content": "My content",
  "author_id": "uuid",
  "author_username": "user"
}
```

---

#### PUT `/articles/{id}` 🔒

Update article (only owner).

Request body (JSON):

```
{
  "title": "New title",
  "content": "New content"
}
```

Response (200):

```
{
  "status": "ok"
}
```

---

#### DELETE `/articles/{id}` 🔒

Delete article (only owner).

Response (200):

```
{
  "status": "ok"
}
```

---

## 🔌 gRPC API

The project also exposes a gRPC API intended for internal services, desktop clients, or other non-browser clients.

Authentication in gRPC is done via **JWT in metadata**:

```
authorization: Bearer <token>
```

---

### AuthService (public)

Service: `auth.AuthService`

#### Register

```
rpc Register(RegisterRequest) returns (RegisterResponse)
```

Registers a new user.

---

#### Login

```
rpc Login(LoginRequest) returns (LoginResponse)
```

Authenticates user and returns JWT token.

---

### ArticleService

Service: `article.ArticleService`

#### Public methods

* `List`
* `Get`

#### Protected methods (require JWT metadata)

* `Create`
* `Update`
* `Delete`

---

### gRPC Authentication

For protected gRPC methods, the client must send metadata:

```
authorization: Bearer <jwt_token>
```

If metadata is missing or token is invalid, the server returns `Unauthenticated`.

---

## ❗ Error Responses

### HTTP

* `400 Bad Request` — invalid input data
* `401 Unauthorized` — not authenticated
* `404 Not Found` — resource not found
* `500 Internal Server Error` — server-side error

### gRPC

* `InvalidArgument`
* `Unauthenticated`
* `NotFound`
* `Internal`
