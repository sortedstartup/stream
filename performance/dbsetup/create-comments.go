package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open SQLite database
	db, err := sql.Open("sqlite3", "comments.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		video_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		parent_comment_id TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Insert 10M records (100 videos, 1000 comments, 100 replies per comment)
	numVideos := 100
	commentsPerVideo := 1000 // 1000 comments per video
	repliesPerComment := 100 // 100 replies per comment

	start := time.Now()
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Error starting transaction:", err)
	}

	stmt, err := tx.Prepare("INSERT INTO comments (id, content, video_id, user_id, parent_comment_id) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal("Error preparing statement:", err)
	}
	defer stmt.Close()

	count := 0
	for v := 1; v <= numVideos; v++ {
		videoID := fmt.Sprintf("video_%d", v)

		for c := 1; c <= commentsPerVideo; c++ {
			commentID := fmt.Sprintf("comment_%d_%d", v, c)
			_, err := stmt.Exec(commentID, "This is a comment", videoID, "user_1", nil)
			if err != nil {
				log.Fatal("Error inserting comment:", err)
			}
			count++

			// Insert replies
			for r := 1; r <= repliesPerComment; r++ {
				replyID := fmt.Sprintf("reply_%d_%d_%d", v, c, r)
				_, err := stmt.Exec(replyID, "This is a reply", videoID, "user_2", commentID)
				if err != nil {
					log.Fatal("Error inserting reply:", err)
				}
				count++
			}

			// Commit batch every 10,000 records
			if count%10000 == 0 {
				err := tx.Commit()
				if err != nil {
					log.Fatal("Error committing transaction:", err)
				}
				tx, _ = db.Begin()
				stmt, _ = tx.Prepare("INSERT INTO comments (id, content, video_id, user_id, parent_comment_id) VALUES (?, ?, ?, ?, ?)")
				fmt.Println("Inserted", count, "records...")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal("Error committing final transaction:", err)
	}

	fmt.Println("Bulk insert completed!")
	fmt.Println("Data inserted in", time.Since(start))

	// Index creation
	fmt.Println("Creating indexes for faster queries...")
	indexStart := time.Now()

	indexes := []string{
		"CREATE INDEX idx_video_id ON comments(video_id);",
		"CREATE INDEX idx_user_id ON comments(user_id);",
		"CREATE INDEX idx_parent_comment ON comments(parent_comment_id);",
	}

	for _, query := range indexes {
		fmt.Println("Creating index:", query)
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal("Error creating index:", err)
		}
	}

	fmt.Println("Indexes created in", time.Since(indexStart))

	// Measure query performance
	benchmarkQueries(db)
}

// Benchmark queries with and without indexes
func benchmarkQueries(db *sql.DB) {

	// Define test video ID
	videoID := "video_50"

	// SQL query to fetch comments with nested replies
	query := `
		SELECT 
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
		ORDER BY c1.created_at DESC;
	`

	// Run query and measure time
	start := time.Now()
	rows, err := db.Query(query, videoID)
	if err != nil {
		log.Fatal("Error running query:", err)
	}
	defer rows.Close()

	// Fetch results (simulating actual usage)
	var count int
	for rows.Next() {
		count++
	}
	fmt.Printf("Query executed in %v | Rows fetched: %d\n", time.Since(start), count)
}
