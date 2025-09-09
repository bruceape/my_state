package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
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
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if issuer != "" && claims.Issuer != issuer {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

type AuthHandler struct {
	userRepository models.UserRepository
	cfg            AuthConfig
}

func NewAuthHander(db *sql.DB, secret string, issuer string) *AuthHandler {
	return &AuthHandler{
		userRepository: *models.NewUserRepository(db),
		cfg:            AuthConfig{Secret: secret, Issuer: issuer, TTL: 15 * time.Minute},
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

	if len(request.Password) < 8 || len(request.Password) > 24 {
		WriteError(w, http.StatusBadRequest, "invalid password, must be between 8 and 24 characters.")
		return
	}

	hashedPassword, err := HashPassword(request.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user := models.User{
		Email:        strings.ToLower(request.Email),
		PasswordHash: hashedPassword,
	}
	if err := auth.userRepository.CreateUser(&user); err != nil {
		w.Header().Set("Content-Type", "application/json")
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

	user, err := auth.userRepository.FindUserByEmail(strings.ToLower(request.Email))
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "user not found")
		return
	}

	correctPass := CheckPassword(user.PasswordHash, request.Password)
	if !correctPass {
		WriteError(w, http.StatusUnauthorized, "incorrect password")
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
