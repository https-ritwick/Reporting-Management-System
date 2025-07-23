// handlers/super_admin.go
package handlers

import (
	"archive/zip"
	"crypto/sha512"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"html/template"
)

type Admin struct {
	ID    int
	Email string
	Role  string
}

func SuperAdminPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/super_admin.html")
		if err != nil {
			http.Error(w, "Template parsing error", http.StatusInternalServerError)
			return
		}
		rows, err := db.Query("SELECT id, email, role FROM user_login")
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		admins := []Admin{}
		for rows.Next() {
			var a Admin
			if err := rows.Scan(&a.ID, &a.Email, &a.Role); err == nil {
				admins = append(admins, a)
			}
		}
		tmpl.Execute(w, admins)
	}
}

func DeleteAdminHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		_, err := db.Exec("DELETE FROM user_login WHERE id = ?", id)
		if err != nil {
			http.Error(w, "Delete failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/superadmin", http.StatusSeeOther)
	}
}

func ExportTableHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		table := r.URL.Query().Get("table")
		if table == "" {
			http.Error(w, "Missing table name", http.StatusBadRequest)
			return
		}
		rows, err := db.Query("SELECT * FROM " + table)
		if err != nil {
			http.Error(w, "Query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		cols, _ := rows.Columns()
		csvWriter := csv.NewWriter(w)
		w.Header().Set("Content-Disposition", "attachment; filename="+table+".csv")
		w.Header().Set("Content-Type", "text/csv")
		csvWriter.Write(cols)
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for rows.Next() {
			for i := range cols {
				valPtrs[i] = &vals[i]
			}
			rows.Scan(valPtrs...)
			s := make([]string, len(cols))
			for i, val := range vals {
				if b, ok := val.([]byte); ok {
					s[i] = string(b)
				} else if val == nil {
					s[i] = ""
				} else {
					s[i] = fmt.Sprintf("%v", val)
				}
			}
			csvWriter.Write(s)
		}
		csvWriter.Flush()
	}
}

func ResetStudentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		keyword := r.FormValue("confirm")
		if keyword != "SUMMIT" {
			http.Error(w, "Invalid confirmation keyword", http.StatusForbidden)
			return
		}

		// Set up ZIP response
		filename := fmt.Sprintf("backup_%s.zip", time.Now().Format("2006-01-02_15-04-05"))
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		zipWriter := zip.NewWriter(w)

		// Helper: Export CSV and write to zip
		writeCSVToZip := func(query, name string) error {
			rows, err := db.Query(query)
			if err != nil {
				return err
			}
			defer rows.Close()

			cols, _ := rows.Columns()
			vals := make([]interface{}, len(cols))
			valPtrs := make([]interface{}, len(cols))

			fileWriter, err := zipWriter.Create(name + ".csv")
			if err != nil {
				return err
			}
			csvWriter := csv.NewWriter(fileWriter)
			csvWriter.Write(cols)

			for rows.Next() {
				for i := range cols {
					valPtrs[i] = &vals[i]
				}
				rows.Scan(valPtrs...)
				row := make([]string, len(cols))
				for i, val := range vals {
					if b, ok := val.([]byte); ok {
						row[i] = string(b)
					} else if val == nil {
						row[i] = ""
					} else {
						row[i] = fmt.Sprintf("%v", val)
					}
				}
				csvWriter.Write(row)
			}
			csvWriter.Flush()
			return nil
		}

		// Write both tables to zip
		if err := writeCSVToZip("SELECT * FROM students", "students"); err != nil {
			http.Error(w, "Error exporting students", http.StatusInternalServerError)
			return
		}
		if err := writeCSVToZip("SELECT * FROM uploads", "uploads"); err != nil {
			http.Error(w, "Error exporting uploads", http.StatusInternalServerError)
			return
		}

		// Close ZIP before truncating
		if err := zipWriter.Close(); err != nil {
			http.Error(w, "Failed to finalize ZIP", http.StatusInternalServerError)
			return
		}

		// Truncate tables
		if _, err := db.Exec("TRUNCATE TABLE students"); err != nil {
			fmt.Println("Error truncating students:", err)
		}
		if _, err := db.Exec("TRUNCATE TABLE uploads"); err != nil {
			fmt.Println("Error truncating uploads:", err)
		}
	}
}

func ResetReportingStudentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			keyword := r.FormValue("confirm-report")
			if keyword != "SUMMIT" {
				http.Error(w, "Invalid confirmation keyword", http.StatusForbidden)
				return
			}

			rows, err := db.Query("SELECT * FROM reporting_status")
			if err != nil {
				http.Error(w, "Query error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			cols, _ := rows.Columns()
			csvWriter := csv.NewWriter(w)
			w.Header().Set("Content-Disposition", "attachment; filename=reporting_status.csv")
			w.Header().Set("Content-Type", "text/csv")
			csvWriter.Write(cols)
			vals := make([]interface{}, len(cols))
			valPtrs := make([]interface{}, len(cols))
			for rows.Next() {
				for i := range cols {
					valPtrs[i] = &vals[i]
				}
				rows.Scan(valPtrs...)
				s := make([]string, len(cols))
				for i, val := range vals {
					if b, ok := val.([]byte); ok {
						s[i] = string(b)
					} else if val == nil {
						s[i] = ""
					} else {
						s[i] = fmt.Sprintf("%v", val)
					}
				}
				csvWriter.Write(s)
			}
			csvWriter.Flush()
			_, err = db.Exec("TRUNCATE TABLE reporting_status")
			if err != nil {
				http.Error(w, "Truncate failed", http.StatusInternalServerError)
				return
			}
		}
	}
}

// CreateAdminHandler handles creation of new admin accounts
func CreateAdminHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		role := r.FormValue("role")

		if email == "" || password == "" || role == "" {
			http.Error(w, "Missing fields", http.StatusBadRequest)
			return
		}

		// Hash the password using SHA512
		hash := sha512.Sum512([]byte(password))
		hashedPassword := fmt.Sprintf("%x", hash[:])

		// Insert into user_details table
		_, err := db.Exec("INSERT INTO user_login (email, password_hash, role) VALUES (?, ?, ?)", email, hashedPassword, role)
		if err != nil {
			http.Error(w, "Failed to create admin", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		http.Redirect(w, r, "/superadmin", http.StatusSeeOther)
	}
}
