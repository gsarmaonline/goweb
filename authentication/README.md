# Authentication Module

This module provides a complete JWT-based authentication system for Go web applications using Gin framework. It includes user management, session handling, and middleware for protecting routes.

## Features

- User registration and login
- JWT-based session management
- Secure password hashing using bcrypt
- Session tracking with IP and user agent
- Middleware for protecting routes
- Automatic token expiration handling

## Usage

### 1. Setup

First, initialize the `SessionManager` with your application's dependencies:

```go
import "github.com/gsarmaonline/goweb/authentication"

func SetupAuth(db *gorm.DB, router *gin.Engine) (*authentication.SessionManager, error) {
    // Initialize session manager with database and router
    sessMgr, err := authentication.NewSessionManager(
        context.Background(),
        db,
        router,
    )
    if err != nil {
        return nil, err
    }

    return sessMgr, nil
}
```

### 2. Register Routes

Add the authentication routes to your application:

```go
func SetupRoutes(router *gin.Engine, sessMgr *authentication.SessionManager) {
    // Public routes
    router.POST("/register", sessMgr.Register)
    router.POST("/login", sessMgr.Login)

    // Protected routes group
    protected := router.Group("/")
    protected.Use(sessMgr.AuthMiddleware)
    {
        protected.POST("/logout", sessMgr.Logout)
        // Add other protected routes here
    }
}
```

### 3. API Endpoints

#### Register a New User

```http
POST /register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

Response (201 Created):
```json
{
    "user": {
        "id": 1,
        "email": "user@example.com",
        "created_at": "2024-04-03T20:30:00Z",
        "updated_at": "2024-04-03T20:30:00Z"
    }
}
```

#### Login

```http
POST /login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

Response (200 OK):
```json
{
    "user": {
        "id": 1,
        "email": "user@example.com",
        "created_at": "2024-04-03T20:30:00Z",
        "updated_at": "2024-04-03T20:30:00Z"
    },
    "session": {
        "id": 1,
        "user_id": 1,
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_at": "2024-04-04T20:30:00Z",
        "last_used_at": "2024-04-03T20:30:00Z",
        "last_used_ip": "127.0.0.1",
        "last_used_loc": "Mozilla/5.0..."
    }
}
```

#### Logout

```http
POST /logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

Response (200 OK):
```json
{
    "message": "Successfully logged out"
}
```

### 4. Protecting Routes

Use the `AuthMiddleware` to protect your routes:

```go
func SetupProtectedRoutes(router *gin.Engine, sessMgr *authentication.SessionManager) {
    protected := router.Group("/api")
    protected.Use(sessMgr.AuthMiddleware)
    {
        protected.GET("/profile", getProfile)
        protected.PUT("/profile", updateProfile)
    }
}
```

### 5. Getting User ID in Protected Routes

In protected routes, you can get the authenticated user's ID using `GetUserID`:

```go
func getProfile(c *gin.Context) {
    sessMgr := c.MustGet("session_manager").(*authentication.SessionManager)
    userID := sessMgr.GetUserID(c)
    
    // Use userID to fetch user data
    // ...
}
```

## Security Features

1. **Password Security**:
   - Passwords are hashed using bcrypt before storage
   - Password validation requires minimum 6 characters
   - Original passwords are never stored or returned in responses

2. **Session Security**:
   - JWT tokens with configurable expiration
   - Session tracking with IP and user agent
   - Automatic session cleanup on logout
   - Token validation on every request

3. **Input Validation**:
   - Email format validation
   - Password length requirements
   - Duplicate email prevention
   - Request body validation

## Error Handling

The module provides consistent error responses:

- 400 Bad Request: Invalid input data
- 401 Unauthorized: Invalid credentials or missing token
- 409 Conflict: Email already registered
- 500 Internal Server Error: Database or server errors

## Database Schema

The module uses two main models:

1. `SessionUser`:
   - ID (uint)
   - Email (string, unique)
   - Password (string, hashed)
   - CreatedAt (time.Time)
   - UpdatedAt (time.Time)
   - DeletedAt (time.Time, nullable)

2. `Session`:
   - ID (uint)
   - UserID (uint, foreign key)
   - ExpiresAt (time.Time)
   - LastUsedAt (time.Time)
   - LastUsedIP (string)
   - LastUsedLoc (string)
   - CreatedAt (time.Time)
   - UpdatedAt (time.Time)
   - DeletedAt (time.Time, nullable)

## Environment Variables

Required environment variables:

- `JWT_SECRET_KEY`: Secret key for JWT token signing (required) 