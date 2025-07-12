package handlers

import (
	"Batch/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type UpgradeValidationRequest struct {
	ApplicationNumber string `json:"application_number"`
	DOB               string `json:"dob"`
	PreviousBranch    string `json:"previous_branch"`
}

type UpgradeValidationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// POST /upgrade/validate
func UpgradeValidateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid Form", http.StatusBadRequest)
			return
		}

		app := r.FormValue("application_number")
		dob := r.FormValue("dob")
		prev := r.FormValue("previous_branch")

		var dbDOB, dbBranch string
		err = db.QueryRow("SELECT dob, branch FROM students WHERE application_number = ?", app).Scan(&dbDOB, &dbBranch)
		if err != nil {
			json.NewEncoder(w).Encode(UpgradeValidationResponse{Success: false, Message: "Student not found"})
			return
		}

		if dbDOB != dob || dbBranch != prev {
			json.NewEncoder(w).Encode(UpgradeValidationResponse{Success: false, Message: "Details mismatch"})
			return
		}

		json.NewEncoder(w).Encode(UpgradeValidationResponse{Success: true})
	}
}

// POST /upgrade/submit
func UpgradeSubmitHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid Form", http.StatusBadRequest)
			return
		}

		app := r.FormValue("application_number")
		newBranch := r.FormValue("new_branch")

		// Fetch current student details
		var fullName, email, previousBranch string
		err = db.QueryRow(`SELECT full_name, email, branch FROM students WHERE application_number = ?`, app).
			Scan(&fullName, &email, &previousBranch)
		if err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		// Assign new batch and group
		batch, group := utils.AssignBatchAndGroup(db, newBranch)

		// Update DB
		_, err = db.Exec(`UPDATE students SET branch = ?, batch = ?, group_name = ? WHERE application_number = ?`,
			newBranch, batch, group, app)
		if err != nil {
			http.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}

		// Prepare and send email
		html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Branch Upgrade Successful</title>
		</head>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<table width="100%%" style="max-width: 600px; margin: auto; background: #fff; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1);">
				<tr style="background-color: #003366; color: white;">
					<td style="padding: 20px;">
						<img src="https://upload.wikimedia.org/wikipedia/en/5/5a/Guru_Gobind_Singh_Indraprastha_University_Logo.png" width="60" style="float: left;">
						<img src="https://upload.wikimedia.org/wikipedia/en/5/5a/Guru_Gobind_Singh_Indraprastha_University_Logo.png" width="60" style="float: right;">
						<h2 style="text-align: center; margin: 0;">University School of Automation & Robotics</h2>
						<p style="text-align: center; margin: 0;">Guru Gobind Singh Indraprastha University, East Delhi Campus</p>
					</td>
				</tr>
				<tr>
					<td style="padding: 20px;">
						<h3>Dear %s,</h3>
						<p>Congratulations! Your branch upgradation has been successfully processed. Please find your new details below:</p>
						<table cellpadding="8" style="width: 100%%; border-collapse: collapse;">
							<tr><td><strong>Application Number</strong></td><td>%s</td></tr>
							<tr><td><strong>Previous Branch</strong></td><td>%s</td></tr>
							<tr><td><strong>New Allotted Branch</strong></td><td>%s</td></tr>
							<tr><td><strong>Batch</strong></td><td>%s</td></tr>
							<tr><td><strong>Group</strong></td><td>%s</td></tr>
						</table>

						<h4>📌 Instructions</h4>
						<ul>
							<li>Please verify the above information.</li>
							<li>If any discrepancies are found, please reply to this email with the correct information.</li>
							<li>Your previous branch record has been updated accordingly in the system.</li>
						</ul>

						<p style="margin-top: 30px;">Regards,<br><strong>USAR Student Cell</strong><br>GGSIPU</p>
					</td>
				</tr>
			</table>
		</body>
		</html>
		`, fullName, app, previousBranch, newBranch, batch, group)

		err = utils.SendHTMLEmail(email, "✅ Branch Upgraded Successfully - USAR", html)
		if err != nil {
			fmt.Println("❌ Failed to send upgrade email:", err)
		} else {
			fmt.Println("✅ Upgrade email sent to:", email)
		}

		// Now redirect
		http.Redirect(w, r, "/confirmation?app="+app, http.StatusSeeOther)
	}
}
