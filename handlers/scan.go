package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

type StudentInput struct {
	ApplicationNo string `json:"application_no"`
	CandidateName string `json:"candidate_name"`
	Branch        string `json:"branch"`
}

func AddStudentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s StudentInput
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		s.ApplicationNo = strings.TrimSpace(s.ApplicationNo)
		s.CandidateName = strings.TrimSpace(s.CandidateName)
		s.Branch = strings.TrimSpace(s.Branch)

		match, _ := regexp.MatchString(`^(131|128|127)\d{9}$`, s.ApplicationNo)
		if !match || s.CandidateName == "" || s.Branch == "" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Invalid input format or missing data",
			})
			return
		}

		// Check for duplicate application_no
		var exists bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM reporting_status WHERE application_no = ?)`, s.ApplicationNo).Scan(&exists)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		if exists {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Student already exists",
			})
			return
		}

		// Insert new student
		_, err = db.Exec(`INSERT INTO reporting_status (application_no, candidate_name, branch, status) VALUES (?, ?, ?, 'Present')`,
			s.ApplicationNo, s.CandidateName, s.Branch)
		if err != nil {
			http.Error(w, "Failed to insert student", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}
