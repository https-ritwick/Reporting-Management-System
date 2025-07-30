package handlers

import (
	"Batch/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DashboardData struct {
	Total         int
	Reported      int
	Withdrawn     int
	BranchStats   map[string]int
	GenderStats   []GenderStatsRow
	CategoryStats []CategoryStatsRow
}

type Student struct {
	ApplicationNumber string `json:"application_number"`
	FullName          string `json:"full_name"`
	FatherName        string `json:"father_name"`
	ContactNumber     string `json:"contact_number"`
	Email             string `json:"email"`
	Branch            string `json:"branch"`
	Batch             string `json:"batch"`
	Group             string `json:"group" db:"group_name"`
	Status            string `json:"status"`
}
type GenderStatsRow struct {
	Branch string
	Male   int
	Female int
	Other  int
}

type CategoryStatsRow struct {
	Branch  string
	General int
	EWS     int
	OBC     int
	SC      int
	ST      int
}

// DashboardStatsHandler - Home Tab
func DashboardStatsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var total, reported, withdrawn int
		branchStats := make(map[string]int)
		var genderStats []GenderStatsRow
		var categoryStats []CategoryStatsRow

		// Existing stats
		db.QueryRow("SELECT COUNT(*) FROM students").Scan(&total)
		db.QueryRow("SELECT COUNT(*) FROM students WHERE status = 'Reported'").Scan(&reported)
		db.QueryRow("SELECT COUNT(*) FROM students WHERE status = 'Withdrawn'").Scan(&withdrawn)

		// Branch-wise count
		rows, _ := db.Query("SELECT branch, COUNT(*) FROM students GROUP BY branch")
		defer rows.Close()
		for rows.Next() {
			var branch string
			var count int
			rows.Scan(&branch, &count)
			branchStats[branch] = count
		}

		// Gender stats: only for Reported students
		genderQuery := `
			SELECT branch,
				SUM(CASE WHEN gender = 'Male' THEN 1 ELSE 0 END) AS Male,
				SUM(CASE WHEN gender = 'Female' THEN 1 ELSE 0 END) AS Female,
				SUM(CASE WHEN gender = 'Other' THEN 1 ELSE 0 END) AS Other
			FROM students
			WHERE status = 'Reported'
			GROUP BY branch
		`
		rows, _ = db.Query(genderQuery)
		for rows.Next() {
			var row GenderStatsRow
			rows.Scan(&row.Branch, &row.Male, &row.Female, &row.Other)
			genderStats = append(genderStats, row)
		}

		// Category stats: only for Reported students
		categoryQuery := `
			SELECT branch,
				SUM(CASE WHEN category = 'General' THEN 1 ELSE 0 END) AS General,
				SUM(CASE WHEN category = 'General-EWS' THEN 1 ELSE 0 END) AS EWS,
				SUM(CASE WHEN category = 'OBC' THEN 1 ELSE 0 END) AS OBC,
				SUM(CASE WHEN category = 'SC' THEN 1 ELSE 0 END) AS SC,
				SUM(CASE WHEN category = 'ST' THEN 1 ELSE 0 END) AS ST
			FROM students
			WHERE status = 'Reported'
			GROUP BY branch
		`
		rows, _ = db.Query(categoryQuery)
		for rows.Next() {
			var row CategoryStatsRow
			rows.Scan(&row.Branch, &row.General, &row.EWS, &row.OBC, &row.SC, &row.ST)
			categoryStats = append(categoryStats, row)
		}

		tmpl := template.Must(template.ParseFiles("templates/dashboard_stats.html"))
		// Fetch distinct branches (only Reported students)
		branches := fetchDistinctOptions(db, "branch")

		type FullDashboardData struct {
			Total         int
			Reported      int
			Withdrawn     int
			BranchStats   map[string]int
			GenderStats   []GenderStatsRow
			CategoryStats []CategoryStatsRow
			BranchList    []string
		}

		tmpl = template.Must(template.ParseFiles("templates/dashboard_stats.html"))
		tmpl.Execute(w, FullDashboardData{
			Total:         total,
			Reported:      reported,
			Withdrawn:     withdrawn,
			BranchStats:   branchStats,
			GenderStats:   genderStats,
			CategoryStats: categoryStats,
			BranchList:    branches,
		})

	}
}

