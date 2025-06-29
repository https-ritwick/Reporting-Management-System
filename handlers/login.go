package handlers

import (
	"Batch/middleware"
	"Batch/models"
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("YourSuperSecretKey") // Move to .env later

func generateJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl := template.Must(template.ParseFiles("templates/login.html"))
			tmpl.Execute(w, nil)
			return
		}

		_ = r.ParseForm()
		email := r.FormValue("email")
		password := r.FormValue("password")
		hashed := middleware.HashPassword(password)

		user, err := models.GetUserByEmail(db, email)
		if err != nil || user.PasswordHash != hashed {
			tmpl := template.Must(template.ParseFiles("templates/login.html"))
			tmpl.Execute(w, map[string]string{"Error": "Invalid email or password"})
			return
		}

		// Generate JWT token
		token, err := generateJWT(user.Email)
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			return
		}

		// Set JWT token as cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			// Secure: true, // use in production
		})

		if user.Role == "attendance" {
			http.Redirect(w, r, "/scan", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}
	}
}
func LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // Expire immediately
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
