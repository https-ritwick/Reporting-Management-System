package handlers

import (
	"Batch/utils"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ConfirmationData struct {
	FullName          string
	ApplicationNumber string
	Branch            string
	Batch             string
	Group             string
}

func UpgradeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl := template.Must(template.ParseFiles("templates/upgrade.html"))
			tmpl.Execute(w, nil)
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		appNum := r.FormValue("application_number")
		dob := r.FormValue("dob")
		prevBranch := r.FormValue("prev_branch")
		newBranch := r.FormValue("new_branch")

		var dbDOB time.Time
		var dbPrevBranch, fullName string
		var lastUpgradeAt sql.NullTime
		var lateralEntry int

		query := `SELECT dob, branch, full_name, last_upgrade_at,lateral_entry FROM students WHERE application_number = ?`
		err := db.QueryRow(query, appNum).Scan(&dbDOB, &dbPrevBranch, &fullName, &lastUpgradeAt, &lateralEntry)
		if err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		// Verify DOB and previous branch
		// Parse form DOB into time.Time
		parsedDOB, err := time.Parse("2006-01-02", dob)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		fmt.Println(parsedDOB, dbDOB)
		fmt.Println(prevBranch, dbPrevBranch)
		// Compare DOB and Previous Branch
		if !parsedDOB.Equal(dbDOB) || prevBranch != dbPrevBranch {
			http.Error(w, "DOB or Previous Branch does not match our records", http.StatusForbidden)
			return
		}

		// Enforce 3-day cooldown
		if lastUpgradeAt.Valid {
			duration := time.Since(lastUpgradeAt.Time)
			if duration.Hours() < 72 {
				http.Error(w, "Upgrade not allowed in this Round of Counselling", http.StatusForbidden)
				return
			}
		}
		batch := ""
		group := ""
		fmt.Println(lateralEntry)

		// Assign batch/group only if not lateral
		if lateralEntry == 0 {
			batch, group = utils.AssignBatchAndGroup(db, newBranch)
		}

		// Update student's new branch and batch/group
		updateQuery := `
			UPDATE students 
			SET branch = ?, batch = ?, group_name = ?, last_upgrade_at = NOW() 
			WHERE application_number = ?`
		_, err = db.Exec(updateQuery, newBranch, batch, group, appNum)
		if err != nil {
			http.Error(w, "Failed to update student data", http.StatusInternalServerError)
			return
		}

		// Handle file upload
		file, _, err := r.FormFile("reporting_slip")
		if err != nil {
			http.Error(w, "Failed to receive reporting slip", http.StatusBadRequest)
			return
		}
		defer file.Close()

		saveDir := filepath.Join("static", "uploads", appNum)
		os.MkdirAll(saveDir, os.ModePerm)

		filePath := filepath.Join(saveDir, "upgrade_reporting_slip.pdf")

		outFile, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = outFile.ReadFrom(file)
		if err != nil {
			http.Error(w, "Error saving file content", http.StatusInternalServerError)
			return
		}
		// Update uploads table with upgrade_reporting_slip path
		relativePath := filepath.ToSlash(filepath.Join("static", "uploads", appNum, "upgrade_reporting_slip.pdf"))
		_, err = db.Exec(`UPDATE uploads SET reporting_slip_path = ? WHERE application_number = ?`, relativePath, appNum)
		if err != nil {
			http.Error(w, "Failed to update uploads table", http.StatusInternalServerError)
			return
		}

		// Render confirmation
		confirmation := ConfirmationData{
			FullName:          fullName,
			ApplicationNumber: appNum,
			Branch:            newBranch,
			Batch:             batch,
			Group:             group,
		}

		tmpl := template.Must(template.ParseFiles("templates/confirmation.html"))
		tmpl.Execute(w, confirmation)
	}
}
