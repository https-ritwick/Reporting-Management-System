package handlers

import (
	"Batch/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type DashboardData struct {
	Total       int
	Reported    int
	Withdrawn   int
	BranchStats map[string]int
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

// DashboardStatsHandler - Home Tab
func DashboardStatsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var total, reported, withdrawn int
		branchStats := make(map[string]int)

		db.QueryRow("SELECT COUNT(*) FROM students").Scan(&total)
		db.QueryRow("SELECT COUNT(*) FROM students WHERE status = 'Reported'").Scan(&reported)
		db.QueryRow("SELECT COUNT(*) FROM students WHERE status = 'Withdrawn'").Scan(&withdrawn)

		rows, _ := db.Query("SELECT branch, COUNT(*) FROM students GROUP BY branch")
		defer rows.Close()
		for rows.Next() {
			var branch string
			var count int
			rows.Scan(&branch, &count)
			branchStats[branch] = count
		}

		tmpl := template.Must(template.ParseFiles("templates/dashboard_stats.html"))
		tmpl.Execute(w, DashboardData{
			Total:       total,
			Reported:    reported,
			Withdrawn:   withdrawn,
			BranchStats: branchStats,
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

		htmlBody := fmt.Sprintf(`
			<html>
			<body style="font-family:Arial,sans-serif;">
				<p>Dear <strong>%s</strong>,</p>
				<p>We are pleased to inform you that your <strong>Batch and Group</strong> have been successfully allotted as follows:</p>
				<ul>
					<li><strong>Batch:</strong> %s</li>
					<li><strong>Group:</strong> %s</li>
				</ul>
				<p>Please keep this information safe for future reference.</p>
				<br/>
				<p>Regards,<br/><strong>Student Cell</strong><br/>University School of Automation & Robotics, GGSIPU</p>
			</body>
			</html>
		`, name, batch, group)

		err = utils.SendHTMLEmail(email, subject, htmlBody)
		if err != nil {
			http.Error(w, "❌ Failed to send email", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("✅ Email sent successfully to " + email))
	}
}
