package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type FormSettings struct {
	Error                string
	Success              string
	AcademicSession      string
	RegistrationFormOpen bool
	UpgradeFormOpen      bool
	EditFormOpen         bool
	RegistrationDeadline string
	UpgradeDeadline      string
	MaxStudentsPerGroup  int
	PublicNotice         string
}

func FormControlsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/form_controls.html"))

		if r.Method == http.MethodPost {
			// Parse form
			err := r.ParseForm()
			if err != nil {
				log.Println("Form parse error:", err)
				tmpl.Execute(w, FormSettings{Error: "Failed to parse form"})
				return
			}

			// Extract values
			academicSession := r.FormValue("academic_session")
			regFormOpen := r.FormValue("registration_open") == "on"
			upgradeFormOpen := r.FormValue("upgrade_open") == "on"
			editFormOpen := r.FormValue("edit_open") == "on"
			publicNotice := r.FormValue("public_notice")

			// Deadlines: allow NULL
			var regDeadline, upgDeadline interface{}
			if r.FormValue("registration_deadline") != "" {
				regDeadline = r.FormValue("registration_deadline")
			} else {
				regDeadline = nil
			}
			if r.FormValue("upgrade_deadline") != "" {
				upgDeadline = r.FormValue("upgrade_deadline")
			} else {
				upgDeadline = nil
			}

			// Max students per group
			maxGroup, err := strconv.Atoi(r.FormValue("max_students_per_group"))
			if err != nil {
				maxGroup = 0
			}

			// Update DB
			_, err = db.Exec(`
				UPDATE settings 
				SET academic_session=?, registration_form_open=?, upgrade_form_open=?, edit_form_open=?,
					registration_deadline=?, upgrade_deadline=?, max_students_per_group=?, public_notice=?
				WHERE id = 1
			`, academicSession, regFormOpen, upgradeFormOpen, editFormOpen, regDeadline, upgDeadline, maxGroup, publicNotice)

			if err != nil {
				log.Println("Update failed:", err)
				tmpl.Execute(w, FormSettings{Error: "Update failed"})
				return
			}

			// Success
			tmpl.Execute(w, FormSettings{
				Success:              "Settings saved successfully!",
				AcademicSession:      academicSession,
				RegistrationFormOpen: regFormOpen,
				UpgradeFormOpen:      upgradeFormOpen,
				EditFormOpen:         editFormOpen,
				RegistrationDeadline: r.FormValue("registration_deadline"),
				UpgradeDeadline:      r.FormValue("upgrade_deadline"),
				MaxStudentsPerGroup:  maxGroup,
				PublicNotice:         publicNotice,
			})
			return
		}

		// GET method: load values from DB
		var fs FormSettings
		var regDeadline, upgDeadline sql.NullString

		err := db.QueryRow(`
			SELECT academic_session, registration_form_open, upgrade_form_open, edit_form_open,
			       registration_deadline, upgrade_deadline, max_students_per_group, public_notice
			FROM settings WHERE id = 1
		`).Scan(&fs.AcademicSession, &fs.RegistrationFormOpen, &fs.UpgradeFormOpen, &fs.EditFormOpen,
			&regDeadline, &upgDeadline, &fs.MaxStudentsPerGroup, &fs.PublicNotice)

		if err != nil {
			log.Println("Load failed:", err)
			fs.Error = "Load error"
			tmpl.Execute(w, fs)
			return
		}

		if regDeadline.Valid {
			fs.RegistrationDeadline = regDeadline.String
		}
		if upgDeadline.Valid {
			fs.UpgradeDeadline = upgDeadline.String
		}

		tmpl.Execute(w, fs)
	}
}
