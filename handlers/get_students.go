package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type StudentRecord struct {
	ApplicationNo string `json:"application_no"`
	CandidateName string `json:"candidate_name"`
	Branch        string `json:"branch"`
	Status        string `json:"status"`
}

func GetStudentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT application_no, candidate_name, branch, status FROM reporting_status ORDER BY created_at DESC`)
		if err != nil {
			http.Error(w, "Failed to fetch records", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var students []StudentRecord
		for rows.Next() {
			var s StudentRecord
			if err := rows.Scan(&s.ApplicationNo, &s.CandidateName, &s.Branch, &s.Status); err == nil {
				students = append(students, s)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(students)
	}
}
