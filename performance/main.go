package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const numExecutions = 100

// Define Prometheus metrics
var queryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "query_execution_time_seconds",
		Help:    "Time taken to execute queries",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"index"},
)

func main() {
	// Register Prometheus metrics
	prometheus.MustRegister(queryDuration)

	// Start HTTP server for Prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("🚀 Prometheus metrics available at :8080/metrics")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Open SQLite database
	db, err := sql.Open("sqlite3", "comments.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Fetch 100 unique video IDs
	videoIDs := fetchVideoIDs(db, numExecutions)
	if len(videoIDs) < numExecutions {
		log.Fatalf("Not enough unique video IDs in the database! Found: %d\n", len(videoIDs))
	}

	// Drop indexes for testing without indexes
	fmt.Println("\n🔴 Dropping indexes (testing without indexes)...")
	dropIndexes(db)

	// Run performance test without indexes
	fmt.Println("\n🚀 Running queries WITHOUT indexes...")
	benchmarkQueryPerformance(db, videoIDs, "without_index")

	// Create indexes for testing with indexes
	fmt.Println("\n🟢 Creating indexes (testing with indexes)...")
	createIndexes(db)

	// Run performance test with indexes
	fmt.Println("\n🚀 Running queries WITH indexes...")
	benchmarkQueryPerformance(db, videoIDs, "with_index")
}

// Function to fetch 100 unique video IDs
func fetchVideoIDs(db *sql.DB, limit int) []string {
	query := `SELECT DISTINCT video_id FROM comments LIMIT ?`
	rows, err := db.Query(query, limit)
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

// Function to drop indexes
func dropIndexes(db *sql.DB) {
	_, _ = db.Exec("DROP INDEX IF EXISTS idx_video_id;")
	_, _ = db.Exec("DROP INDEX IF EXISTS idx_user_id;")
	_, _ = db.Exec("DROP INDEX IF EXISTS idx_parent_comment;")
}

// Function to create indexes
func createIndexes(db *sql.DB) {
	_, _ = db.Exec("CREATE INDEX idx_video_id ON comments(video_id);")
	_, _ = db.Exec("CREATE INDEX idx_user_id ON comments(user_id);")
	_, _ = db.Exec("CREATE INDEX idx_parent_comment ON comments(parent_comment_id);")
}

// Function to benchmark query performance and expose Prometheus metrics
func benchmarkQueryPerformance(db *sql.DB, videoIDs []string, indexLabel string) {
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

	executionTimes := make([]time.Duration, len(videoIDs))

	for i, videoID := range videoIDs {
		start := time.Now()

		rows, err := db.Query(query, videoID)
		if err != nil {
			log.Fatal("Error running query:", err)
		}
		defer rows.Close()

		// Count rows (simulating real query execution)
		count := 0
		for rows.Next() {
			count++
		}

		duration := time.Since(start)
		executionTimes[i] = duration
		queryDuration.WithLabelValues(indexLabel).Observe(duration.Seconds())

		fmt.Printf("Run %d | Video ID: %s | Time: %v | Rows fetched: %d\n", i+1, videoID, duration, count)
	}

	// Sort execution times
	sort.Slice(executionTimes, func(i, j int) bool {
		return executionTimes[i] < executionTimes[j]
	})

	// Display results
	fmt.Println("\n🔝 Top 5 Fastest Executions:")
	for i := 0; i < 5 && i < len(executionTimes); i++ {
		fmt.Printf("%d. %v\n", i+1, executionTimes[i])
	}

	fmt.Println("\n🐌 Top 5 Slowest Executions:")
	for i := len(executionTimes) - 5; i < len(executionTimes) && i >= 0; i++ {
		fmt.Printf("%d. %v\n", len(executionTimes)-i, executionTimes[i])
	}
}
