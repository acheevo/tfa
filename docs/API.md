# API Documentation

Complete API reference for the Fullstack Template, including authentication, user management, and admin operations.

## Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://your-domain.com`

## Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Response Format

All API responses follow a consistent format:

### Success Response
```json
{
  "data": { /* response data */ },
  "meta": { /* optional metadata */ }
}
```

### Error Response
```json
{
  "error": "Error message",
  "details": {
    "field": "Specific field error"
  }
}
```

## Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request (validation error)
- `401` - Unauthorized (authentication required)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (duplicate resource)
- `422` - Unprocessable Entity (business logic error)
- `429` - Too Many Requests (rate limited)
- `500` - Internal Server Error

---

## Authentication Endpoints

### Register User

Create a new user account.

**POST** `/auth/register`

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Validation Rules
- `email`: Valid email format, unique
- `password`: Minimum 8 characters
- `first_name`: Required, 1-50 characters
- `last_name`: Required, 1-50 characters

#### Response
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": false,
    "role": "user",
    "status": "active",
    "preferences": {},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

#### Error Responses
- `400` - Invalid input data
- `409` - Email already exists

---

### Login User

Authenticate user with email and password.

**POST** `/auth/login`

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

#### Response
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": true,
    "role": "user",
    "status": "active",
    "preferences": {
      "theme": "light",
      "language": "en"
    },
    "last_login_at": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

#### Error Responses
- `400` - Invalid input data
- `401` - Invalid credentials
- `403` - Account inactive or suspended

---

### Refresh Token

Refresh access token using refresh token.

**POST** `/auth/refresh`

