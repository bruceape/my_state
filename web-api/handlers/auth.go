package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"web-api/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthConfig struct {
	Secret string
	Issuer string
	TTL    time.Duration
}

const maxBodyBytes = 1 << 20

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(id string, email string, secret string, issuer string, expiration time.Duration) (string, error) {
	claims := Claims{
		UserID: id,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id,
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string, secret string, issuer string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithLeeway(1*time.Minute),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

type AuthHandler struct {
	userRepository models.UserRepository
	cfg            AuthConfig
}

func NewAuthHandler(db *sql.DB, secret string, issuer string, ttl time.Duration) *AuthHandler {
	return &AuthHandler{
		userRepository: *models.NewUserRepository(db),
		cfg:            AuthConfig{Secret: secret, Issuer: issuer, TTL: ttl},
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticateResponse struct {
	Token string `json:"token"`
}

func (auth *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request RegisterRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			WriteError(w, http.StatusRequestEntityTooLarge, "request too large") // 413
			return
		}
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.Trim(strings.ToLower(request.Email), " ")
	if !IsValidEmail(email) {
		WriteError(w, http.StatusBadRequest, "invalid email.")
		return
	}

	if err := ValidatePassword(request.Password); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := HashPassword(request.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user := models.User{
		Email:        email,
		PasswordHash: hashedPassword,
	}
	if err := auth.userRepository.CreateUser(r.Context(), &user); err != nil {
		switch {
		case errors.Is(err, models.ErrUserExists):
			WriteError(w, http.StatusConflict, "email already exists") // 409
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	token, err := GenerateToken(user.ID, user.Email, auth.cfg.Secret, auth.cfg.Issuer, auth.cfg.TTL)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	res := AuthenticateResponse{
		Token: token,
	}
	WriteJSON(w, http.StatusCreated, res)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (auth *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request LoginRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			WriteError(w, http.StatusRequestEntityTooLarge, "request too large") // 413
			return
		}
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.Trim(strings.ToLower(request.Email), " ")
	user, err := auth.userRepository.FindUserByEmail(r.Context(), email)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	correctPass := CheckPassword(user.PasswordHash, request.Password)
	if !correctPass {
		WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := GenerateToken(user.ID, user.Email, auth.cfg.Secret, auth.cfg.Issuer, auth.cfg.TTL)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	res := AuthenticateResponse{
		Token: token,
	}
	WriteJSON(w, http.StatusOK, res)
}

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return errors.New("password must be less than 128 characters")
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'a' <= char && char <= 'z':
			hasLower = true
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' ||
			char == '%' || char == '^' || char == '&' || char == '*':
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("password must contain uppercase, lowercase, numbers, and special characters.")
	}

	return nil
}
