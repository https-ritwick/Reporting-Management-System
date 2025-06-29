package utils

import (
	"database/sql"
	"log"
)

func AssignBatchAndGroup(db *sql.DB, branch string) (string, string) {
	type GroupCount struct {
		Batch string
		Group string
		Count int
	}

	// Query: count students per batch & group for the given branch who are 'Reported'
	query := `
		SELECT batch, group_name, COUNT(*) as count
		FROM students
		WHERE branch = ? AND status = 'Reported'
		GROUP BY batch, group_name
	`

	rows, err := db.Query(query, branch)
	if err != nil {
		log.Println("Error fetching group counts:", err)
		// fallback default
		return "B1", "A"
	}
	defer rows.Close()

	// Initialize all groups to 0
	counts := map[string]int{
		"B1-A": 0,
		"B1-B": 0,
		"B2-A": 0,
		"B2-B": 0,
	}

	for rows.Next() {
		var batch, group string
		var count int
		if err := rows.Scan(&batch, &group, &count); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		key := batch + "-" + group
		counts[key] = count
	}

	// Find the least filled group
	minKey := "B1-A"
	minCount := counts[minKey]
	for key, count := range counts {
		if count < minCount {
			minKey = key
			minCount = count
		}
	}

	// Return batch and group from minKey
	batch := minKey[:2] // B1 or B2
	group := minKey[3:] // A or B

	return batch, group
}
