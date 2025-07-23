package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

type Notices struct {
	ID          int
	Title       string
	Description string
	Link        string
	CreatedAt   string
}

// Renders the page
func ManageNoticesPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/notice.html"))
		tmpl.Execute(w, nil)
	}
}

// API to return all notices
func GetNoticesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, title, description, link, created_at FROM notices ORDER BY created_at DESC")
		if err != nil {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var notices []Notice
		for rows.Next() {
			var n Notice
			if err := rows.Scan(&n.ID, &n.Title, &n.Description, &n.Link, &n.CreatedAt); err == nil {
				notices = append(notices, n)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notices)
	}
}

// Add Notice
func AddNoticeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/manage-notices", http.StatusSeeOther)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		description := r.FormValue("description")
		link := r.FormValue("link")

		_, err := db.Exec("INSERT INTO notices (title, description, link, created_at) VALUES (?, ?, ?, ?)",
			title, description, link, time.Now())

		if err != nil {
			http.Error(w, "Failed to insert notice", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/manage-notices", http.StatusSeeOther)
	}
}

// Delete Notice
func DeleteNoticeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/notices", http.StatusSeeOther)
			return
		}
		r.ParseForm()
		id := r.FormValue("id")

		_, err := db.Exec("DELETE FROM notices WHERE id = ?", id)
		if err != nil {
			http.Error(w, "Failed to delete notice", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/manage-notices", http.StatusSeeOther)
	}
}
