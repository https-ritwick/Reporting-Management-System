package handlers

import (
	"Batch/models"
	"Batch/utils"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

		err := r.ParseMultipartForm(20 << 20) // 20MB limit
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Extract and parse fields
		rank, _ := strconv.Atoi(r.FormValue("rank"))
		appNo := r.FormValue("application_number")
		fullName := r.FormValue("full_name")

		student := models.Student{
			FullName:           fullName,
			ApplicationNumber:  appNo,
			FatherName:         r.FormValue("father_name"),
			DOB:                r.FormValue("dob"),
			Gender:             r.FormValue("gender"),
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
			FeeMode:            r.FormValue("fee_mode"),
			FeeReference:       r.FormValue("fee_reference"),
			Status:             "Reported",
		}
		student.Batch = ""
		student.Group = ""
		var lateralInt int
		var isLE int
		if student.LateralEntry == "Yes" {
			isLE = 1
		} else {
			isLE = 0
			student.Batch, student.Group = utils.AssignBatchAndGroup(db, student.Branch)
		}

		// Save student in students table
		insertQuery := `INSERT INTO students (
			application_number, full_name, father_name, dob, gender, contact_number,
			email, correspondence_address, permanent_address, branch, lateral_entry,
			category, sub_category, exam_rank, seat_quota, batch, group_name, status,
			has_edited, fee_mode, fee_reference
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = db.Exec(insertQuery,
			student.ApplicationNumber, student.FullName, student.FatherName, student.DOB, student.Gender,
			student.ContactNumber, student.Email, student.CorrespondenceAddr, student.PermanentAddr,
			student.Branch, isLE, student.Category, student.SubCategory, student.Rank,
			student.SeatQuota, student.Batch, student.Group, student.Status, 0,
			student.FeeMode, student.FeeReference,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		// --- Handle File Uploads ---
		uploadDir := fmt.Sprintf("static/uploads/%s", appNo)
		os.MkdirAll(uploadDir, os.ModePerm)

		paths := map[string]string{}
		files := map[string]string{
			"photo":             "photo",
			"jee_scorecard":     "jee_scorecard",
			"candidate_profile": "candidate_profile",
			"fee_receipt":       "fee_receipt",
			"reporting_slip":    "reporting_slip",
		}

		for field, filename := range files {
			file, handler, err := r.FormFile(field)
			if err != nil {
				log.Printf("❌ Missing file for %s: %v\n", field, err)
				continue
			}
			defer file.Close()

			filePath := fmt.Sprintf("%s/%s%s", uploadDir, filename, filepath.Ext(handler.Filename))
			dst, err := os.Create(filePath)
			if err != nil {
				log.Printf("❌ Failed to create file: %v\n", err)
				continue
			}
			defer dst.Close()

			_, err = io.Copy(dst, file)
			if err != nil {
				log.Printf("❌ Failed to save file: %v\n", err)
				continue
			}

			paths[field+"_path"] = filePath
		}

		// --- Insert into uploads table ---
		_, err = db.Exec(`INSERT INTO uploads (
			application_number, full_name,
			photo_path, jee_scorecard_path, candidate_profile_path,
			fee_receipt_path, reporting_slip_path
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			appNo, fullName,
			paths["photo_path"],
			paths["jee_scorecard_path"],
			paths["candidate_profile_path"],
			paths["fee_receipt_path"],
			paths["reporting_slip_path"],
		)
		if err != nil {
			log.Println("❌ Failed to insert into uploads table:", err)
		}
		groupDisplay := student.Group
		batchDisplay := student.Batch
		if lateralInt == 1 {
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
					<p>Welcome to University School of Automation & Robotics! Thank you for registering. Your details have been successfully recorded in the system. Please find your details below:</p>
					<table cellpadding="8" style="width: 100%%; border-collapse: collapse;">
					<tr><td><strong>Application Number</strong></td><td>%s</td></tr>
					<tr><td><strong>Allotted Branch</strong></td><td>%s</td></tr>
					<tr><td><strong>Batch</strong></td><td>%s</td></tr>
					<tr><td><strong>Group</strong></td><td>%s</td></tr>
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
		`, student.FullName, student.ApplicationNumber, student.Branch, batchDisplay, groupDisplay)

		err = utils.SendHTMLEmail(student.Email, "Registration Confirmation - USAR", html)
		if err != nil {
			log.Println("❌ Failed to send registration email:", err)
		} else {
			log.Println("✅ Email sent to:", student.Email)
		}
		// Redirect to confirmation page
		http.Redirect(w, r, "/confirmation?app="+student.ApplicationNumber, http.StatusSeeOther)
	}
}
