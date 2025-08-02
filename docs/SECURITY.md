# Security Documentation

This document outlines the security features, best practices, and guidelines implemented in the Fullstack Template to ensure robust application security.

## Table of Contents

- [Security Overview](#security-overview)
- [Authentication Security](#authentication-security)
- [Authorization & RBAC](#authorization--rbac)
- [Input Validation & Sanitization](#input-validation--sanitization)
- [Session Management](#session-management)
- [Data Protection](#data-protection)
- [Infrastructure Security](#infrastructure-security)
- [Security Monitoring](#security-monitoring)
- [Security Best Practices](#security-best-practices)
- [Incident Response](#incident-response)
- [Security Checklist](#security-checklist)

---

## Security Overview

The Fullstack Template implements **defense-in-depth** security principles with multiple layers of protection:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Transport Security                           │
│  • HTTPS/TLS • Security Headers • CORS • Rate Limiting         │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                 Authentication Security                         │
│  • JWT Tokens • Password Hashing • Token Refresh • MFA Ready   │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                  Authorization Security                         │
│  • RBAC • Route Protection • Component Guards • API Middleware │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                    Input Security                               │
│  • Validation • Sanitization • SQL Injection Prevention        │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                     Data Security                               │
│  • Encryption • Secure Storage • Audit Trails • Privacy        │
└─────────────────────────────────────────────────────────────────┘
```

### Security Principles

1. **Zero Trust**: Never trust, always verify
2. **Least Privilege**: Minimum required permissions
3. **Defense in Depth**: Multiple security layers
4. **Fail Securely**: Secure defaults and error handling
5. **Security by Design**: Built-in, not bolted-on

---

## Authentication Security

### JWT Implementation

The system uses **JSON Web Tokens (JWT)** with secure configuration:

```go
// JWT Configuration
type JWTConfig struct {
    Secret               string        // 256-bit secret key
    AccessTokenDuration  time.Duration // Short-lived (1 hour)
    RefreshTokenDuration time.Duration // Longer-lived (30 days)
    Issuer              string        // Token issuer
    Audience            string        // Token audience
}

// Secure JWT Claims
type JWTClaims struct {
    UserID    uint     `json:"user_id"`
    Email     string   `json:"email"`
    Role      UserRole `json:"role"`
    TokenType string   `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}
```

### Token Security Features

1. **Short-lived Access Tokens**: 1-hour expiration to limit exposure
2. **Secure Refresh Tokens**: Database-stored, revocable tokens
3. **Token Rotation**: New refresh token on each refresh
4. **Token Blacklisting**: Ability to revoke tokens immediately
5. **Audience/Issuer Validation**: Prevents token reuse across systems

### Password Security

```go
// Secure password hashing with bcrypt
func hashPassword(password string) (string, error) {
    // Use bcrypt with appropriate cost (12+ for production)
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// Password strength validation
func validatePassword(password string) error {
    if len(password) < 8 {
        return ErrPasswordTooWeak
    }
    
    // Additional checks:
    // - Minimum 8 characters
    // - At least one uppercase letter
    // - At least one lowercase letter  
    // - At least one digit
    // - At least one special character
    
    return nil
}
```

### Multi-Factor Authentication (MFA) Ready

The architecture supports MFA implementation:

```go
// MFA types
type MFAMethod string

const (
    MFAMethodTOTP = "totp"   // Time-based OTP (Google Authenticator)
    MFAMethodSMS  = "sms"    // SMS-based OTP
    MFAMethodEmail = "email" // Email-based OTP
)

// User MFA configuration
type UserMFA struct {
    UserID     uint      `json:"user_id"`
    Method     MFAMethod `json:"method"`
    Secret     string    `json:"secret"` // Encrypted TOTP secret
    BackupCodes []string `json:"backup_codes"` // Encrypted backup codes
    Verified   bool      `json:"verified"`
    CreatedAt  time.Time `json:"created_at"`
}
```

---

## Authorization & RBAC

### Role-Based Access Control

The system implements a hierarchical RBAC model:

```go
// User roles with inheritance
type UserRole string

const (
    RoleUser  UserRole = "user"  // Basic user permissions
    RoleAdmin UserRole = "admin" // Full system access
)

// Permission matrix
var PermissionMatrix = map[UserRole][]Permission{
    RoleUser: {
        PermissionReadOwnProfile,
        PermissionUpdateOwnProfile,
        PermissionReadOwnData,
        PermissionUpdateOwnData,
    },
    RoleAdmin: {
        // Inherits all user permissions plus:
        PermissionReadAllUsers,
        PermissionUpdateAllUsers,
        PermissionDeleteUsers,
        PermissionManageRoles,
        PermissionViewAuditLogs,
        PermissionSystemConfig,
    },
}
```

### Backend Authorization Middleware

```go
// Role-based route protection
func RequireRole(role UserRole) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUserFromContext(c)
        
        if user == nil {
            c.JSON(401, gin.H{"error": "Authentication required"})
            c.Abort()
            return
        }
        
        if !user.HasRole(role) {
            logger.Warn("access denied", 
                "user_id", user.ID, 
                "required_role", role, 
                "user_role", user.Role,
                "endpoint", c.Request.URL.Path,
            )
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Permission-based protection
func RequirePermission(permission Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUserFromContext(c)
        
        if !user.HasPermission(permission) {
            c.JSON(403, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Frontend Authorization Guards

```typescript
// Role-based component rendering
interface RoleGuardProps {
  requiredRole: UserRole | UserRole[];
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export const RoleGuard: React.FC<RoleGuardProps> = ({
  requiredRole,
  children,
  fallback = null,
}) => {
  const { user } = useAuth();
  
  if (!user || !user.isActive()) {
    return fallback;
  }
  
  const hasRequiredRole = Array.isArray(requiredRole)
    ? requiredRole.some(role => user.hasRole(role))
    : user.hasRole(requiredRole);
    
  return hasRequiredRole ? children : fallback;
};

// Protected route component
export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requiredRole,
  requireEmailVerification = false,
}) => {
  const { user, isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();
  
  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        navigate('/login');
        return;
      }
      
      if (requireEmailVerification && !user?.email_verified) {
        navigate('/verify-email');
        return;
      }
      
      if (requiredRole && !user?.hasRole(requiredRole)) {
        navigate('/unauthorized');
        return;
      }
    }
  }, [isAuthenticated, isLoading, user, requiredRole]);
  
  if (isLoading) {
    return <LoadingSpinner />;
  }
  
  return isAuthenticated ? children : null;
};
```

---

## Input Validation & Sanitization

### Backend Validation

```go
// Request validation with binding tags
type CreateUserRequest struct {
    Email     string `json:"email" binding:"required,email,max=255"`
    Password  string `json:"password" binding:"required,min=8,max=128"`
    FirstName string `json:"first_name" binding:"required,min=1,max=50,alpha"`
    LastName  string `json:"last_name" binding:"required,min=1,max=50,alpha"`
}

// Custom validators
func validateEmail(fl validator.FieldLevel) bool {
    email := fl.Field().String()
    // Additional email validation beyond basic format
    return emailRegex.MatchString(email) && !isDisposableEmail(email)
}

// SQL injection prevention with GORM
func (r *userRepository) GetByEmail(email string) (*User, error) {
    var user User
    // GORM automatically uses prepared statements
    err := r.db.Where("email = ?", email).First(&user).Error
    return &user, err
}

// Manual query with parameterization
func (r *userRepository) SearchUsers(query string) ([]*User, error) {
    var users []*User
    // Always use parameterized queries
    err := r.db.Raw(
        "SELECT * FROM users WHERE LOWER(first_name || ' ' || last_name) LIKE LOWER(?)",
        "%"+query+"%",
    ).Scan(&users).Error
    return users, err
}
```

### Frontend Validation

```typescript
// Form validation schema
interface ValidationSchema {
  email: (value: string) => string | null;
  password: (value: string) => string | null;
  name: (value: string) => string | null;
}

const validationSchema: ValidationSchema = {
  email: (value: string) => {
    if (!value) return 'Email is required';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      return 'Please enter a valid email address';
    }
    if (value.length > 255) return 'Email is too long';
    return null;
  },
  
  password: (value: string) => {
    if (!value) return 'Password is required';
    if (value.length < 8) return 'Password must be at least 8 characters';
    if (value.length > 128) return 'Password is too long';
    if (!/(?=.*[a-z])/.test(value)) return 'Password must contain lowercase letter';
    if (!/(?=.*[A-Z])/.test(value)) return 'Password must contain uppercase letter';
    if (!/(?=.*\d)/.test(value)) return 'Password must contain a number';
    if (!/(?=.*[!@#$%^&*])/.test(value)) return 'Password must contain special character';
    return null;
  },
  
  name: (value: string) => {
    if (!value) return 'Name is required';
    if (value.length > 50) return 'Name is too long';
    if (!/^[a-zA-Z\s-']+$/.test(value)) return 'Name contains invalid characters';
    return null;
  },
};

// XSS prevention with sanitization
import DOMPurify from 'dompurify';

const sanitizeInput = (input: string): string => {
  return DOMPurify.sanitize(input, { 
    ALLOWED_TAGS: [], 
    ALLOWED_ATTR: [] 
  });
};

// Content Security Policy (CSP) headers
const cspDirectives = {
  'default-src': ["'self'"],
  'script-src': ["'self'", "'unsafe-inline'"],
  'style-src': ["'self'", "'unsafe-inline'"],
  'img-src': ["'self'", "data:", "https:"],
  'connect-src': ["'self'"],
  'font-src': ["'self'"],
  'object-src': ["'none'"],
  'media-src': ["'self'"],
  'frame-src': ["'none'"],
};
```

---

## Session Management

### Secure Session Handling

```go
// Session configuration
type SessionConfig struct {
    MaxConcurrentSessions int           // Limit concurrent sessions
    SessionTimeout        time.Duration // Automatic timeout
    RefreshTokenRotation  bool          // Rotate refresh tokens
    DeviceTracking        bool          // Track device information
}

// Session tracking
type UserSession struct {
    ID           string    `json:"id"`
    UserID       uint      `json:"user_id"`
    DeviceInfo   string    `json:"device_info"`
    IPAddress    string    `json:"ip_address"`
    UserAgent    string    `json:"user_agent"`
    LastActivity time.Time `json:"last_activity"`
    CreatedAt    time.Time `json:"created_at"`
    ExpiresAt    time.Time `json:"expires_at"`
}

// Session middleware
func SessionTracking() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUserFromContext(c)
        if user != nil {
            // Update last activity
            sessionService.UpdateActivity(user.ID, c.ClientIP(), c.GetHeader("User-Agent"))
            
            // Check for suspicious activity
            if sessionService.IsSuspiciousActivity(user.ID, c.ClientIP()) {
                logger.Warn("suspicious activity detected",
                    "user_id", user.ID,
                    "ip", c.ClientIP(),
                    "user_agent", c.GetHeader("User-Agent"),
                )
                // Could force re-authentication or lock account
            }
        }
        c.Next()
    }
}
```

### Frontend Session Management

```typescript
// Automatic token refresh
class ApiClient {
  private refreshPromise: Promise<void> | null = null;

  private async request<T>(endpoint: string, options: RequestInit): Promise<T> {
    let response = await fetch(url, config);
    
    // Handle token refresh on 401 errors
    if (response.status === 401 && endpoint !== '/auth/refresh' && endpoint !== '/auth/login') {
      const refreshed = await this.refreshToken();
      if (refreshed) {
        // Retry original request with new token
        response = await fetch(url, config);
      } else {
        // Redirect to login
        window.location.href = '/login';
        return Promise.reject(new Error('Authentication failed'));
      }
    }
    
    return this.handleResponse(response);
  }

  async refreshToken(): Promise<boolean> {
    // Prevent multiple simultaneous refresh attempts
    if (this.refreshPromise) {
      await this.refreshPromise;
      return true;
    }

    this.refreshPromise = this.performRefresh();
    
    try {
      await this.refreshPromise;
      return true;
    } catch (error) {
      console.error('Token refresh failed:', error);
      return false;
    } finally {
      this.refreshPromise = null;
    }
  }
}

// Session timeout warning
export const useSessionTimeout = () => {
  const { logout } = useAuth();
  
  useEffect(() => {
    let timeoutId: NodeJS.Timeout;
    let warningId: NodeJS.Timeout;
    
    const resetTimer = () => {
      clearTimeout(timeoutId);
      clearTimeout(warningId);
      
      // Show warning 5 minutes before timeout
      warningId = setTimeout(() => {
        const extend = confirm('Your session will expire soon. Continue?');
        if (!extend) {
          logout();
        }
      }, 55 * 60 * 1000); // 55 minutes
      
      // Force logout after 1 hour
      timeoutId = setTimeout(() => {
        logout();
      }, 60 * 60 * 1000); // 60 minutes
    };
    
    // Reset timer on user activity
    const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart'];
    events.forEach(event => {
      document.addEventListener(event, resetTimer, true);
    });
    
    resetTimer(); // Initial timer
    
    return () => {
      clearTimeout(timeoutId);
      clearTimeout(warningId);
      events.forEach(event => {
        document.removeEventListener(event, resetTimer, true);
      });
    };
  }, [logout]);
};
```

---

## Data Protection

### Data Encryption

```go
// Encryption utilities
type Encryptor struct {
    key []byte // 32-byte AES-256 key
}

func NewEncryptor(key string) (*Encryptor, error) {
    keyBytes := sha256.Sum256([]byte(key))
    return &Encryptor{key: keyBytes[:]}, nil
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
    if plaintext == "" {
        return "", nil
    }
    
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Sensitive data encryption
type EncryptedField struct {
    Value string `gorm:"column:encrypted_value"`
}

func (ef *EncryptedField) Set(value string, encryptor *Encryptor) error {
    encrypted, err := encryptor.Encrypt(value)
    if err != nil {
        return err
    }
    ef.Value = encrypted
    return nil
}

func (ef *EncryptedField) Get(encryptor *Encryptor) (string, error) {
    return encryptor.Decrypt(ef.Value)
}

// User model with encrypted fields
type User struct {
    ID              uint           `json:"id" gorm:"primarykey"`
    Email           string         `json:"email"`
    PasswordHash    string         `json:"-"`
    SSN             EncryptedField `json:"-" gorm:"embedded;embeddedPrefix:ssn_"`
    CreditCard      EncryptedField `json:"-" gorm:"embedded;embeddedPrefix:cc_"`
}
```

### Data Privacy

```go
// GDPR compliance utilities
type PrivacyService struct {
    userRepo UserRepository
    logger   *slog.Logger
}

// Right to be forgotten (GDPR Article 17)
func (s *PrivacyService) DeleteUserData(userID uint, reason string) error {
    // Log the deletion request
    s.logger.Info("user data deletion requested",
        "user_id", userID,
        "reason", reason,
    )
    
    // Anonymize audit logs instead of deletion for compliance
    if err := s.anonymizeAuditLogs(userID); err != nil {
        return fmt.Errorf("failed to anonymize audit logs: %w", err)
    }
    
    // Delete user account and related data
    if err := s.userRepo.HardDelete(userID); err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    
    return nil
}

// Data export (GDPR Article 20)
func (s *PrivacyService) ExportUserData(userID uint) (*UserDataExport, error) {
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        return nil, err
    }
    
    export := &UserDataExport{
        PersonalData: user.ToExportFormat(),
        ActivityData: s.getUserActivity(userID),
        Preferences:  user.Preferences,
        ExportedAt:   time.Now(),
    }
    
    return export, nil
}

// Data minimization
func (s *PrivacyService) CleanupExpiredData() error {
    // Delete old password reset tokens
    if err := s.deleteExpiredTokens(); err != nil {
        return err
    }
    
    // Anonymize old audit logs (after retention period)
    if err := s.anonymizeOldAuditLogs(time.Now().AddDate(-7, 0, 0)); err != nil {
        return err
    }
    
    return nil
}
```

---

## Infrastructure Security

### Security Headers

```go
// Security middleware
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Prevent MIME type sniffing
        c.Header("X-Content-Type-Options", "nosniff")
        
        // Prevent framing (clickjacking protection)
        c.Header("X-Frame-Options", "DENY")
        
        // XSS protection
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // HSTS (HTTPS Strict Transport Security)
        if c.Request.TLS != nil {
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        
        // Content Security Policy
        c.Header("Content-Security-Policy", 
            "default-src 'self'; "+
            "script-src 'self' 'unsafe-inline'; "+
            "style-src 'self' 'unsafe-inline'; "+
            "img-src 'self' data: https:; "+
            "connect-src 'self'; "+
            "font-src 'self'; "+
            "object-src 'none'; "+
            "media-src 'self'; "+
            "frame-src 'none';")
        
        // Referrer Policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Feature Policy
        c.Header("Permissions-Policy", 
            "geolocation=(), microphone=(), camera=(), payment=()")
        
        c.Next()
    }
}
```

### Rate Limiting

```go
// Rate limiting configuration
type RateLimitConfig struct {
    GlobalRPS    int // Requests per second globally
    UserRPS      int // Requests per second per user
    AuthRPS      int // Special limit for auth endpoints
    WindowSize   time.Duration
    BurstAllowed int
}

// Advanced rate limiting
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
    // IP-based rate limiting
    ipLimiter := rate.NewLimiter(rate.Limit(config.GlobalRPS), config.BurstAllowed)
    ipLimiters := make(map[string]*rate.Limiter)
    mu := sync.RWMutex{}
    
    return func(c *gin.Context) {
        ip := c.ClientIP()
        
        // Get or create limiter for IP
        mu.RLock()
        limiter, exists := ipLimiters[ip]
        mu.RUnlock()
        
        if !exists {
            mu.Lock()
            limiter = rate.NewLimiter(rate.Limit(config.GlobalRPS), config.BurstAllowed)
            ipLimiters[ip] = limiter
            mu.Unlock()
        }
        
        if !limiter.Allow() {
            c.JSON(429, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": 60,
            })
            c.Abort()
            return
        }
        
        // User-specific rate limiting (if authenticated)
        if userID := GetUserIDFromContext(c); userID != 0 {
            userKey := fmt.Sprintf("user:%d", userID)
            userLimiter := getUserLimiter(userKey, config.UserRPS)
            
            if !userLimiter.Allow() {
                c.JSON(429, gin.H{
                    "error": "User rate limit exceeded",
                    "retry_after": 60,
                })
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}

// Auth-specific rate limiting
func AuthRateLimit() gin.HandlerFunc {
    return RateLimit(RateLimitConfig{
        GlobalRPS:    10,  // Very restrictive for auth
        UserRPS:      5,   // Per-user auth rate limit
        WindowSize:   time.Minute,
        BurstAllowed: 3,
    })
}
```

### CORS Configuration

```go
// CORS configuration
func ConfigureCORS() gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins: []string{
            "https://yourdomain.com",
            "https://app.yourdomain.com",
        },
        AllowMethods: []string{
            "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
        },
        AllowHeaders: []string{
            "Origin", "Content-Type", "Authorization",
            "X-Requested-With", "X-Client-Version",
        },
        ExposeHeaders: []string{
            "X-RateLimit-Limit", "X-RateLimit-Remaining",
        },
        AllowCredentials: true,
        MaxAge:          12 * time.Hour,
    }
    
    // Development environment
    if os.Getenv("ENVIRONMENT") == "development" {
        config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
    }
    
    return cors.New(config)
}
```

---

## Security Monitoring

### Audit Logging

```go
// Comprehensive audit logging
type AuditLogger struct {
    repo   AuditRepository
    logger *slog.Logger
}

func (a *AuditLogger) LogSecurityEvent(
    userID *uint,
    action AuditAction,
    level AuditLevel,
    resource string,
    description string,
    metadata map[string]interface{},
    c *gin.Context,
) {
    auditLog := &AuditLog{
        UserID:      userID,
        Action:      action,
        Level:       level,
        Resource:    resource,
        Description: description,
        IPAddress:   c.ClientIP(),
        UserAgent:   c.GetHeader("User-Agent"),
        Metadata:    metadata,
        CreatedAt:   time.Now(),
    }
    
    // Log to database
    if err := a.repo.Create(auditLog); err != nil {
        a.logger.Error("failed to create audit log", "error", err)
    }
    
    // Log to structured logger for monitoring
    a.logger.Info("security event",
        "action", action,
        "level", level,
        "user_id", userID,
        "ip", c.ClientIP(),
        "resource", resource,
    )
    
    // Send to security monitoring system
    if level == AuditLevelError || level == AuditLevelWarning {
        a.sendSecurityAlert(auditLog)
    }
}

// Security events to monitor
const (
    AuditActionLoginSuccess       AuditAction = "login_success"
    AuditActionLoginFailed        AuditAction = "login_failed"
    AuditActionPasswordChanged    AuditAction = "password_changed"
    AuditActionRoleChanged        AuditAction = "role_changed"
    AuditActionSuspiciousActivity AuditAction = "suspicious_activity"
    AuditActionRateLimitExceeded  AuditAction = "rate_limit_exceeded"
    AuditActionUnauthorizedAccess AuditAction = "unauthorized_access"
)
```

### Intrusion Detection

```go
// Security monitoring service
type SecurityMonitor struct {
    redis      *redis.Client
    logger     *slog.Logger
    auditLogger *AuditLogger
}

// Detect suspicious patterns
func (s *SecurityMonitor) AnalyzeUserActivity(userID uint, ip string, userAgent string) {
    key := fmt.Sprintf("user_activity:%d", userID)
    
    // Track login locations
    s.trackLoginLocation(userID, ip)
    
    // Track device changes
    s.trackDeviceChange(userID, userAgent)
    
    // Track rapid requests
    s.trackRequestPattern(userID, ip)
    
    // Detect brute force attempts
    s.detectBruteForce(ip)
}

func (s *SecurityMonitor) trackLoginLocation(userID uint, ip string) {
    // Get geolocation for IP
    location := s.getIPLocation(ip)
    
    // Check if login from new country/region
    lastLocation := s.getLastLoginLocation(userID)
    if lastLocation != "" && location.Country != lastLocation {
        s.auditLogger.LogSecurityEvent(
            &userID,
            AuditActionSuspiciousActivity,
            AuditLevelWarning,
            "auth",
            fmt.Sprintf("Login from new location: %s", location.Country),
            map[string]interface{}{
                "ip": ip,
                "country": location.Country,
                "previous_country": lastLocation,
            },
            nil,
        )
        
        // Could trigger email notification or require additional verification
        s.sendLocationChangeAlert(userID, location)
    }
    
    s.setLastLoginLocation(userID, location.Country)
}

func (s *SecurityMonitor) detectBruteForce(ip string) {
    key := fmt.Sprintf("failed_logins:%s", ip)
    
    // Increment failed login counter
    count, err := s.redis.Incr(context.Background(), key).Result()
    if err != nil {
        s.logger.Error("failed to track failed logins", "error", err)
        return
    }
    
    // Set expiration on first failed login
    if count == 1 {
        s.redis.Expire(context.Background(), key, 15*time.Minute)
    }
    
    // Alert on suspicious activity
    if count >= 5 {
        s.auditLogger.LogSecurityEvent(
            nil,
            AuditActionSuspiciousActivity,
            AuditLevelError,
            "auth",
            fmt.Sprintf("Brute force attempt detected from IP %s", ip),
            map[string]interface{}{
                "failed_attempts": count,
                "ip": ip,
            },
            nil,
        )
        
        // Could trigger IP blocking
        s.blockSuspiciousIP(ip, time.Hour)
    }
}
```

### Security Alerts

```go
// Security alert system
type SecurityAlerts struct {
    emailService EmailService
    slackWebhook string
    logger       *slog.Logger
}

type SecurityAlert struct {
    Level       string                 `json:"level"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    UserID      *uint                  `json:"user_id,omitempty"`
    IPAddress   string                 `json:"ip_address"`
    Timestamp   time.Time             `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata"`
}

func (s *SecurityAlerts) SendAlert(alert SecurityAlert) {
    // Log alert
    s.logger.Warn("security alert",
        "level", alert.Level,
        "title", alert.Title,
        "user_id", alert.UserID,
        "ip", alert.IPAddress,
    )
    
    // Send to Slack (for immediate notification)
    if s.slackWebhook != "" {
        s.sendSlackAlert(alert)
    }
    
    // Send email to security team
    s.emailService.SendSecurityAlert(alert)
    
    // Could integrate with:
    // - PagerDuty for critical alerts
    // - SIEM systems
    // - Security incident response tools
}

func (s *SecurityAlerts) sendSlackAlert(alert SecurityAlert) {
    payload := map[string]interface{}{
        "text": fmt.Sprintf("🚨 Security Alert: %s", alert.Title),
        "attachments": []map[string]interface{}{
            {
                "color":  s.getColorForLevel(alert.Level),
                "fields": []map[string]interface{}{
                    {"title": "Level", "value": alert.Level, "short": true},
                    {"title": "IP Address", "value": alert.IPAddress, "short": true},
                    {"title": "Description", "value": alert.Description, "short": false},
                    {"title": "Timestamp", "value": alert.Timestamp.Format(time.RFC3339), "short": true},
                },
            },
        },
    }
    
    // Send webhook
    s.sendWebhook(s.slackWebhook, payload)
}
```

---

## Security Best Practices

### Development Security

1. **Secure Coding Practices**:
   ```go
   // Always validate input
   func (s *UserService) UpdateProfile(userID uint, req *UpdateProfileRequest) error {
       if err := s.validateProfileRequest(req); err != nil {
           return err
       }
       // ... process request
   }
   
   // Use parameterized queries
   func (r *repository) GetUser(id uint) (*User, error) {
       var user User
       err := r.db.Where("id = ?", id).First(&user).Error // ✅ Safe
       // err := r.db.Raw("SELECT * FROM users WHERE id = " + strconv.Itoa(id)).Scan(&user).Error // ❌ Dangerous
       return &user, err
   }
   
   // Handle errors securely
   func (h *AuthHandler) Login(c *gin.Context) {
       user, err := h.authService.Login(req)
       if err != nil {
           // Don't leak internal errors
           h.logger.Error("login failed", "error", err)
           c.JSON(401, gin.H{"error": "Invalid credentials"}) // Generic message
           return
       }
   }
   ```

2. **Secret Management**:
   ```bash
   # Use environment variables for secrets
   export JWT_SECRET="$(openssl rand -base64 32)"
   export DATABASE_PASSWORD="secure-random-password"
   
   # Never commit secrets to git
   echo "*.env" >> .gitignore
   echo "secrets/" >> .gitignore
   
   # Use secret management tools in production
   # - AWS Secrets Manager
   # - HashiCorp Vault
   # - Azure Key Vault
   ```

3. **Dependency Security**:
   ```bash
   # Regularly update dependencies
   go mod tidy
   npm audit fix
   
   # Check for vulnerabilities
   go list -json -m all | nancy sleuth
   npm audit
   
   # Use dependency scanning tools
   # - Snyk
   # - OWASP Dependency Check
   # - GitHub Security Advisories
   ```

### Deployment Security

1. **Production Configuration**:
   ```bash
   # Use HTTPS in production
   SECURE_COOKIES=true
   CORS_ORIGINS=https://yourdomain.com
   
   # Disable debug modes
   ENVIRONMENT=production
   LOG_LEVEL=warn
   
   # Set secure JWT settings
   JWT_SECRET="$(openssl rand -base64 32)"
   JWT_ACCESS_DURATION=1h
   JWT_REFRESH_DURATION=720h
   ```

2. **Infrastructure Security**:
   ```yaml
   # docker-compose.prod.yml
   version: '3.8'
   services:
     app:
       image: your-app:latest
       environment:
         - ENVIRONMENT=production
       secrets:
         - jwt_secret
         - db_password
       deploy:
         replicas: 3
         update_config:
           parallelism: 1
           delay: 10s
   
   secrets:
     jwt_secret:
       external: true
     db_password:
       external: true
   ```

3. **Network Security**:
   ```nginx
   # nginx.conf
   server {
       listen 443 ssl http2;
       server_name yourdomain.com;
       
       # SSL configuration
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       ssl_protocols TLSv1.2 TLSv1.3;
       ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
       
       # Security headers
       add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
       add_header X-Frame-Options DENY always;
       add_header X-Content-Type-Options nosniff always;
       
       location / {
           proxy_pass http://app:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   ```

---

## Incident Response

### Security Incident Response Plan

1. **Incident Classification**:
   - **P0 (Critical)**: Data breach, system compromise
   - **P1 (High)**: Authentication bypass, privilege escalation
   - **P2 (Medium)**: Suspicious activity, failed attacks
   - **P3 (Low)**: Policy violations, minor issues

2. **Response Procedures**:
   ```go
   // Emergency response functions
   func (s *SecurityService) EmergencyLockdown(userID uint, reason string) error {
       // Immediately suspend user account
       if err := s.userService.SuspendUser(userID, reason); err != nil {
           return err
       }
       
       // Revoke all tokens
       if err := s.authService.RevokeAllTokens(userID); err != nil {
           return err
       }
       
       // Log incident
       s.auditLogger.LogSecurityEvent(
           &userID,
           "emergency_lockdown",
           AuditLevelError,
           "security",
           fmt.Sprintf("Emergency lockdown: %s", reason),
           nil,
           nil,
       )
       
       // Notify security team
       s.alerts.SendCriticalAlert("User account locked down", userID, reason)
       
       return nil
   }
   
   func (s *SecurityService) SystemLockdown() error {
       // Enable maintenance mode
       s.setMaintenanceMode(true)
       
       // Block new logins
       s.setLoginDisabled(true)
       
       // Revoke active sessions
       s.revokeAllActiveSessions()
       
       // Notify administrators
       s.alerts.SendSystemAlert("System lockdown activated")
       
       return nil
   }
   ```

3. **Forensics and Investigation**:
   ```sql
   -- Query suspicious activity
   SELECT * FROM audit_logs 
   WHERE level = 'error' 
   AND created_at > NOW() - INTERVAL '24 hours'
   ORDER BY created_at DESC;
   
   -- Analyze failed login attempts
   SELECT ip_address, COUNT(*) as attempts
   FROM audit_logs 
   WHERE action = 'login_failed'
   AND created_at > NOW() - INTERVAL '1 hour'
   GROUP BY ip_address
   HAVING COUNT(*) > 5;
   
   -- Track user activity timeline
   SELECT action, description, ip_address, created_at
   FROM audit_logs
   WHERE user_id = $1
   AND created_at BETWEEN $2 AND $3
   ORDER BY created_at;
   ```

---

## Security Checklist

### Development Checklist

- [ ] **Authentication**
  - [ ] Passwords hashed with bcrypt (cost ≥ 12)
  - [ ] JWT tokens with short expiration (≤ 1 hour)
  - [ ] Secure refresh token rotation
  - [ ] Account lockout after failed attempts
  - [ ] Email verification required

- [ ] **Authorization**
  - [ ] RBAC implemented correctly
  - [ ] All endpoints protected appropriately
  - [ ] Frontend route protection
  - [ ] Component-level access control

- [ ] **Input Security**
  - [ ] All inputs validated and sanitized
  - [ ] Parameterized database queries
  - [ ] XSS prevention measures
  - [ ] File upload restrictions

- [ ] **Session Security**
  - [ ] Secure session management
  - [ ] Session timeout configured
  - [ ] Concurrent session limits
  - [ ] Device tracking enabled

### Deployment Checklist

- [ ] **Infrastructure**
  - [ ] HTTPS enabled with valid certificates
  - [ ] Security headers configured
  - [ ] CORS properly configured
  - [ ] Rate limiting enabled
  - [ ] Firewall rules applied

- [ ] **Configuration**
  - [ ] Production environment variables set
  - [ ] Debug modes disabled
  - [ ] Strong JWT secrets
  - [ ] Database access restricted
  - [ ] Log levels appropriate

- [ ] **Monitoring**
  - [ ] Security logging enabled
  - [ ] Audit trails configured
  - [ ] Alert system functional
  - [ ] Intrusion detection active
  - [ ] Backup procedures tested

### Compliance Checklist

- [ ] **GDPR Compliance**
  - [ ] Data processing lawful basis documented
  - [ ] Privacy policy updated
  - [ ] User consent mechanisms
  - [ ] Data export functionality
  - [ ] Right to be forgotten implemented
  - [ ] Data retention policies defined

- [ ] **Security Standards**
  - [ ] OWASP Top 10 addressed
  - [ ] Security code review completed
  - [ ] Penetration testing performed
  - [ ] Vulnerability scanning regular
  - [ ] Incident response plan documented

---

This security documentation provides comprehensive coverage of the security measures implemented in the Fullstack Template. Regular security reviews and updates should be performed to maintain the security posture as the application evolves.