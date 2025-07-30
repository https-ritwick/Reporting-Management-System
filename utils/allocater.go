package utils

import (
	"database/sql"
	"log"
	"strings"
)

func AssignBatchAndGroup(db *sql.DB, branch string) (string, string) {
	log.Println("📥 AssignBatchAndGroup called for branch:", branch)

	// Default groups
	validGroups := []string{"B1-A", "B1-B", "B2-A", "B2-B"}
	counts := map[string]int{}
	for _, g := range validGroups {
		counts[g] = 0
	}

	// Query: count students per batch & group for the given branch who are 'Reported'
	query := `
		SELECT batch, group_name, COUNT(*) as count
		FROM students
		WHERE branch = ? AND status = 'Reported'
		GROUP BY batch, group_name
	`

	log.Println("🔍 Executing group count query...")
	rows, err := db.Query(query, branch)
	if err != nil {
		log.Println("❌ DB query failed:", err)
		log.Println("🛑 Returning default fallback: B1-A")
		return "B1", "A"
	}
	defer rows.Close()

	log.Println("✅ Query successful. Parsing results...")

	for rows.Next() {
		var batch, group string
		var count int

		err := rows.Scan(&batch, &group, &count)
		if err != nil {
			log.Println("⚠️ Row scan failed:", err)
			continue
		}

		// Ensure valid keys
		if batch == "" || group == "" {
			log.Printf("⚠️ Skipping invalid row: batch='%s', group='%s'\n", batch, group)
			continue
		}
		key := batch + "-" + group
		if _, ok := counts[key]; !ok {
			log.Printf("⚠️ Skipping unknown group combo: %s\n", key)
			continue
		}
		counts[key] = count
		log.Printf("📊 %s => %d students\n", key, count)
	}
	if err := rows.Err(); err != nil {
		log.Println("⚠️ Rows iteration error:", err)
	}

	log.Println("📋 Final group counts:")
	for _, g := range validGroups {
		log.Printf("   %s => %d\n", g, counts[g])
	}

	// Choose least-filled group
	minKey := validGroups[0]
	minCount := counts[minKey]
	for _, key := range validGroups {
		if counts[key] < minCount {
			minKey = key
			minCount = counts[key]
		}
	}

	// Safe split
	parts := strings.Split(minKey, "-")
	if len(parts) != 2 {
		log.Println("❌ minKey format error. Using fallback B1-A")
		return "B1", "A"
	}

	batch := parts[0]
	group := parts[1]
	log.Printf("✅ Selected least-filled group: %s (%d students)\n", minKey, minCount)
	log.Printf("🚀 Returning Batch: %s, Group: %s\n", batch, group)

	return batch, group
}
