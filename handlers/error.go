package handlers

import (
	"html/template"
	"net/http"
)

type ErrorPageData struct {
	Message       string
	Error_message error
}

func RenderErrorPage(w http.ResponseWriter, message string, error_message error) {
	tmpl := template.Must(template.ParseFiles("templates/error.html"))
	data := ErrorPageData{Message: message, Error_message: error_message}
	w.WriteHeader(http.StatusBadRequest) // or http.StatusNotFound, etc.
	tmpl.Execute(w, data)
}
