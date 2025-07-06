package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"Batch/models"
)

type ViewData struct {
	Students []models.Student
	Search   string
	Branch   string
	Batch    string
	Group    string
}

func ViewStudentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		branch := r.URL.Query().Get("branch")
		batch := r.URL.Query().Get("batch")
		group := r.URL.Query().Get("group")

		query := `
				SELECT 
				full_name, application_number, father_name, dob, 
				contact_number, email, correspondence_address, permanent_address,
				branch, lateral_entry, category, sub_category, exam_rank,
				seat_quota, batch, group_name, status, has_edited
				FROM students WHERE status = 'Reported'`

		args := []interface{}{}

		if search != "" {
			query += ` AND (LOWER(full_name) LIKE ? OR application_number LIKE ?)`
			searchTerm := "%" + strings.ToLower(search) + "%"
			args = append(args, searchTerm, searchTerm)
		}
		if branch != "" {
			query += ` AND branch = ?`
			args = append(args, branch)
		}
		if batch != "" {
			query += ` AND batch = ?`
			args = append(args, batch)
		}
		if group != "" {
			query += ` AND ` + "`group_name`" + ` = ?`
			args = append(args, group)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var students []models.Student
		for rows.Next() {
			var s models.Student
			err := rows.Scan(
				&s.FullName, &s.ApplicationNumber, &s.FatherName, &s.DOB,
				&s.ContactNumber, &s.Email, &s.CorrespondenceAddr, &s.PermanentAddr,
				&s.Branch, &s.LateralEntry, &s.Category, &s.SubCategory, &s.Rank,
				&s.SeatQuota, &s.Batch, &s.Group, &s.Status, &s.HasEdited,
			)
			if err != nil {
				http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			students = append(students, s)
		}

		tmpl, err := template.ParseFiles("templates/view.html")
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := ViewData{
			Students: students,
			Search:   search,
			Branch:   branch,
			Batch:    batch,
			Group:    group,
		}
		tmpl.Execute(w, data)
	}
}
