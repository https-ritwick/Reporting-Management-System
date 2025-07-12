package handlers

import (
	"Batch/models"
	"Batch/utils"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Error loading .env file")
	}
}

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
		http.Redirect(w, r, "/confirmation?app="+student.ApplicationNumber, http.StatusSeeOther)
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
					<tr><td><strong>Allotted Branch</strong></td><td>%s</td></tr>
					<tr><td><strong>Batch</strong></td><td>%s</td></tr>
					<tr><td><strong>Group</strong></td><td>%s</td></tr>
					</table>

					<h4><strong> Important Instructions </strong></h4>
					<ul>
					<li>Please ensure all details are correct.</li>
					<li>Please Note Down your Allotted Batch & Group for Future Reference</li>
					<li>Students may fill the Hostel Admission Form available on the University Website.</li>
					<li>If any discrepancies are found, Students are required to reply back to this mail, with detailed explanation and corrections, if any.</li>
					<li>Join the official WhatsApp group shared during orientation for updates.</li>
					</ul>

					<p style="margin-top: 30px;">Regards,<br><strong>USAR Student Cell</strong><br>GGSIPU</p>
				</td>
				</tr>
			</table>S
			</body>
			</html>
		`, student.FullName, student.ApplicationNumber, student.Branch, student.Batch, student.Group)

		err = utils.SendHTMLEmail(student.Email, "🎓 Registration Confirmation - USAR", html)
		if err != nil {
			log.Println("❌ Failed to send registration email:", err)
		} else {
			log.Println("✅ Email sent to:", student.Email)
		}
		// Redirect to confirmation page
	}
}
