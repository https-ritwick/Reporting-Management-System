package handlers

import (
	"Batch/models"
	"Batch/utils"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
				RenderErrorPage(w, "Failed to update record", err)
				fmt.Println("Update error:", err)
				return
			}
			uploadPath := fmt.Sprintf("static/uploads/%s/edit", app)
			os.MkdirAll(uploadPath, os.ModePerm)

			fileFields := map[string]string{
				"photo":             "photo_path",
				"jee_scorecard":     "jee_scorecard_path",
				"candidate_profile": "candidate_profile_path",
				"fee_receipt":       "fee_receipt_path",
			}

			updates := []string{}
			args := []interface{}{}

			for field, col := range fileFields {
				file, header, err := r.FormFile(field)
				if err != nil {
					continue // skip if not uploaded
				}
				defer file.Close()

				ext := filepath.Ext(header.Filename)
				filePath := fmt.Sprintf("%s/%s%s", uploadPath, field, ext)

				dst, err := os.Create(filePath)
				if err != nil {
					log.Println("❌ Error creating file:", err)
					continue
				}
				defer dst.Close()

				if _, err := io.Copy(dst, file); err != nil {
					log.Println("❌ Error saving file:", err)
					continue
				}

				// Append to update query
				updates = append(updates, fmt.Sprintf("%s = ?", col))
				args = append(args, strings.ReplaceAll(filePath, "\\", "/"))
			}

			if len(updates) > 0 {
				args = append(args, app) // For WHERE clause
				updateQuery := "UPDATE uploads SET " + strings.Join(updates, ", ") + " WHERE application_number = ?"
				_, err = db.Exec(updateQuery, args...)
				if err != nil {
					log.Println("❌ Failed to update uploads table:", err)
				}
			}

			// Fetch batch/group/branch for email after update
			row := db.QueryRow("SELECT full_name, branch, batch, group_name FROM students WHERE application_number = ?", app)
			var name, branch, batch, group string
			err = row.Scan(&name, &branch, &batch, &group)
			if err != nil {
				log.Println("❌ Failed to fetch updated details:", err)
				http.Redirect(w, r, "/confirmation?app="+app, http.StatusSeeOther)
				return
			}

			html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
			<meta charset="UTF-8">
			<title>Student Record Updation</title>
			</head>
			<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<table width="100%%" style="max-width: 600px; margin: auto; background: #fff; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1);">
				<tr style="background-color: #003366; color: white;">
				<td style="padding: 20px;">
					<img src="https://upload.wikimedia.org/wikipedia/en/thumb/b/b8/GGSIU_logo.svg/1200px-GGSIU_logo.svg.png" alt="IPU Logo" width="60" style="float: left;">
					<h2 style="text-align: center; margin: 0;">University School of Automation & Robotics</h2>
					<p style="text-align: center; margin: 0;">Guru Gobind Singh Indraprastha University, East Delhi Campus</p>
				</td>
				</tr>
				<tr>
				<td style="padding: 20px;">
					<h3>Dear %s,</h3>
					<p>This is to inform you that your student details have been successfully updated in the system. Please find your updated academic details below:</p>
					<table cellpadding="8" style="width: 100%%; border-collapse: collapse;">
					<tr><td><strong>Application Number</strong></td><td>%s</td></tr>
					<tr><td><strong>Branch</strong></td><td>%s</td></tr>
					<tr><td><strong>Batch</strong></td><td>%s</td></tr>
					<tr><td><strong>Group</strong></td><td>%s</td></tr>
					</table>

					<h4><strong> Note:</strong></h4>
					<ul>
					<li>You are allowed to edit your profile only once.</li>
					<li>If any discrepancies are found, please reply to this email with the correct information.</li>
					</ul>

					<p style="margin-top: 30px;">Regards,<br><strong>USAR Student Cell</strong><br>GGSIPU</p>
				</td>
				</tr>
			</table>
			</body>
			</html>
			`, name, app, branch, batch, group)

			err = utils.SendHTMLEmail(email, "Student Profile Updated - USAR", html)
			if err != nil {
				log.Println("❌ Failed to send update email:", err)
			} else {
				log.Println("✅ Update email sent to:", email)
			}

			http.Redirect(w, r, "/confirmation?app="+app, http.StatusSeeOther)

		}
	}
}
