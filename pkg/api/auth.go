package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 8 * time.Hour

type signInRequest struct {
	Password string `json:"password"`
}

type signInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// Пароль и JWT-токен, подписанный этим паролем.
func signInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, signInResponse{Error: "ошибка десериализации JSON"})
		return
	}

	pass := os.Getenv("TODO_PASSWORD")
	if pass == "" || req.Password != pass {
		writeJSON(w, http.StatusUnauthorized, signInResponse{Error: "неверный пароль"})
		return
	}

	claims := jwt.MapClaims{
		"hash": passwordHash(pass),
		"exp":  time.Now().Add(tokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(pass))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, signInResponse{Error: "ошибка генерации токена"})
		return
	}

	writeJSON(w, http.StatusOK, signInResponse{Token: signed})
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if pass == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "требуется авторизация")
			return
		}

		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
			return []byte(pass), nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "невалидный токен")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, "невалидный токен")
			return
		}

		hash, _ := claims["hash"].(string)
		if hash != passwordHash(pass) {
			writeError(w, http.StatusUnauthorized, "невалидный токен")
			return
		}

		next(w, r)
	}
}

func passwordHash(pass string) string {
	sum := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sum[:])
}
