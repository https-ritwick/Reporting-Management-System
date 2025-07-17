package handlers

import (
	"Batch/utils"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type UpgradeFormData struct {
	Step            string
	Error           string
	AppNumber       string
	PrevBranch      string
	AllowedBranches []string
}

func UpgradeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/upgrade.html"))

		if r.Method == http.MethodGet || r.URL.Query().Get("step") == "" {
			tmpl.Execute(w, UpgradeFormData{Step: "verify"})
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}

		step := r.URL.Query().Get("step")

		// Step 2: Verification
		if step == "verify" {
			app := r.FormValue("application_number")
			dob := r.FormValue("dob")
			prevBranch := r.FormValue("prev_branch")

			var dbDOBRaw, dbBranch string
			var lastUpgrade sql.NullTime

			err := db.QueryRow("SELECT dob, branch, last_upgrade_at FROM students WHERE application_number = ?", app).
				Scan(&dbDOBRaw, &dbBranch, &lastUpgrade)

			if err != nil {
				tmpl.Execute(w, UpgradeFormData{Step: "verify", Error: "Student not found"})
				return
			}

			if len(dbDOBRaw) < 10 || dbDOBRaw[:10] != dob || dbBranch != prevBranch {
				tmpl.Execute(w, UpgradeFormData{
					Step:  "verify",
					Error: "Invalid details. Please check and try again.",
				})
				return
			}

			// Restrict upgrades within 3 days
			if lastUpgrade.Valid {
				daysSince := time.Since(lastUpgrade.Time).Hours() / 24
				if daysSince < 3 {
					tmpl.Execute(w, UpgradeFormData{
						Step:  "verify",
						Error: "You can only upgrade your branch once in this Round of Counselling",
					})
					return
				}
			}

			allBranches := []string{"AI-DS", "AI-ML", "IIOT", "A&R"}
			allowed := []string{}
			for _, b := range allBranches {
				if b != prevBranch {
					allowed = append(allowed, b)
				}
			}

			tmpl.Execute(w, UpgradeFormData{
				Step:            "upgrade",
				AppNumber:       app,
				PrevBranch:      prevBranch,
				AllowedBranches: allowed,
			})
			return
		}

		// Step 3: Perform upgrade
		if step == "submit" {
			app := r.FormValue("application_number")
			newBranch := r.FormValue("new_branch")

			// Fetch current student info
			var fullName, email, prevBranch string
			var lateralInt int
			err := db.QueryRow("SELECT full_name, email, branch, lateral_entry FROM students WHERE application_number = ?", app).
				Scan(&fullName, &email, &prevBranch, &lateralInt)
			if err != nil {
				http.Error(w, "Student not found", http.StatusNotFound)
				return
			}

			// Allot new batch & group based on LE status
			var batch, group string
			if lateralInt == 1 {
				batch = ""
				group = ""
			} else {
				// Regular student
				batch, group = utils.AssignBatchAndGroup(db, newBranch)
			}

			// Update DB
			_, err = db.Exec(`UPDATE students 
				SET branch = ?, batch = ?, group_name = ?, last_upgrade_at = ? 
				WHERE application_number = ?`,
				newBranch, batch, group, time.Now(), app)

			if err != nil {
				http.Error(w, "Database update failed", http.StatusInternalServerError)
				return
			}

			// Group display message
			groupDisplay := group
			batchDisplay := batch
			if lateralInt == 1 {
				batchDisplay = "Not Applicable (LE Student)"
				groupDisplay = "Not Applicable (LE Student)"
			}

			// Email HTML content
			html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
			<meta charset="UTF-8">
			<title>Branch Upgrade Confirmation</title>
			</head>
			<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<table width="100%%" style="max-width: 600px; margin: auto; background: #fff; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1);">
				<tr style="background-color: #003366; color: white;">
				<td style="padding: 20px;">
					<img src="https://upload.wikimedia.org/wikipedia/en/5/5a/Guru_Gobind_Singh_Indraprastha_University_Logo.png" width="60" style="float: left;">
					<h2 style="text-align: center; margin: 0;">University School of Automation & Robotics</h2>
					<p style="text-align: center; margin: 0;">Guru Gobind Singh Indraprastha University, East Delhi Campus</p>
				</td>
				</tr>
				<tr>
				<td style="padding: 20px;">
					<h3>Dear %s,</h3>
					<p>Congratulations! Your branch upgradation has been successfully processed. Please find your updated details below:</p>
					<table cellpadding="8" style="width: 100%%; border-collapse: collapse;">
					<tr><td><strong>Application Number</strong></td><td>%s</td></tr>
					<tr><td><strong>Previous Branch</strong></td><td>%s</td></tr>
					<tr><td><strong>New Upgraded Branch</strong></td><td>%s</td></tr>
					<tr><td><strong>Batch</strong></td><td>%s</td></tr>
					<tr><td><strong>Group</strong></td><td>%s</td></tr>
					</table>

					<h4>Instructions</h4>
					<ul>
					<li>Please verify the above information.</li>
					<li>If any discrepancies are found, reply to this email with the correct details.</li>
					<li>Your records have been successfully updated in the Student Management System.</li>
					</ul>

					<p style="margin-top: 30px;">Regards,<br><strong>USAR Student Cell</strong><br>GGSIPU</p>
				</td>
				</tr>
			</table>
			</body>
			</html>
			`, fullName, app, prevBranch, newBranch, batchDisplay, groupDisplay)

			// Send email
			err = utils.SendHTMLEmail(email, "Branch Upgraded Successfully - USAR", html)
			if err != nil {
				fmt.Println("❌ Failed to send email:", err.Error())
			} else {
				fmt.Println("✅ Upgrade email sent to:", email)
			}

			http.Redirect(w, r, "/confirmation?app="+app, http.StatusSeeOther)
			return
		}
	}
}