// StudentListHandler - Student List Tab
func StudentListHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT application_number, full_name, father_name, contact_number, email, branch, batch, group_name, status FROM students`
		filters := []string{}
		args := []interface{}{}

		if branch := r.URL.Query().Get("branch"); branch != "" {
			filters = append(filters, "branch = ?")
			args = append(args, branch)
		}
		if batch := r.URL.Query().Get("batch"); batch != "" {
			filters = append(filters, "batch = ?")
			args = append(args, batch)
		}
		if group := r.URL.Query().Get("group"); group != "" {
			filters = append(filters, "group_name = ?")
			args = append(args, group)
		}
		if status := r.URL.Query().Get("status"); status != "" {
			filters = append(filters, "status = ?")
			args = append(args, status)
		}
		if search := r.URL.Query().Get("search"); search != "" {
			filters = append(filters, "(application_number LIKE ? OR full_name LIKE ?)")
			args = append(args, "%"+search+"%", "%"+search+"%")
		}
		if len(filters) > 0 {
			query += " WHERE " + strings.Join(filters, " AND ")
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		students := []Student{}
		for rows.Next() {
			var s Student
			_ = rows.Scan(&s.ApplicationNumber, &s.FullName, &s.FatherName, &s.ContactNumber, &s.Email, &s.Branch, &s.Batch, &s.Group, &s.Status)
			students = append(students, s)
		}

		// Fetch distinct values dynamically
		branchOptions := fetchDistinctOptions(db, "branch")
		batchOptions := fetchDistinctOptions(db, "batch")
		groupOptions := fetchDistinctOptions(db, "group_name")
		statusOptions := []string{"Reported", "Withdrawn"}

		tmpl := template.Must(template.ParseFiles("templates/student_list.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Students":      students,
			"BranchOptions": branchOptions,
			"BatchOptions":  batchOptions,
			"GroupOptions":  groupOptions,
			"StatusOptions": statusOptions,
		})
	}
}

func fetchDistinctOptions(db *sql.DB, column string) []string {
	query := "SELECT DISTINCT " + column + " FROM students ORDER BY " + column
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error fetching options:", err)
		return []string{}
	}
	defer rows.Close()

	var options []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err == nil && val != "" {
			options = append(options, val)
		}
	}
	return options
}

// UpdateStudentHandler - saves edits made on student list rows
func UpdateStudentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var s Student
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			log.Println("❌ JSON decode error:", err)
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		log.Printf("🔄 Received update for AppNo %s: %+v\n", s.ApplicationNumber, s)

		// Execute the update
		query := `
			UPDATE students SET 
				full_name=?, 
				father_name=?, 
				contact_number=?, 
				email=?, 
				branch=?, 
				batch=?, 
				group_name=?, 
				status=? 
			WHERE application_number=?;
		`

		res, err := db.Exec(query,
			s.FullName,
			s.FatherName,
			s.ContactNumber,
			s.Email,
			s.Branch,
			s.Batch,
			s.Group, // maps to `group_name` in DB
			s.Status,
			s.ApplicationNumber,
		)
		if err != nil {
			log.Println("❌ DB Exec error:", err)
			http.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := res.RowsAffected()
		log.Printf("✅ Update complete. Rows affected: %d\n", rowsAffected)

		if rowsAffected == 0 {
			http.Error(w, "No student found with given application number", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
func ResendEmailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		appNo := r.FormValue("application_number")
		if appNo == "" {
			http.Error(w, "Application number is required", http.StatusBadRequest)
			return
		}

		var email, name, batch, group string
		query := `SELECT email, full_name, batch, group_name FROM students WHERE application_number = ?`
		err := db.QueryRow(query, appNo).Scan(&email, &name, &batch, &group)
		if err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		subject := " Your Batch and Group Allotment | USAR"

		html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
			<meta charset="UTF-8">
			<title>Registration Confirmation</title>
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
					<p>Welcome to University School of Automation & Robotics! Thank you for registering. Your details have been successfully recorded in the system. Please find your details below:</p>
					<table cellpadding="8" style="width: 100%%; border-collapse: collapse;">
					<tr><td><strong>Application Number</strong></td><td>%s</td></tr>
					<tr><td><strong>Batch</strong></td><td>%s</td></tr>
					<tr><td><strong>Group</strong></td><td>%s</td></tr>
					</table>

					<h4><strong> Important Instructions </strong></h4>
					<ul>
					<li><a class="underline"href="https://docs.google.com/document/d/1B3zj4LK8akjsmjB_nNKSfM9_Tmv4j_D_00z0W6nx14k/edit?usp=sharing" target="_blank">
      				Click Here to Read Important Instructions for Newly Admitted Candidates.
      				</a></li>
					<li>Please ensure all details are correct.</li>
					<li>Please Note Down your Allotted Batch & Group for Future Reference</li>
					<li>Students may fill out the Hostel Admission Form available on the University Website.</li>
					<li>If any discrepancies are found, please reply to this email with the correct information.</li>
					<li>Join the official WhatsApp Group.</li>
					</ul>

					<p style="margin-top: 30px;">Regards,<br><strong>USAR Student Cell</strong><br>GGSIPU</p>
				</td>
				</tr>
			</table>
			</body>
			</html>
		`, name, appNo, batch, group)

		err = utils.SendHTMLEmail(email, subject, html)
		if err != nil {
			http.Error(w, "❌ Failed to send email", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("✅ Email sent successfully to " + email))
	}
}

