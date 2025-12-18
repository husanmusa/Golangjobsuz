package extraction

import (
	"time"
)

// CandidateProfile captures structured details extracted from unstructured candidate data.
type CandidateProfile struct {
	Name              string            `json:"name" validate:"required"`
	Contacts          map[string]string `json:"contacts,omitempty"`
	Location          string            `json:"location,omitempty"`
	Skills            []string          `json:"skills" validate:"required,dive,required"`
	ExperienceYears   float64           `json:"experience_years" validate:"gte=0"`
	Seniority         string            `json:"seniority,omitempty"`
	SalaryExpectation string            `json:"salary_expectation,omitempty"`
	Links             []string          `json:"links,omitempty"`
	Summary           string            `json:"summary,omitempty"`
}

// Draft stores an extracted profile together with traceable AI metadata.
type Draft struct {
	Profile     CandidateProfile `json:"profile"`
	RawResponse string           `json:"raw_response"`
	Model       string           `json:"model"`
	ExtractedAt time.Time        `json:"extracted_at"`
}
