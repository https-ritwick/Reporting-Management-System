package handlers

import (
	"Batch/utils"
	"database/sql"
	"fmt"
	"html/template"
	"log"
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
		var email string

		query := `SELECT  dob, branch, full_name, last_upgrade_at,lateral_entry, email FROM students WHERE application_number = ?`
		err := db.QueryRow(query, appNum).Scan(&dbDOB, &dbPrevBranch, &fullName, &lastUpgradeAt, &lateralEntry, &email)
		if err != nil {
			RenderErrorPage(w, "Student not found. Please check your details.", err)
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
			RenderErrorPage(w, "DOB or Previous Branch does not match our records", nil)
			return
		}

		// Enforce 3-day cooldown
		if lastUpgradeAt.Valid {
			duration := time.Since(lastUpgradeAt.Time)
			if duration.Hours() < 72 {
				RenderErrorPage(w, "Upgrade not allowed in this Round of Counselling", nil)
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
		groupDisplay := group
		batchDisplay := batch
		if lateralEntry == 1 {
			batchDisplay = "Not Applicable (LE Student)"
			groupDisplay = "Not Applicable (LE Student)"
		}
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
					<p>Your branch has been successfully updated. You can start attending your classes in the new branch according to the assigned batch and group</p>
					<table cellpadding="8" style="width: 100%%; border-collapse: collapse;">
					<tr><td><strong>Application Number</strong></td><td>%s</td></tr>
					<tr><td><strong>Branch</strong></td><td>%s</td></tr>
					<tr><td><strong>Allotted Batch</strong></td><td>%s</td></tr>
					<tr><td><strong>Allotted Group</strong></td><td>%s</td></tr>
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
			</table>S
			</body>
			</html>
		`, fullName, appNum, newBranch, batchDisplay, groupDisplay)

		err = utils.SendHTMLEmail(email, "Branch Upgrade Confirmation - USAR", html)
		if err != nil {
			log.Println("❌ Failed to send registration email:", err)
		} else {
			log.Println("✅ Email sent to:", email)
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