type UploadRecord struct {
	AppNumber     string
	FullName      string
	Photo         string
	JEEScorecard  string
	Profile       string
	FeeReceipt    string
	ReportingSlip string
}

// GET /dashboard/uploads
func UploadsDashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		var query strings.Builder
		query.WriteString(`
			SELECT 
				u.application_number, s.full_name, 
				u.photo_path, u.jee_scorecard_path, 
				u.candidate_profile_path, u.fee_receipt_path, u.reporting_slip_path
			FROM uploads u
			JOIN students s ON u.application_number = s.application_number
		`)

		var args []interface{}
		if search != "" {
			query.WriteString(`
				WHERE LOWER(s.full_name) LIKE ? OR u.application_number LIKE ?
			`)
			searchTerm := "%" + strings.ToLower(search) + "%"
			args = append(args, searchTerm, searchTerm)
		}

		query.WriteString(" ORDER BY u.uploaded_at DESC")

		rows, err := db.Query(query.String(), args...)
		if err != nil {
			http.Error(w, "Database error while fetching uploads", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var uploads []UploadRecord
		for rows.Next() {
			var u UploadRecord
			err := rows.Scan(&u.AppNumber, &u.FullName, &u.Photo, &u.JEEScorecard, &u.Profile, &u.FeeReceipt, &u.ReportingSlip)
			if err == nil {
				u.Photo = strings.ReplaceAll(u.Photo, "\\", "/")
				u.JEEScorecard = strings.ReplaceAll(u.JEEScorecard, "\\", "/")
				u.Profile = strings.ReplaceAll(u.Profile, "\\", "/")
				u.FeeReceipt = strings.ReplaceAll(u.FeeReceipt, "\\", "/")
				u.ReportingSlip = strings.ReplaceAll(u.ReportingSlip, "\\", "/")
				uploads = append(uploads, u)
			}
		}

		tmpl := template.Must(template.ParseFiles("templates/upload_list.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Uploads": uploads,
		})
	}
}

func ReuploadDocumentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // 10MB
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		appNo := r.FormValue("app_number")
		if appNo == "" {
			http.Error(w, "Application number missing", http.StatusBadRequest)
			return
		}

		basePath := fmt.Sprintf("static/uploads/%s", appNo)
		os.MkdirAll(basePath, os.ModePerm)

		fields := map[string]string{
			"photo":          "photo",
			"jee_scorecard":  "jee_scorecard",
			"profile":        "candidate_profile",
			"fee_receipt":    "fee_receipt",
			"reporting_slip": "reporting_slip",
		}
		updateMap := map[string]string{}

		for field, filename := range fields {
			fKey := fmt.Sprintf("%s_%s", field, appNo)
			file, handler, err := r.FormFile(fKey)
			if err != nil {
				continue // Skip if not uploaded
			}
			defer file.Close()

			ext := filepath.Ext(handler.Filename)
			savePath := fmt.Sprintf("%s/%s%s", basePath, filename, ext)

			out, err := os.Create(savePath)
			if err != nil {
				log.Println("❌ Cannot save:", err)
				continue
			}
			defer out.Close()
			io.Copy(out, file)

			updateMap[filename+"_path"] = savePath
		}

		// Update DB
		if len(updateMap) > 0 {
			query := "UPDATE uploads SET "
			args := []interface{}{}
			setParts := []string{}
			for col, path := range updateMap {
				setParts = append(setParts, fmt.Sprintf("%s = ?", col))
				args = append(args, strings.ReplaceAll(path, "\\", "/"))
			}
			query += strings.Join(setParts, ", ") + " WHERE application_number = ?"
			args = append(args, appNo)
			log.Printf("➡️  Updating uploads for %s with fields: %+v", appNo, updateMap)
			log.Printf("Generated SQL: %s", query)
			log.Printf("Args: %v", args)

			_, err := db.Exec(query, args...)
			if err != nil {
				http.Error(w, "Failed to update database", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/dashboard/uploads", http.StatusSeeOther)
	}
}
