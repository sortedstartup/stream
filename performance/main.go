package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Comment struct {
	ID              string  `json:"id"`
	Content         string  `json:"content"`
	VideoID         string  `json:"video_id"`
	UserID          string  `json:"user_id"`
	ParentCommentID *string `json:"parent_comment_id"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func main() {
	db, err := sql.Open("sqlite3", "./comments.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	createTable(db)

	insertComments(db)

	videoIDs := getValidVideoIDs(db, 100)

	executeQueries(db, videoIDs)
}

func createTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		video_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		parent_comment_id TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}
}

func insertComments(db *sql.DB) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&count)
	if err != nil {
		log.Fatal("Error counting rows:", err)
	}

	if count > 0 {
		fmt.Println("Database already has data, skipping insertion.")
		return
	}

	fmt.Println("Inserting 1 million comments...")

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Error starting transaction:", err)
	}

	stmt, err := tx.Prepare("INSERT INTO comments (id, content, video_id, user_id, parent_comment_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal("Error preparing insert statement:", err)
	}
	defer stmt.Close()

	for i := 1; i <= 1000000; i++ {
		id := fmt.Sprintf("comment-%d", i)
		content := fmt.Sprintf("This is comment number %d", i)
		videoID := fmt.Sprintf("video-%d", rand.Intn(10000))
		userID := fmt.Sprintf("user-%d", rand.Intn(50000))
		parentCommentID := getParentCommentID(i)
		timestamp := time.Now().Format(time.RFC3339)

		_, err := stmt.Exec(id, content, videoID, userID, parentCommentID, timestamp, timestamp)
		if err != nil {
			log.Fatal("Error inserting comment:", err)
		}

		if i%100000 == 0 {
			fmt.Printf("Inserted %d comments...\n", i)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal("Error committing transaction:", err)
	}

	fmt.Println("Insertion completed.")
}

func getParentCommentID(i int) *string {
	if i%10 == 0 {
		parentID := fmt.Sprintf("comment-%d", rand.Intn(i))
		return &parentID
	}
	return nil
}

func getValidVideoIDs(db *sql.DB, limit int) []string {
	rows, err := db.Query("SELECT DISTINCT video_id FROM comments LIMIT ?", limit)
	if err != nil {
		log.Fatal("Error fetching video IDs:", err)
	}
	defer rows.Close()

	var videoIDs []string
	for rows.Next() {
		var videoID string
		if err := rows.Scan(&videoID); err != nil {
			log.Fatal("Error scanning video ID:", err)
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs
}

func executeQueries(db *sql.DB, videoIDs []string) {
	query := `SELECT 
		c1.id, 
		c1.content, 
		c1.video_id, 
		c1.user_id, 
		c1.parent_comment_id,
		c1.created_at,  
		c1.updated_at,  
		COALESCE(
			json_group_array(
				json_object(
					'id', c2.id,
					'content', c2.content,
					'user_id', c2.user_id,
					'video_id', c2.video_id,
					'parent_comment_id', c2.parent_comment_id,
					'created_at', datetime(c2.created_at, 'unixepoch'), 
					'updated_at', datetime(c2.updated_at, 'unixepoch')   
				)
			) FILTER (WHERE c2.id IS NOT NULL), 
			'[]'
		) AS replies
	FROM comments c1
	LEFT JOIN comments c2 ON c1.id = c2.parent_comment_id
	WHERE c1.video_id = ?
	AND c1.parent_comment_id IS NULL
	GROUP BY c1.id
	ORDER BY c1.created_at DESC;`

	queryTimes := make([]time.Duration, 0)

	for _, videoID := range videoIDs {
		startTime := time.Now()

		rows, err := db.Query(query, videoID)
		if err != nil {
			log.Fatal("Error executing query:", err)
		}

		for rows.Next() {
			var id, content, videoID, userID, parentCommentID, createdAt, updatedAt, replies string
			if err := rows.Scan(&id, &content, &videoID, &userID, &parentCommentID, &createdAt, &updatedAt, &replies); err != nil {
				log.Fatal("Error scanning row:", err)
			}
		}
		rows.Close()

		elapsedTime := time.Since(startTime)
		queryTimes = append(queryTimes, elapsedTime)
		fmt.Printf("Query for video_id=%s took %s\n", videoID, elapsedTime)
	}

	sort.Slice(queryTimes, func(i, j int) bool {
		return queryTimes[i] > queryTimes[j]
	})

	fmt.Println("\nTop 5 slowest queries:")
	for i := 0; i < 5 && i < len(queryTimes); i++ {
		fmt.Printf("Rank %d: %s\n", i+1, queryTimes[i])
	}
}
