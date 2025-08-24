package api

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sortedstartup.com/stream/userservice/db"
)

// backfillUserCounts backfills user count data for existing users
func (s *UserAPI) backfillUserCounts(ctx context.Context) error {
	s.log.Info("Starting user count backfill")

	// Get all unique tenant owners
	query := `SELECT DISTINCT created_by FROM userservice_tenants WHERE created_by IS NOT NULL`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get tenant owners: %w", err)
	}
	defer rows.Close()

	var ownerIDs []string
	for rows.Next() {
		var ownerID string
		if err := rows.Scan(&ownerID); err != nil {
			s.log.Error("Failed to scan owner ID", "error", err)
			continue
		}
		ownerIDs = append(ownerIDs, ownerID)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating owner rows: %w", err)
	}

	s.log.Info("Found tenant owners for backfill", "ownerCount", len(ownerIDs))

	var updated, failed int
	for _, ownerID := range ownerIDs {
		// Calculate actual user count for this owner
		userCount, err := s.calculateUserCountForOwner(ctx, ownerID)
		if err != nil {
			s.log.Error("Failed to calculate user count for owner", "ownerID", ownerID, "error", err)
			failed++
			continue
		}

		// Update user limits table
		now := time.Now().Unix()
		err = s.dbQueries.UpsertUserLimits(ctx, db.UpsertUserLimitsParams{
			UserID:           ownerID,
			UsersCount:       sql.NullInt64{Int64: userCount, Valid: true},
			LastCalculatedAt: sql.NullInt64{Int64: now, Valid: true},
			CreatedAt:        now,
			UpdatedAt:        now,
		})
		if err != nil {
			s.log.Error("Failed to update user count for owner", "ownerID", ownerID, "userCount", userCount, "error", err)
			failed++
			continue
		}

		s.log.Info("Updated user count for owner", "ownerID", ownerID, "userCount", userCount)
		updated++
	}

	s.log.Info("User count backfill completed", "updated", updated, "failed", failed)
	return nil
}

// calculateUserCountForOwner calculates total users across all tenants owned by a user
func (s *UserAPI) calculateUserCountForOwner(ctx context.Context, ownerUserID string) (int64, error) {
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
