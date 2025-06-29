package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
)

func ConfirmationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appNum := r.URL.Query().Get("app")
		if appNum == "" {
			http.Error(w, "Application number missing", http.StatusBadRequest)
			return
		}

		// Fetch student from DB
		query := `
			SELECT full_name, application_number, branch, batch, group_name
			FROM students
			WHERE application_number = ?
		`

		var name, appNo, branch, batch, group string
		err := db.QueryRow(query, appNum).Scan(&name, &appNo, &branch, &batch, &group)
		if err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		data := struct {
			FullName          string
			ApplicationNumber string
			Branch            string
			Batch             string
			Group             string
		}{
			FullName: name, ApplicationNumber: appNo, Branch: branch, Batch: batch, Group: group,
		}

		tmpl := template.Must(template.ParseFiles("templates/confirmation.html"))
		tmpl.Execute(w, data)
	}
}
