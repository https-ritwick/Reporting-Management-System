package handlers

import (
	"Batch/models"
	"Batch/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

func SubmitHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form values
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid Form Submission", http.StatusBadRequest)
			return
		}

		// Convert rank to integer
		rank, _ := strconv.Atoi(r.FormValue("rank"))

		// Create Student object
		student := models.Student{
			FullName:           r.FormValue("full_name"),
			ApplicationNumber:  r.FormValue("application_number"),
			FatherName:         r.FormValue("father_name"),
			DOB:                r.FormValue("dob"),
			ContactNumber:      r.FormValue("contact_number"),
			Email:              r.FormValue("email"),
			CorrespondenceAddr: r.FormValue("correspondence_address"),
			PermanentAddr:      r.FormValue("permanent_address"),
			Branch:             r.FormValue("branch"),
			LateralEntry:       r.FormValue("lateral_entry"),
			Category:           r.FormValue("category"),
			SubCategory:        r.FormValue("sub_category"),
			Rank:               rank,
			SeatQuota:          r.FormValue("seat_quota"),
			Status:             "Reported",
		}
		var lateralInt int
		if student.LateralEntry == "Yes" {
			lateralInt = 1
		} else {
			lateralInt = 0
		}

		// Batch & group allocation logic
		student.Batch, student.Group = utils.AssignBatchAndGroup(db, student.Branch)

		// Final insert query using double quotes and escaped backticked column `rank`
		insertQuery := "INSERT INTO students (" +
			"full_name, application_number, father_name, dob, contact_number, " +
			"email, correspondence_address, permanent_address, branch, lateral_entry, " +
			"category, sub_category, exam_rank, seat_quota, batch, group_name, status" +
			") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

		_, err = db.Exec(insertQuery,
			student.FullName, student.ApplicationNumber, student.FatherName, student.DOB,
			student.ContactNumber, student.Email, student.CorrespondenceAddr, student.PermanentAddr,
			student.Branch, lateralInt, student.Category, student.SubCategory,
			student.Rank, student.SeatQuota, student.Batch, student.Group, student.Status,
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to confirmation page
		http.Redirect(w, r, "/confirmation?app="+student.ApplicationNumber, http.StatusSeeOther)
	}
}
