package handlers

import (
	"Batch/models"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type EditFormData struct {
	Step    string
	Error   string
	Student *models.Student
}

func EditHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/edit.html"))

		// Step 1: Show verify form
		if r.Method == http.MethodGet || r.URL.Query().Get("step") == "" {
			tmpl.Execute(w, EditFormData{Step: "verify"})
			return
		}

		// Step 2: Verification
		if r.URL.Query().Get("step") == "verify" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Form parsing failed", http.StatusBadRequest)
				return
			}

			app := r.FormValue("application_number")
			dob := r.FormValue("dob")
			branch := r.FormValue("branch")

			parsedDOB, err := time.Parse("2006-01-02", dob)
			if err != nil {
				tmpl.Execute(w, EditFormData{
					Step:  "verify",
					Error: "Invalid date format. Use yyyy-mm-dd.",
				})
				return
			}
			dobFormatted := parsedDOB.Format("2006-01-02")

			fmt.Println("App:", app)
			fmt.Println("DOB Parsed:", dobFormatted)
			fmt.Println("Branch:", branch)

			query := `
				SELECT application_number, full_name, father_name, dob, contact_number, email, 
				correspondence_address, permanent_address, branch, batch, group_name, category, 
				sub_category, exam_rank, seat_quota, has_edited
				FROM students
				WHERE application_number = ? AND dob = ? AND branch = ?
			`

			row := db.QueryRow(query, app, dobFormatted, branch)

			var s models.Student
			err = row.Scan(&s.ApplicationNumber, &s.FullName, &s.FatherName, &s.DOB, &s.ContactNumber, &s.Email,
				&s.CorrespondenceAddr, &s.PermanentAddr, &s.Branch, &s.Batch, &s.Group,
				&s.Category, &s.SubCategory, &s.Rank, &s.SeatQuota, &s.HasEdited)

			if err != nil {
				if err == sql.ErrNoRows {
					fmt.Println("No matching student found.")
				} else {
					fmt.Println("Database error:", err)
				}
				tmpl.Execute(w, EditFormData{
					Step:  "verify",
					Error: "No matching student found with the given details.",
				})
				return
			}

			if s.HasEdited == 1 {
				tmpl.Execute(w, EditFormData{
					Step:  "verify",
					Error: "You have already edited your details once.",
				})
				return
			}

			tmpl.Execute(w, EditFormData{
				Step:    "edit",
				Student: &s,
			})
			return
		}

		// Step 3: Submit updates
		if r.URL.Query().Get("step") == "submit" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Form parsing failed", http.StatusBadRequest)
				return
			}

			app := r.FormValue("application_number")
			fullName := r.FormValue("full_name")
			fatherName := r.FormValue("father_name")
			contact := r.FormValue("contact_number")
			email := r.FormValue("email")
			corrAddr := r.FormValue("correspondence_address")
			permAddr := r.FormValue("permanent_address")
			category := r.FormValue("category")
			subCat := r.FormValue("sub_category")
			rank := r.FormValue("rank")
			quota := r.FormValue("seat_quota")

			_, err := db.Exec(`
				UPDATE students 
				SET full_name=?, father_name=?, contact_number=?, email=?, 
				correspondence_address=?, permanent_address=?, category=?, 
				sub_category=?, exam_rank=?, seat_quota=?, has_edited=1 
				WHERE application_number=?
			`, fullName, fatherName, contact, email, corrAddr, permAddr, category, subCat, rank, quota, app)

			if err != nil {
				http.Error(w, "Failed to update record", http.StatusInternalServerError)
				fmt.Println("Update error:", err)
				return
			}

			http.Redirect(w, r, "/confirmation?app="+app, http.StatusSeeOther)
		}
	}
}
