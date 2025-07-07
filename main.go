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

	// ---------------------------
	// Public Routes
	// ---------------------------
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
	http.HandleFunc("/view", handlers.ViewStudentsHandler(db.Conn))
	http.HandleFunc("/export-students", handlers.ExportStudentsHandler(db.Conn))

	// ---------------------------
	// Admin Routes (Protected)
	// ---------------------------
	http.HandleFunc("/admin/dashboard", middleware.AuthMiddleware(handlers.DashboardStatsHandler(db.Conn)))
	http.HandleFunc("/admin/students", middleware.AuthMiddleware(handlers.StudentListHandler(db.Conn)))
	http.HandleFunc("/update-student", middleware.AuthMiddleware(handlers.UpdateStudentHandler(db.Conn)))

	// Scan & Token System
	http.HandleFunc("/scan", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/scan.html")
	}))
	http.HandleFunc("/add_student", middleware.AuthMiddleware(handlers.AddStudentHandler(db.Conn)))
	http.HandleFunc("/get_students", middleware.AuthMiddleware(handlers.GetStudentsHandler(db.Conn)))
	http.HandleFunc("/update_student", middleware.AuthMiddleware(handlers.UpdateStudentStatusHandler(db.Conn)))
	http.HandleFunc("/bulk_update_status", middleware.AuthMiddleware(handlers.BulkUpdateStatusHandler(db.Conn)))
	http.HandleFunc("/update-status", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/update_status.html")
	}))

	// ---------------------------
	// Super Admin Routes (Protected)
	// ---------------------------
	http.HandleFunc("/superadmin", middleware.AuthMiddleware(handlers.SuperAdminPageHandler(db.Conn)))
	http.HandleFunc("/superadmin/create", middleware.AuthMiddleware(handlers.CreateAdminHandler(db.Conn)))
	http.HandleFunc("/superadmin/delete", middleware.AuthMiddleware(handlers.DeleteAdminHandler(db.Conn)))
	http.HandleFunc("/superadmin/export", middleware.AuthMiddleware(handlers.ExportTableHandler(db.Conn)))
	http.HandleFunc("/superadmin/reset", middleware.AuthMiddleware(handlers.ResetStudentsHandler(db.Conn)))
	http.HandleFunc("/superadmin/reset-reporting", middleware.AuthMiddleware(handlers.ResetReportingStudentsHandler(db.Conn)))

	// ---------------------------
	// Static Files
	// ---------------------------
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// ---------------------------
	// Start Server
	// ---------------------------
	log.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
