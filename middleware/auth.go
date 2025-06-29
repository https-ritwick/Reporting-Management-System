package middleware

import (
	"crypto/sha512"
	"encoding/hex"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// HashPassword returns the SHA-512 hash of the input password
func HashPassword(password string) string {
	hash := sha512.Sum512([]byte(password))
	return hex.EncodeToString(hash[:])
}

var jwtSecret = []byte("YourSuperSecretKey") // same secret as login.go

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tokenStr := cookie.Value

		// Validate the JWT
		_, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Make sure the signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNoCookie
			}
			return jwtSecret, nil
		})

		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	}
}
