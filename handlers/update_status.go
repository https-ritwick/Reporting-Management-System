package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

func UpdateStudentStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type StudentPayload struct {
			ApplicationNo string `json:"application_no"`
			CandidateName string `json:"candidate_name"`
			Branch        string `json:"branch"`
			Status        string `json:"status"`
		}

		var s StudentPayload
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		s.ApplicationNo = strings.TrimSpace(s.ApplicationNo)
		s.CandidateName = strings.TrimSpace(s.CandidateName)
		s.Branch = strings.TrimSpace(s.Branch)
		s.Status = strings.TrimSpace(s.Status)

		match, _ := regexp.MatchString(`^(131|128|127)[0-9]{9}$`, s.ApplicationNo)
		if !match || s.CandidateName == "" || s.Branch == "" || s.Status == "" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Invalid input or missing fields",
			})
			return
		}

		var exists bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM reporting_status WHERE application_no = ?)`, s.ApplicationNo).Scan(&exists)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if exists {
			_, err = db.Exec(`UPDATE reporting_status SET candidate_name = ?, branch = ?, status = ? WHERE application_no = ?`,
				s.CandidateName, s.Branch, s.Status, s.ApplicationNo)
		} else {
			_, err = db.Exec(`INSERT INTO reporting_status (application_no, candidate_name, branch, status) VALUES (?, ?, ?, ?)`,
				s.ApplicationNo, s.CandidateName, s.Branch, s.Status)
		}

		if err != nil {
			http.Error(w, "Failed to save student", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}
}
