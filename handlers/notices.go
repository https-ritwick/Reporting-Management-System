package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Notice struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	CreatedAt   string `json:"created_at"`
}

func MainGetNoticesHandler(db *sql.DB) http.HandlerFunc {
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
		json.NewEncoder(w).Encode(notices)
	}
}
