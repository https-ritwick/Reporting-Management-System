package handlers

import (
	"Batch/utils"
	"database/sql"
	"encoding/json"
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

		batch, group := utils.AssignBatchAndGroup(db, newBranch)

		_, err = db.Exec(`UPDATE students SET branch = ?, batch = ?, group_name = ? WHERE application_number = ?`,
			newBranch, batch, group, app)
		if err != nil {
			http.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/confirmation?app="+app, http.StatusSeeOther)
	}
}
