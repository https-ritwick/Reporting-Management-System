package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type CutoffRequest struct {
	Branch   string `json:"branch"`
	Quota    string `json:"quota"`
	Category string `json:"category"`
	Gender   string `json:"gender"`
}

type CutoffResponse struct {
	MinRank *int `json:"min_rank"`
	MaxRank *int `json:"max_rank"`
	Count   int  `json:"count"`
}

func CutoffHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var req CutoffRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		query := `SELECT exam_rank FROM students WHERE status = 'Reported' AND lateral_entry = 0`
		args := []interface{}{}

		if req.Branch != "All" {
			query += " AND branch = ?"
			args = append(args, req.Branch)
		}
		if req.Quota != "All" {
			query += " AND seat_quota = ?"
			args = append(args, req.Quota)
		}
		if req.Category != "All" {
			query += " AND category = ?"
			args = append(args, req.Category)
		}
		if req.Gender != "All" {
			query += " AND gender = ?"
			args = append(args, req.Gender)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var ranks []int
		for rows.Next() {
			var r int
			if err := rows.Scan(&r); err == nil {
				ranks = append(ranks, r)
			}
		}

		if len(ranks) == 0 {
			json.NewEncoder(w).Encode(CutoffResponse{
				MinRank: nil,
				MaxRank: nil,
				Count:   0,
			})
			return
		}

		min, max := ranks[0], ranks[0]
		for _, r := range ranks {
			if r < min {
				min = r
			}
			if r > max {
				max = r
			}
		}

		json.NewEncoder(w).Encode(CutoffResponse{
			MinRank: &min,
			MaxRank: &max,
			Count:   len(ranks),
		})
	}
}
