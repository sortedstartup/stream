package api

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	paymentProto "sortedstartup.com/stream/paymentservice/proto"
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

	// Update payment service usage for each user
	if len(userStorageUpdates) > 0 {
		s.log.Info("Updating payment service storage usage for users", "userCount", len(userStorageUpdates))

		var usersUpdated, usersFailed int
		for userID, storageBytes := range userStorageUpdates {
			// First, ensure user has a payment service record (defensive approach)
			// Check if user has subscription, if not initialize them
			userSubscription, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
				UserId: userID,
			})
			s.log.Info("userSubscription", "userSubscription", userSubscription, "err", err)
			if err != nil || (userSubscription != nil && !userSubscription.Success) {
				s.log.Info("User has no payment service record, initializing", "userID", userID)
				// Initialize payment service for this user
				initResp, err := s.paymentServiceClient.InitializeUser(ctx, &paymentProto.InitializeUserRequest{
					UserId: userID,
				})
				s.log.Info("initResp", "initResp", initResp)
				if err != nil {
					s.log.Error("Failed to initialize payment service for user during backfill",
						"userID", userID,
						"error", err)
					usersFailed++
					continue
				}
				if initResp != nil && !initResp.Success {
					s.log.Error("Payment service initialization failed for user during backfill",
						"userID", userID,
						"error", initResp.ErrorMessage)
					usersFailed++
					continue
				}
				s.log.Info("Payment service initialized for user during backfill", "userID", userID)
			}

			// Update payment service with the storage used by this user
			_, err = s.paymentServiceClient.UpdateUserUsage(ctx, &paymentProto.UpdateUserUsageRequest{
				UserId:      userID,
				UsageType:   "storage",
				UsageChange: storageBytes, // Add the storage bytes found during backfill
			})
			if err != nil {
				s.log.Error("Failed to update payment service storage for user",
					"userID", userID,
					"storageBytes", storageBytes,
					"error", err)
				usersFailed++
				continue
			}

			s.log.Info("Updated payment service storage for user",
				"userID", userID,
				"addedBytes", storageBytes)
			usersUpdated++
		}

		s.log.Info("Payment service storage updates completed",
			"usersUpdated", usersUpdated,
			"usersFailed", usersFailed)

		// Update user count for each unique user
		s.log.Info("Updating payment service user counts", "userCount", len(uniqueUsers))
		var userCountUpdated, userCountFailed int
		for userID := range uniqueUsers {
			// Calculate actual user count for this user across all their tenants
			userCount, err := s.calculateUserCountForOwner(ctx, userID)
			if err != nil {
				s.log.Error("Failed to calculate user count for owner",
					"userID", userID,
					"error", err)
				userCountFailed++
				continue
			}

			if userCount == 0 {
				s.log.Info("User has no tenants, skipping user count update", "userID", userID)
				continue
			}

			// Get current user count from payment service to calculate the difference
			currentUsage, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
				UserId: userID,
			})
			if err != nil || !currentUsage.Success {
				s.log.Error("Failed to get current usage for user count calculation", "userID", userID, "error", err)
				userCountFailed++
				continue
			}

			currentUserCount := int64(0)
			if currentUsage.SubscriptionInfo != nil && currentUsage.SubscriptionInfo.Usage != nil {
				currentUserCount = int64(currentUsage.SubscriptionInfo.Usage.UsersCount)
			}

			// Calculate the difference needed to reach the correct count
			userCountDifference := userCount - currentUserCount

			if userCountDifference == 0 {
				s.log.Info("User count already correct", "userID", userID, "count", userCount)
				continue
			}

			// Update payment service with the difference (can be positive or negative)
			_, err = s.paymentServiceClient.UpdateUserUsage(ctx, &paymentProto.UpdateUserUsageRequest{
				UserId:      userID,
				UsageType:   "users",
				UsageChange: userCountDifference, // Difference to reach correct count
			})
			if err != nil {
				s.log.Error("Failed to update payment service user count",
					"userID", userID,
					"userCount", userCount,
					"currentCount", currentUserCount,
					"difference", userCountDifference,
					"error", err)
				userCountFailed++
				continue
			}

			s.log.Info("Updated payment service user count for user",
				"userID", userID,
				"previousCount", currentUserCount,
				"newCount", userCount,
				"difference", userCountDifference)
			userCountUpdated++
		}

		s.log.Info("Payment service user count updates completed",
			"userCountUpdated", userCountUpdated,
			"userCountFailed", userCountFailed)
	}

	return nil
}

// calculateUserCountForOwner calculates total users across all tenants owned by a user
func (s *VideoAPI) calculateUserCountForOwner(ctx context.Context, ownerUserID string) (int64, error) {
	// Query to get total user count across all tenants owned by this user
	// This includes both personal and organizational tenants

	query := `
		SELECT COUNT(DISTINCT tu.user_id) as total_users
		FROM userservice_tenants t
		JOIN userservice_tenant_users tu ON t.id = tu.tenant_id
		WHERE t.created_by = ?
	`

	var totalUsers int64
	err := s.db.QueryRowContext(ctx, query, ownerUserID).Scan(&totalUsers)
	if err != nil {
		return 0, fmt.Errorf("failed to count users for owner %s: %w", ownerUserID, err)
	}

	s.log.Info("Calculated user count for owner",
		"ownerUserID", ownerUserID,
		"totalUsers", totalUsers)

	return totalUsers, nil
}
