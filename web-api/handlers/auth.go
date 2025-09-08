package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
	"web-api/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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

func GenerateToken(id string, email string, secret string, expiration time.Duration) (string, error) {
	claims := Claims{
		UserID: id,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

type AuthHandler struct {
	userRepository models.UserRepository
}

func NewAuthHander(db *sql.DB) *AuthHandler {
	return &AuthHandler{userRepository: *models.NewUserRepository(db)}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticateResponse struct {
	Token string `json:"token"`
}

const Secret = "1234"

func (auth *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var request RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return
		}

		log.Printf("Attempting to register with account: %s and password: %s", request.Email, request.Password)

		hashedPassword, err := HashPassword(request.Password)
		if err != nil {
			log.Printf("There was a problem hashing the password: %v", err)
		}
		user := models.User{
			Email:        request.Email,
			PasswordHash: hashedPassword,
		}
		err = auth.userRepository.CreateUser(&user)
		if err != nil {
			log.Printf("Error creating user. %v", err)
			return
		}

		log.Printf("User Created: %v", user)
		token, err := GenerateToken(user.ID, user.Email, Secret, 24*time.Hour)
		log.Println(token)
		response, _ := json.Marshal(user)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (auth *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
	}
	log.Printf("Attempting to login with account: %s and password: %s", request.Email, request.Password)
	// check if username is there
	user, err := auth.userRepository.FindUserByEmail(request.Email)
	if err != nil {
		log.Printf("Could not find user. %v", err)
		return
	}
	log.Printf("Found user: %v", user)
	correctPass := CheckPassword(user.PasswordHash, request.Password)
	if !correctPass {
		log.Printf("Wrong pass!")
		return
	}
	token, err := GenerateToken(user.ID, user.Email, Secret, 24*time.Hour)
	log.Println(token)
	log.Printf("You're in!")

}