#### Request Body
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Response
```json
{
  "user": { /* user object */ },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

#### Error Responses
- `401` - Invalid or expired refresh token

---

### Check Authentication

Check if current token is valid.

**GET** `/auth/check`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Response
```json
{
  "authenticated": true,
  "user": { /* user object */ }
}
```

---

### Logout

Invalidate current refresh token.

**POST** `/auth/logout`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Response
```json
{
  "message": "Logged out successfully"
}
```

---

### Logout All Devices

Invalidate all refresh tokens for the user.

**POST** `/auth/logout-all`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Response
```json
{
  "message": "Logged out from all devices"
}
```

---

## Password Management

### Forgot Password

Request password reset email.

**POST** `/auth/forgot-password`

#### Request Body
```json
{
  "email": "user@example.com"
}
```

#### Response
```json
{
  "message": "If the email exists, a reset link has been sent"
}
```

#### Notes
- Always returns success for security (doesn't reveal if email exists)
- Rate limited to prevent abuse

---

### Reset Password

Reset password using reset token.

**POST** `/auth/reset-password`

#### Request Body
```json
{
  "token": "reset-token-from-email",
  "password": "NewSecurePassword123!",
  "confirm_password": "NewSecurePassword123!"
}
```

#### Response
```json
{
  "message": "Password reset successfully"
}
```

#### Error Responses
- `400` - Invalid token, expired token, or password validation errors
- `422` - Passwords don't match

---

### Change Password

Change password for authenticated user.

**POST** `/auth/change-password`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Request Body
```json
{
  "current_password": "CurrentPassword123!",
  "new_password": "NewSecurePassword123!",
  "confirm_password": "NewSecurePassword123!"
}
```

#### Response
```json
{
  "message": "Password changed successfully"
}
```

#### Error Responses
- `401` - Current password incorrect
- `400` - Password validation errors

---

## Email Verification

### Verify Email

Verify email address using verification token.

**POST** `/auth/verify-email`

#### Request Body
```json
{
  "token": "verification-token-from-email"
}
```

#### Response
```json
{
  "message": "Email verified successfully"
}
```

---

### Resend Email Verification

Resend email verification email.

**POST** `/auth/resend-verification`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Response
```json
{
  "message": "Verification email sent"
}
```

---

## User Management

### Get User Profile

Get current user's profile information.

**GET** `/user/profile`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Response
```json
{
  "id": 1,
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "email_verified": true,
  "role": "user",
  "status": "active",
  "preferences": {
    "theme": "dark",
    "language": "en",
    "timezone": "UTC",
    "notifications": {
      "email": true,
      "sms": false,
      "push": true
    },
    "privacy": {
      "profile_visible": true,
      "show_email": false
    }
  },
  "avatar": "https://example.com/avatar.jpg",
  "last_login_at": "2024-01-01T00:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### Update User Profile

Update user profile information.

**PUT** `/user/profile`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Request Body
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "avatar": "https://example.com/new-avatar.jpg"
}
```

#### Response
```json
{
  /* Updated user object */
}
```

#### Validation Rules
- `first_name`: 1-50 characters
- `last_name`: 1-50 characters
- `avatar`: Valid URL (optional)

---

### Get User Preferences

Get user preferences.

**GET** `/user/preferences`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Response
```json
{
  "theme": "dark",
  "language": "en",
  "timezone": "UTC",
  "notifications": {
    "email": true,
    "sms": false,
    "push": true
  },
  "privacy": {
    "profile_visible": true,
    "show_email": false
  },
  "custom": {
    "sidebar_collapsed": true,
    "items_per_page": 25
  }
}
```

---

### Update User Preferences

Update user preferences.

**PUT** `/user/preferences`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Request Body
```json
{
  "theme": "dark",
  "language": "es",
  "timezone": "America/New_York",
  "notifications": {
    "email": false,
    "sms": false,
    "push": true
  },
  "privacy": {
    "profile_visible": false,
    "show_email": true
  },
  "custom": {
    "sidebar_collapsed": false,
    "items_per_page": 50
  }
}
```

#### Response
```json
{
  /* Updated preferences object */
}
```

#### Validation Rules
- `theme`: "light", "dark", or "system"
- `language`: Valid language code (e.g., "en", "es", "fr")
- `timezone`: Valid timezone (e.g., "UTC", "America/New_York")

---

### Change Email

Request email change (requires verification).

**POST** `/user/change-email`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Request Body
```json
{
  "new_email": "newemail@example.com",
  "password": "CurrentPassword123!"
}
```

#### Response
```json
{
  "message": "Email change verification sent to new address"
}
```

---

### Get User Dashboard

Get user dashboard with statistics and recent activity.

**GET** `/user/dashboard`

#### Headers
```
Authorization: Bearer <access-token>
```

#### Response
```json
{
  "user": { /* user object */ },
  "stats": {
    "total_logins": 42,
    "last_login_at": "2024-01-01T00:00:00Z",
    "account_age_days": 30,
    "profile_complete": true
  },
  "recent_logins": [
    {
      "id": 1,
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "success": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "notifications": [
    {
      "id": 1,
      "type": "info",
      "title": "Welcome!",
      "message": "Welcome to the platform",
      "read": false,
      "priority": "medium",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

## Admin Endpoints

All admin endpoints require admin role (`role: "admin"`).

### Get Users List

Get paginated list of users with filtering options.

**GET** `/admin/users`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Query Parameters
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)
- `search`: Search in email, first_name, last_name
- `role`: Filter by role ("user", "admin")
- `status`: Filter by status ("active", "inactive", "suspended")
- `sort_by`: Sort field ("created_at", "email", "last_login_at")
- `sort_order`: Sort order ("asc", "desc")

#### Example
```
GET /admin/users?page=1&page_size=10&search=john&role=user&status=active&sort_by=created_at&sort_order=desc
```

#### Response
```json
{
  "users": [
    {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "user",
      "status": "active",
      "email_verified": true,
      "avatar": "https://example.com/avatar.jpg",
      "last_login_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 150,
    "total_pages": 15,
    "has_next": true,
    "has_prev": false
  }
}
```

---

### Get User Details

Get detailed information about a specific user.

**GET** `/admin/users/{id}`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Response
```json
{
  "id": 1,
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "email_verified": true,
  "role": "user",
  "status": "active",
  "preferences": { /* preferences object */ },
  "avatar": "https://example.com/avatar.jpg",
  "last_login_at": "2024-01-01T00:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "login_history": [
    {
      "id": 1,
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "success": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "audit_trail": [
    {
      "id": 1,
      "action": "user_created",
      "level": "info",
      "resource": "user",
      "description": "User account created",
      "ip_address": "192.168.1.1",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Update User

Update user information (admin only).

**PUT** `/admin/users/{id}`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Request Body
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "email": "newemail@example.com",
  "email_verified": true,
  "role": "user",
  "status": "active",
  "avatar": "https://example.com/avatar.jpg",
  "reason": "Administrative update"
}
```

#### Response
```json
{
  "message": "User updated successfully"
}
```

---

### Update User Role

Change user role.

**PUT** `/admin/users/{id}/role`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Request Body
```json
{
  "role": "admin",
  "reason": "Promoted to administrator"
}
```

#### Response
```json
{
  "message": "User role updated successfully"
}
```

#### Business Rules
- Cannot change own role
- Must provide reason for audit trail

---

### Update User Status

Change user account status.

**PUT** `/admin/users/{id}/status`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Request Body
```json
{
  "status": "suspended",
  "reason": "Terms of service violation"
}
```

#### Response
```json
{
  "message": "User status updated successfully"
}
```

#### Status Options
- `active`: Normal account access
- `inactive`: Account disabled, cannot login
- `suspended`: Account temporarily suspended

---

### Delete Users

Delete one or more users.

**DELETE** `/admin/users?ids=1,2,3`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Request Body
```json
{
  "reason": "Account cleanup",
  "force": false
}
```

#### Response
```json
{
  "message": "Users deleted successfully",
  "deleted_count": 3
}
```

#### Parameters
- `force`: If true, permanently delete. If false, soft delete.

---

### Bulk User Actions

Perform bulk operations on multiple users.

**POST** `/admin/users/bulk`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Request Body
```json
{
  "user_ids": [1, 2, 3, 4, 5],
  "action": "activate",
  "reason": "Bulk account activation"
}
```

#### Actions
- `activate`: Set status to active
- `deactivate`: Set status to inactive
- `suspend`: Set status to suspended
- `delete`: Delete accounts
- `role_change`: Change role (requires `role` field)

#### Response
```json
{
  "total_requested": 5,
  "successful": 4,
  "failed": 1,
  "results": [
    {
      "user_id": 1,
      "success": true
    },
    {
      "user_id": 2,
      "success": false,
      "error": "Cannot modify admin user"
    }
  ]
}
```

---

### Get Admin Statistics

Get platform-wide statistics.

**GET** `/admin/stats`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Response
```json
{
  "total_users": 1250,
  "active_users": 1100,
  "inactive_users": 100,
  "suspended_users": 50,
  "admin_users": 5,
  "new_users_today": 12,
  "new_users_this_week": 85,
  "user_growth": [
    {
      "date": "2024-01-01",
      "count": 10
    },
    {
      "date": "2024-01-02",
      "count": 15
    }
  ],
  "top_countries": [
    {
      "country": "United States",
      "count": 450
    },
    {
      "country": "Canada",
      "count": 200
    }
  ]
}
```

---

### Get Audit Logs

Get system audit logs with filtering.

**GET** `/admin/audit-logs`

#### Headers
```
Authorization: Bearer <admin-access-token>
```

#### Query Parameters
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 50)
- `user_id`: Filter by user who performed action
- `target_id`: Filter by user who was affected
- `action`: Filter by action type
- `level`: Filter by level ("info", "warning", "error")
- `resource`: Filter by resource type
- `date_from`: Start date (ISO format)
- `date_to`: End date (ISO format)
- `ip_address`: Filter by IP address

#### Response
```json
{
  "logs": [
    {
      "id": 1,
      "action": "user_login",
      "level": "info",
      "resource": "auth",
      "description": "User logged in successfully",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "metadata": {
        "session_id": "sess_123",
        "login_method": "password"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "user": {
        "id": 1,
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe"
      },
      "target": null
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total": 500,
    "total_pages": 10,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## Health & Monitoring

### Health Check

Check application health status.

**GET** `/health`

#### Response
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "environment": "production",
  "checks": {
    "database": "connected",
    "email": "configured",
    "redis": "connected"
  }
}
```

---

### Application Info

Get application information.

**GET** `/api/info`

#### Response
```json
{
  "name": "Fullstack Template",
  "version": "1.0.0",
  "environment": "production",
  "build_time": "2024-01-01T00:00:00Z",
  "go_version": "1.23.0"
}
```

---

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Global**: 1000 requests per hour per IP
- **Auth endpoints**: 10 requests per minute per IP
- **Admin endpoints**: 100 requests per minute per user

### Rate Limit Headers

All responses include rate limit headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

### Rate Limit Exceeded

When rate limit is exceeded:

**Response Code**: `429 Too Many Requests`

```json
{
  "error": "Rate limit exceeded",
  "details": {
    "limit": 1000,
    "window": "1h",
    "retry_after": 3600
  }
}
```

---

## Error Codes Reference

### Authentication Errors

| Code | Message | Description |
|------|---------|-------------|
| `AUTH_001` | Invalid credentials | Email or password incorrect |
| `AUTH_002` | Account inactive | User account is inactive |
| `AUTH_003` | Account suspended | User account is suspended |
| `AUTH_004` | Email not verified | Email verification required |
| `AUTH_005` | Token expired | JWT token has expired |
| `AUTH_006` | Invalid token | JWT token is malformed or invalid |
| `AUTH_007` | Token refresh failed | Refresh token is invalid or expired |

### User Management Errors

| Code | Message | Description |
|------|---------|-------------|
| `USER_001` | User not found | User ID does not exist |
| `USER_002` | Email already exists | Email is already registered |
| `USER_003` | Invalid user role | Role is not valid |
| `USER_004` | Cannot modify self | Admin cannot modify their own account |
| `USER_005` | Password too weak | Password doesn't meet strength requirements |

### Admin Errors

| Code | Message | Description |
|------|---------|-------------|
| `ADMIN_001` | Insufficient permissions | User is not admin |
| `ADMIN_002` | Cannot delete admin | Cannot delete admin users |
| `ADMIN_003` | Bulk action failed | Some bulk operations failed |

---

## SDKs and Examples

### JavaScript/TypeScript

```typescript
import { ApiClient } from './api-client';

const api = new ApiClient('http://localhost:8080');

// Login
const authResponse = await api.login({
  email: 'user@example.com',
  password: 'password123'
});

// Get user profile
const profile = await api.getProfile();

// Update preferences
await api.updatePreferences({
  theme: 'dark',
  notifications: { email: false }
});
```

### Go

```go
package main

import (
    "github.com/your-org/fullstack-template-sdk-go"
)

func main() {
    client := sdk.NewClient("http://localhost:8080")
    
    // Login
    auth, err := client.Login("user@example.com", "password123")
    if err != nil {
        log.Fatal(err)
    }
    
    // Set token for authenticated requests
    client.SetToken(auth.AccessToken)
    
    // Get user profile
    profile, err := client.GetProfile()
    if err != nil {
        log.Fatal(err)
    }
}
```

### cURL Examples

```bash
# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Get profile (with token)
curl -X GET http://localhost:8080/user/profile \
  -H "Authorization: Bearer your-jwt-token"

# Admin: Get users
curl -X GET "http://localhost:8080/admin/users?page=1&page_size=10" \
  -H "Authorization: Bearer admin-jwt-token"
```

---

This API documentation covers all endpoints and provides comprehensive examples for integration. For additional help, see the [Developer Guide](DEVELOPER.md) or [Architecture Documentation](ARCHITECTURE.md).