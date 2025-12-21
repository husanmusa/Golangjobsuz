package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Golangjobsuz/golangjobsuz/internal/store"
)

// ApproveRecruiter marks a user as an approved recruiter and updates their role.
func ApproveRecruiter(s *store.Store, adminID, userID, notes string) error {
	if s == nil {
		return errors.New("store is nil")
	}
	s.EnsureUserRole(userID, "recruiter")
	s.RecruiterAccess[userID] = store.RecruiterAccess{
		UserID:    userID,
		Status:    "approved",
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: adminID,
		Notes:     notes,
	}
	return s.Save()
}

// BanRecruiter revokes recruiter access.
func BanRecruiter(s *store.Store, adminID, userID, notes string) error {
	if s == nil {
		return errors.New("store is nil")
	}
	s.EnsureUserRole(userID, "recruiter_banned")
	s.RecruiterAccess[userID] = store.RecruiterAccess{
		UserID:    userID,
		Status:    "banned",
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: adminID,
		Notes:     notes,
	}
	return s.Save()
}

// AccessSummary returns a human readable string for recruiter access status.
func AccessSummary(s *store.Store, userID string) string {
	ra, ok := s.RecruiterAccess[userID]
	if !ok {
		return fmt.Sprintf("User %s has pending access", userID)
	}
	return fmt.Sprintf("User %s is %s (updated by %s at %s)", userID, strings.ToUpper(ra.Status), ra.UpdatedBy, ra.UpdatedAt.Format(time.RFC822))
}
