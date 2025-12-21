package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Golangjobsuz/golangjobsuz/internal/commands"
	"github.com/Golangjobsuz/golangjobsuz/internal/search"
	"github.com/Golangjobsuz/golangjobsuz/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	storePath := "data/store.json"
	s, err := store.Load(storePath)
	if err != nil {
		log.Fatalf("load store: %v", err)
	}

	switch os.Args[1] {
	case "admin":
		adminCommand(s, os.Args[2:])
	case "search":
		searchCommand(s, os.Args[2:])
	case "profile":
		profileCommand(s, os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage: golangjobsuz <command> [flags]")
	fmt.Println("Commands:")
	fmt.Println("  admin   --action approve|ban --user <id> [--notes <text>] [--admin <id>]")
	fmt.Println("  search  --skills 'go,grpc' --location Tashkent --seniority mid --days 14 --page 1 --page-size 5")
	fmt.Println("  profile --id <profileID>")
}

func adminCommand(s *store.Store, args []string) {
	fs := flag.NewFlagSet("admin", flag.ExitOnError)
	action := fs.String("action", "", "approve or ban")
	userID := fs.String("user", "", "user id to target")
	notes := fs.String("notes", "", "audit notes")
	adminID := fs.String("admin", "admin", "admin user id")
	fs.Parse(args)

	if *action == "" || *userID == "" {
		fs.Usage()
		return
	}

	var err error
	switch strings.ToLower(*action) {
	case "approve":
		err = commands.ApproveRecruiter(s, *adminID, *userID, *notes)
	case "ban":
		err = commands.BanRecruiter(s, *adminID, *userID, *notes)
	default:
		fmt.Printf("unknown action %s\n", *action)
		return
	}
	if err != nil {
		log.Fatalf("admin command failed: %v", err)
	}
	fmt.Println(commands.AccessSummary(s, *userID))
}

func searchCommand(s *store.Store, args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	skills := fs.String("skills", "", "comma separated skills")
	location := fs.String("location", "", "location filter")
	seniority := fs.String("seniority", "", "seniority filter")
	days := fs.Int("days", 0, "maximum age in days")
	page := fs.Int("page", 1, "page number")
	pageSize := fs.Int("page-size", 5, "page size")
	fs.Parse(args)

	filter := search.Filters{
		Skills:     splitSkills(*skills),
		Location:   *location,
		Seniority:  *seniority,
		MaxAgeDays: *days,
		Page:       *page,
		PageSize:   *pageSize,
	}

	results := search.SearchProfiles(s.Profiles, filter)
	fmt.Printf("Found %d profiles (page %d/%d)\n", results.Total, results.CurrentPage, results.TotalPages)
	for _, r := range results.Results {
		fmt.Printf("- [%s] %s (%s, %s) skills=%s updated=%s\n",
			r.Profile.ID,
			r.Profile.Name,
			r.Profile.Seniority,
			r.Profile.Location,
			strings.Join(r.Profile.Skills, ", "),
			r.Profile.UpdatedAt.Format("2006-01-02"),
		)
		fmt.Printf("  Contact: email %s phone %s\n", r.RedactedEmail, r.RedactedPhone)
		fmt.Println("  CTA: Reply /request_contact", r.Profile.ID, "to ask the bot for direct contact access")
	}
}

func profileCommand(s *store.Store, args []string) {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	id := fs.String("id", "", "profile id")
	fs.Parse(args)

	if *id == "" {
		fs.Usage()
		return
	}
	p, ok := s.Profiles[*id]
	if !ok {
		fmt.Printf("profile %s not found\n", *id)
		return
	}
	res := search.RedactContact(p)
	fmt.Printf("Profile %s: %s\n", p.ID, p.Name)
	fmt.Printf("Location: %s | Seniority: %s\n", p.Location, p.Seniority)
	fmt.Printf("Skills: %s\n", strings.Join(p.Skills, ", "))
	fmt.Printf("Summary: %s\n", p.Summary)
	fmt.Printf("Updated: %s\n", p.UpdatedAt.Format("2006-01-02"))
	fmt.Printf("Contact email: %s | phone: %s\n", res.RedactedEmail, res.RedactedPhone)
	fmt.Println("CTA: Reply /request_contact", p.ID, "to ask the bot to share details with you")
}

func splitSkills(input string) []string {
	if input == "" {
		return nil
	}
	raw := strings.Split(input, ",")
	cleaned := make([]string, 0, len(raw))
	for _, s := range raw {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}
