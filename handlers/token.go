package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
)

type TokenDetails struct {
	TokenID       string
	ApplicationNo string
	CandidateName string
	ProgramName   string
	InstituteName string
	Status        string
}

func TokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID := r.URL.Query().Get("token_id")
		if tokenID == "" {
			http.Error(w, "Missing token ID", http.StatusBadRequest)
			return
		}

		var details TokenDetails
		query := `SELECT token, application_no, candidate_name, program_name, institute_name, status 
		          FROM student_tokens WHERE token = ? LIMIT 1`
		err := db.QueryRow(query, tokenID).Scan(
			&details.TokenID, &details.ApplicationNo, &details.CandidateName,
			&details.ProgramName, &details.InstituteName, &details.Status,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Token not found", http.StatusNotFound)
			} else {
				http.Error(w, "Server error", http.StatusInternalServerError)
			}
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/token.html"))
		err = tmpl.Execute(w, details)
		if err != nil {
			http.Error(w, "Template rendering error", http.StatusInternalServerError)
		}
	}
}
