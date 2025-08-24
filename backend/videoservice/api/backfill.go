package api

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sortedstartup.com/stream/videoservice/proto"
)

// backfillFileSizes calculates and updates file sizes for existing videos that don't have this data
func (s *VideoAPI) backfillFileSizes() error {
	ctx := context.Background()

	// Check if there are any videos that need file size calculation
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM videoservice_videos 
		WHERE (file_size_bytes IS NULL OR file_size_bytes = 0) 
		AND is_deleted = FALSE
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check videos needing backfill: %w", err)
	}

	if count == 0 {
		s.log.Info("No videos need file size backfill")
		return nil
	}

	s.log.Info("Found videos needing file size backfill", "count", count)
	return s.processVideosBackfill(ctx, s.db)
}

// processVideosBackfill handles the actual backfill process for a given database
func (s *VideoAPI) processVideosBackfill(ctx context.Context, db *sql.DB) error {
	// Start a transaction to prevent SQLite busy errors
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() succeeds

	// Get videos that need file size calculation, grouped by user
	rows, err := tx.QueryContext(ctx, `
		SELECT id, url, uploaded_user_id FROM videoservice_videos 
		WHERE (file_size_bytes IS NULL OR file_size_bytes = 0) 
		AND is_deleted = FALSE
		ORDER BY uploaded_user_id
	`)
	if err != nil {
		return fmt.Errorf("failed to get videos for backfill: %w", err)
	}
	defer rows.Close()

	var updated, skipped, failed int
	videoDir := s.getVideoDir()

	// Track storage usage per user for payment service updates
	userStorageUpdates := make(map[string]int64)

	// Track unique users for user count updates
	uniqueUsers := make(map[string]bool)

	for rows.Next() {
		var videoID, fileName, userID string
		if err := rows.Scan(&videoID, &fileName, &userID); err != nil {
			s.log.Error("Error scanning video row", "error", err)
			failed++
			continue
		}

		s.log.Info("videoDir", "videoDir", videoDir)
		// Construct full file path
		fullPath := filepath.Join(videoDir, fileName)

		// Get file info
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			s.log.Warn("Video file not found during backfill", "videoID", videoID, "path", fullPath, "error", err)
			skipped++
			continue
		}

		fileSize := fileInfo.Size()
		s.log.Debug("Backfilling file size", "videoID", videoID, "size", fileSize, "file", fileName, "userID", userID)

		// Update video file size in the database
		_, err = tx.ExecContext(ctx, `
			UPDATE videoservice_videos 
			SET file_size_bytes = ?, updated_at = ? 
			WHERE id = ?
		`, fileSize, time.Now(), videoID)
		if err != nil {
			s.log.Error("Failed to update video file size", "videoID", videoID, "error", err)
			failed++
			continue
		}

		// Track storage usage for this user
		userStorageUpdates[userID] += fileSize
		uniqueUsers[userID] = true
		updated++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating video rows: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.log.Info("File size backfill summary",
		"updated", updated,
		"skipped", skipped,
		"failed", failed,
		"usersAffected", len(userStorageUpdates))

	// Update VideoService storage usage for each user
	if len(userStorageUpdates) > 0 {
		s.log.Info("Updating VideoService storage usage for users", "userCount", len(userStorageUpdates))

		var usersUpdated, usersFailed int
		for userID, storageBytes := range userStorageUpdates {
			// Update storage usage in videoservice_user_storage_usage table
			s.log.Info("Updating storage usage for user during backfill",
				"userID", userID,
				"addedBytes", storageBytes)

			// Use UpdateStorageUsage method to set the total storage usage
			_, err := s.UpdateStorageUsage(ctx, &proto.UpdateStorageUsageRequest{
				UserId:      userID,
				UsageChange: storageBytes, // This sets the total usage
			})
			if err != nil {
				s.log.Error("Failed to update storage usage for user during backfill",
					"userID", userID,
					"storageBytes", storageBytes,
					"error", err)
				usersFailed++
				continue
			}

			s.log.Info("Storage usage updated successfully for user during backfill",
				"userID", userID,
				"storageBytes", storageBytes)
			usersUpdated++
		}

		s.log.Info("VideoService storage updates completed",
			"usersUpdated", usersUpdated,
			"usersFailed", usersFailed)

		// Note: User count backfill is now handled by UserService, not VideoService
		s.log.Info("User count backfill is now handled by UserService", "uniqueUsersFound", len(uniqueUsers))
	}

	return nil
}
