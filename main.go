package main

import (
	"log"
	"net/http"
	"text/template"

	"Batch/db"
	"Batch/handlers"
	"Batch/middleware"
)

func main() {
	// Initialize DB connection
	db.Init()

	// Public Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/home.html"))
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/form.html"))
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/faq", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/faq.html"))
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/submit", handlers.SubmitHandler(db.Conn))
	http.HandleFunc("/confirmation", handlers.ConfirmationHandler(db.Conn))
	http.HandleFunc("/login", handlers.LoginHandler(db.Conn))
	http.HandleFunc("/upgrade", handlers.UpgradeHandler(db.Conn))
	http.HandleFunc("/edit", handlers.EditHandler(db.Conn))

	//http.HandleFunc("/logout", handlers.LogoutHandler)

	// Protected Admin Routes
	http.HandleFunc("/admin/dashboard", middleware.AuthMiddleware(handlers.DashboardStatsHandler(db.Conn)))
	http.HandleFunc("/admin/students", middleware.AuthMiddleware(handlers.StudentListHandler(db.Conn)))
	http.HandleFunc("/admin/form-controls", middleware.AuthMiddleware(handlers.FormControlsHandler(db.Conn)))
	http.HandleFunc("/update-student", handlers.UpdateStudentHandler(db.Conn))

	//Scan and Token System
	// Route: QR Scan + Manual Form Page (served from /scan.html or index.html)
	http.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/scan.html")
	})

	// API route to add student
	http.HandleFunc("/add_student", handlers.AddStudentHandler(db.Conn))

	// Route: Display token page (HTML rendered via template)
	http.HandleFunc("/get_students", handlers.GetStudentsHandler(db.Conn))
	http.HandleFunc("/update_student", handlers.UpdateStudentStatusHandler(db.Conn))
	http.HandleFunc("/bulk_update_status", handlers.BulkUpdateStatusHandler(db.Conn))
	http.HandleFunc("/update-status", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/update_status.html")
	})

	// Static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start server
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
