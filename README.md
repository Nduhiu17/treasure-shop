# Treasure Shop Backend

A robust backend API for the Treasure Shop platform, built with Go, Gin, and MongoDB. This project supports multi-role user management, order workflows, role-based access control, and is ready for frontend integration.

## Features

- **User Management**: Register, login, and manage users with support for multiple roles per user (admin, writer, user, etc).
- **Role & Permission System**: Flexible role assignment and CRUD for roles and user-role relationships.
- **Order Management**: Create, assign, submit, approve, and provide feedback on orders. Writers can accept/decline assignments.
- **Order Types**: Admins can manage order types (CRUD).
- **Pagination**: All admin list endpoints support pagination.
- **Security**: JWT authentication, role-based middleware, and CORS configuration for frontend integration.
- **OpenAPI/Swagger Docs**: Full API documentation available at `/docs` and `/openapi.yaml`.
- **No sensitive data exposure**: Passwords are never returned in API responses.

## Tech Stack

- **Language**: Go (Golang)
- **Framework**: Gin
- **Database**: MongoDB
- **Authentication**: JWT
- **API Docs**: OpenAPI 3.0 (Swagger UI)

## Getting Started

### Prerequisites
- Go 1.20+
- MongoDB instance (local or remote)

### Environment Variables
Create a `.env` file in the project root with:
```
MONGODB_URI=mongodb://localhost:27017
DB_NAME=treasure_shop
PORT=8080
```

### Install Dependencies
```
go mod tidy
```

### Run the Server
```
go run cmd/api/main.go
```

The server will start on `http://localhost:8080` by default.

### API Documentation
- Swagger UI: [http://localhost:8080/docs](http://localhost:8080/docs)
- OpenAPI YAML: [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)

## Key Endpoints

### Auth
- `POST /auth/register` — Register a new user
- `POST /auth/login` — Login and receive JWT

### Users & Roles
- `GET /api/admin/users` — List users (admin only, paginated)
- `POST /api/admin/roles` — Create role
- `POST /api/admin/user-roles/assign` — Assign role to user

### Orders
- `POST /api/orders` — Create order (user)
- `GET /api/orders/me` — List my orders (user)
- `PUT /api/admin/orders/:id/assign` — Assign order to writer (admin)
- `GET /api/orders?writer_id=...` — List orders assigned to a writer (supports path and query param)
- `PUT /api/writer/orders/:id/assignment-response` — Writer accepts/declines assignment
- `POST /api/writer/orders/:id/submit` — Writer submits order
- `PUT /api/orders/:id/review/approve` — Approve order (user)
- `PUT /api/orders/:id/review/feedback` — Provide feedback (user)

### Order Types
- `POST /api/admin/order-types` — Create order type (admin)
- `GET /api/admin/order-types` — List order types (admin, paginated)

## CORS
CORS is enabled and configured for integration with a frontend (default: `http://localhost:3000`).

## Development
- Code is organized by domain: `internal/auth`, `internal/orders`, `internal/users`, `internal/writers`.
- Handlers, services, and models are separated for maintainability.
- All endpoints and models are documented in `openapi.yaml`.

## License

**Private** — All rights reserved. This codebase is not open source and may not be distributed, copied, or used without explicit permission from the owner.
