package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type StatusUpdate struct {
	ApplicationNo string `json:"application_no"`
	Status        string `json:"status"`
}

func BulkUpdateStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Updates []StatusUpdate `json:"updates"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		stmt, err := tx.Prepare(`UPDATE reporting_status SET status = ? WHERE application_no = ?`)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		for _, update := range payload.Updates {
			if _, err := stmt.Exec(update.Status, update.ApplicationNo); err != nil {
				tx.Rollback()
				http.Error(w, "Failed to update status", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Transaction failed", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}
}
