package handlers

import (
	"html/template"
	"net/http"
)

func AboutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/team.html"))
		tmpl.Execute(w, nil)
	}
}
