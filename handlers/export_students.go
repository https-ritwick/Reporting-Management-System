package handlers

import (
	"database/sql"

	"net/http"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ExportStudentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		branch := r.URL.Query().Get("branch")
		batch := r.URL.Query().Get("batch")
		group := r.URL.Query().Get("group")

		query := `SELECT application_number, full_name, branch, batch,  group_name  FROM students WHERE status = 'Reported' AND lateral_entry = 'No'`
		args := []interface{}{}

		if search != "" {
			query += ` AND (UPPER(full_name) LIKE ? OR application_number LIKE ?)`
			searchTerm := "%" + strings.ToLower(search) + "%"
			args = append(args, searchTerm, searchTerm)
		}
		if branch != "" {
			query += ` AND branch = ?`
			args = append(args, branch)
		}
		if batch != "" {
			query += ` AND batch = ?`
			args = append(args, batch)
		}
		if group != "" {
			query += ` AND ` + "`group`" + ` = ?`
			args = append(args, group)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		f := excelize.NewFile()
		sheet, err := f.NewSheet("Students")
		if err != nil {
			http.Error(w, "Dwonloading Excel Error: "+err.Error(), http.StatusInternalServerError)
		}
		f.SetActiveSheet(sheet)
		f.SetSheetRow("Students", "A1", &[]interface{}{"Application No", "Full Name", "Branch", "Batch", "Group"})

		rowIndex := 2
		for rows.Next() {
			var appNo, name, branch, batch, group string
			err := rows.Scan(&appNo, &name, &branch, &batch, &group)
			if err != nil {
				http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			row := []interface{}{appNo, name, branch, batch, group}
			cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
			f.SetSheetRow("Students", cell, &row)
			rowIndex++
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="Reported_Students.xlsx"`)
		f.Write(w)
	}
}
