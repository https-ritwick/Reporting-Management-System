package handlers

import (
	"Batch/utils"
	"database/sql"
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

			batch, group := utils.AssignBatchAndGroup(db, newBranch)

			_, err := db.Exec(`UPDATE students 
				SET branch = ?, batch = ?, group_name = ?, last_upgrade_at = ? 
				WHERE application_number = ?`,
				newBranch, batch, group, time.Now(), app)

			if err != nil {
				http.Error(w, "Database update failed", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/confirmation?app="+app, http.StatusSeeOther)
			return
		}

		tmpl.Execute(w, UpgradeFormData{Step: "verify"})
	}
}
