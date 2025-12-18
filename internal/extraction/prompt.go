package extraction

import (
	"fmt"
	"strings"
)

// structuredPrompt describes the expected JSON schema for extraction.
func structuredPrompt(source string) string {
	schema := `{
  "name": "string",
  "contacts": {"email": "string", "phone": "string", "telegram": "string"},
  "location": "string",
  "skills": ["string", "string"],
  "experience_years": "number",
  "seniority": "string",
  "salary_expectation": "string",
  "links": ["string"],
  "summary": "string"
}`

	instructions := []string{
		"Extract a candidate profile as valid JSON only.",
		"Populate missing optional fields with null or empty collections as appropriate.",
		"Keep the response minimal and machine-readable without prose.",
		"Ensure numbers remain numbers and do not include units in numeric fields.",
	}

	return fmt.Sprintf("%s\nExpected schema:%s\n\nSource:\n%s", strings.Join(instructions, " "), schema, source)
}
