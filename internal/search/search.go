package search

import (
	"slices"
	"strings"
	"time"

	"golangjobsuz/internal/store"
)

// Filters used for profile search.
type Filters struct {
	Skills     []string
	Location   string
	Seniority  string
	MaxAgeDays int
	Page       int
	PageSize   int
}

// Result represents a single profile result with redacted contact info.
type Result struct {
	Profile       store.Profile
	RedactedEmail string
	RedactedPhone string
}

// PaginatedResults contains the filtered profiles with pagination metadata.
type PaginatedResults struct {
	Results     []Result
	Total       int
	TotalPages  int
	CurrentPage int
	PageSize    int
}

// RedactContact returns a result with redacted contact data for a single profile.
func RedactContact(p store.Profile) Result {
	return Result{
		Profile:       p,
		RedactedEmail: redactEmail(p.ContactEmail),
		RedactedPhone: redactPhone(p.ContactPhone),
	}
}

// SearchProfiles filters profiles based on the provided filters and paginates the output.
func SearchProfiles(profiles map[string]store.Profile, filters Filters) PaginatedResults {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 5
	}

	lowerSkills := make([]string, 0, len(filters.Skills))
	for _, skill := range filters.Skills {
		trimmed := strings.TrimSpace(skill)
		if trimmed != "" {
			lowerSkills = append(lowerSkills, strings.ToLower(trimmed))
		}
	}
	filters.Skills = lowerSkills

	candidates := make([]Result, 0)
	for _, p := range profiles {
		if filters.Location != "" && !strings.EqualFold(filters.Location, p.Location) {
			continue
		}
		if filters.Seniority != "" && !strings.EqualFold(filters.Seniority, p.Seniority) {
			continue
		}
		if filters.MaxAgeDays > 0 {
			cutoff := time.Now().AddDate(0, 0, -filters.MaxAgeDays)
			if p.UpdatedAt.Before(cutoff) {
				continue
			}
		}

		if len(filters.Skills) > 0 {
			if !skillsMatch(p.Skills, filters.Skills) {
				continue
			}
		}

		candidates = append(candidates, Result{
			Profile:       p,
			RedactedEmail: redactEmail(p.ContactEmail),
			RedactedPhone: redactPhone(p.ContactPhone),
		})
	}

	slices.SortFunc(candidates, func(a, b Result) int {
		if a.Profile.UpdatedAt.Equal(b.Profile.UpdatedAt) {
			return strings.Compare(a.Profile.ID, b.Profile.ID)
		}
		if a.Profile.UpdatedAt.After(b.Profile.UpdatedAt) {
			return -1
		}
		return 1
	})

	total := len(candidates)
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	start := (filters.Page - 1) * filters.PageSize
	if start >= total {
		return PaginatedResults{Results: []Result{}, Total: total, TotalPages: totalPages, CurrentPage: filters.Page, PageSize: filters.PageSize}
	}
	end := start + filters.PageSize
	if end > total {
		end = total
	}

	return PaginatedResults{
		Results:     candidates[start:end],
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: filters.Page,
		PageSize:    filters.PageSize,
	}
}

func skillsMatch(profileSkills, required []string) bool {
	if len(required) == 0 {
		return true
	}
	lowerProfile := make([]string, len(profileSkills))
	for i, s := range profileSkills {
		lowerProfile[i] = strings.ToLower(s)
	}
	for _, req := range required {
		match := false
		for _, skill := range lowerProfile {
			if strings.Contains(skill, req) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func redactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "hidden"
	}
	local := parts[0]
	if len(local) > 2 {
		local = local[:2] + strings.Repeat("*", len(parts[0])-2)
	} else {
		local = strings.Repeat("*", len(parts[0]))
	}
	return local + "@" + parts[1]
}

func redactPhone(phone string) string {
	digits := []rune(phone)
	if len(digits) <= 4 {
		return strings.Repeat("*", len(digits))
	}
	masked := strings.Repeat("*", len(digits)-4) + string(digits[len(digits)-4:])
	return masked
}
