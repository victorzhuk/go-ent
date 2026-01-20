package skill

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/victorzhuk/go-ent/internal/skill"
)

type analysisResult struct {
	Name        string
	Description string
	TotalScore  float64
	Structure   float64
	Content     float64
	Examples    float64
	Triggers    float64
	Conciseness float64
	Category    string
}

type distributionReport struct {
	Total    int
	Average  float64
	MinScore float64
	MaxScore float64
	Pass     int
	Improve  int
	Fail     int
}

type commonIssue struct {
	Category  string
	Threshold float64
	Count     int
	Skills    []string
}

func newAnalyzeCmd() *cobra.Command {
	var (
		allFlag  bool
		jsonFlag bool
		csvFlag  bool
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze skill quality and generate reports",
		Long: `Analyze all skills in the registry and generate quality reports.

Evaluates skills across multiple dimensions:
  • Structure: Section presence and completeness (max 20)
  • Content: Role clarity, instruction quality (max 25)
  • Examples: Count, diversity, format (max 25)
  • Triggers: Presence and quality (max 15)
  • Conciseness: Token count penalty (max 15)

Produces:
  • Overall quality distribution
  • Pass/Improve/Fail classification
  • Common issues per category
  • Export to JSON or CSV format

Classification:
  • Pass: ≥80 points
  • Improve: 60-79 points
  • Fail: <60 points

Examples:
  # Analyze all skills with console output
  ent skill analyze --all

  # Export results to JSON
  ent skill analyze --all --json

  # Export results to CSV
  ent skill analyze --all --csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsPath := getSkillsPath()
			registry := skill.NewRegistry()

			if err := registry.Load(skillsPath); err != nil {
				return fmt.Errorf("load skills: %w", err)
			}

			skills := registry.All()
			if len(skills) == 0 {
				_, _ = fmt.Fprintln(os.Stderr, "No skills found")
				return nil
			}

			scorer := skill.NewQualityScorer()

			results, err := analyzeSkills(skills, scorer)
			if err != nil {
				return fmt.Errorf("analyze skills: %w", err)
			}

			distribution := calculateDistribution(results)
			issues := identifyCommonIssues(results)

			switch {
			case jsonFlag:
				return exportJSON(results, distribution, issues)
			case csvFlag:
				return exportCSV(results)
			default:
				return printReport(results, distribution, issues)
			}
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", true, "Analyze all skills (default)")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Export results to JSON format")
	cmd.Flags().BoolVar(&csvFlag, "csv", false, "Export results to CSV format")

	return cmd
}

func analyzeSkills(skills []skill.SkillMeta, scorer *skill.QualityScorer) ([]analysisResult, error) {
	results := make([]analysisResult, 0, len(skills))

	for _, s := range skills {
		content, err := os.ReadFile(s.FilePath)
		if err != nil {
			continue
		}

		score := scorer.Score(&s, string(content))

		category := "Pass"
		if score.Total < 60 {
			category = "Fail"
		} else if score.Total < 80 {
			category = "Improve"
		}

		result := analysisResult{
			Name:        s.Name,
			Description: s.Description,
			TotalScore:  score.Total,
			Structure:   score.Structure.Total,
			Content:     score.Content.Total,
			Examples:    score.Examples.Total,
			Triggers:    score.Triggers,
			Conciseness: score.Conciseness,
			Category:    category,
		}

		results = append(results, result)
	}

	return results, nil
}

func calculateDistribution(results []analysisResult) distributionReport {
	if len(results) == 0 {
		return distributionReport{}
	}

	report := distributionReport{
		Total:    len(results),
		MinScore: results[0].TotalScore,
		MaxScore: results[0].TotalScore,
	}

	sum := 0.0
	passCount := 0
	improveCount := 0
	failCount := 0

	for _, r := range results {
		sum += r.TotalScore

		if r.TotalScore < report.MinScore {
			report.MinScore = r.TotalScore
		}
		if r.TotalScore > report.MaxScore {
			report.MaxScore = r.TotalScore
		}

		switch r.Category {
		case "Pass":
			passCount++
		case "Improve":
			improveCount++
		case "Fail":
			failCount++
		}
	}

	report.Average = sum / float64(len(results))
	report.Pass = passCount
	report.Improve = improveCount
	report.Fail = failCount

	return report
}

func identifyCommonIssues(results []analysisResult) []commonIssue {
	issues := []commonIssue{
		{
			Category:  "Structure",
			Threshold: 12.0,
		},
		{
			Category:  "Content",
			Threshold: 15.0,
		},
		{
			Category:  "Examples",
			Threshold: 15.0,
		},
		{
			Category:  "Triggers",
			Threshold: 9.0,
		},
		{
			Category:  "Conciseness",
			Threshold: 9.0,
		},
	}

	for i := range issues {
		for _, r := range results {
			var score float64
			switch issues[i].Category {
			case "Structure":
				score = r.Structure
			case "Content":
				score = r.Content
			case "Examples":
				score = r.Examples
			case "Triggers":
				score = r.Triggers
			case "Conciseness":
				score = r.Conciseness
			}

			if score < issues[i].Threshold {
				issues[i].Count++
				issues[i].Skills = append(issues[i].Skills, r.Name)
			}
		}
	}

	for i := range issues {
		sort.Strings(issues[i].Skills)
	}

	return issues
}

func printReport(results []analysisResult, distribution distributionReport, issues []commonIssue) error {
	fmt.Println("Skill Quality Analysis Report")
	fmt.Println("==============================")
	fmt.Println()

	printDistribution(distribution)
	fmt.Println()

	printCommonIssues(issues)
	fmt.Println()

	printSkillDetails(results)

	return nil
}

func printDistribution(d distributionReport) {
	fmt.Println("Distribution")
	fmt.Println("------------")
	fmt.Printf("Total Skills:     %d\n", d.Total)
	fmt.Printf("Average Score:    %.1f\n", d.Average)
	fmt.Printf("Score Range:      %.1f - %.1f\n", d.MinScore, d.MaxScore)
	fmt.Println()
	fmt.Printf("Pass (≥80):       %d\n", d.Pass)
	fmt.Printf("Improve (60-79):  %d\n", d.Improve)
	fmt.Printf("Fail (<60):       %d\n", d.Fail)
}

func printCommonIssues(issues []commonIssue) {
	fmt.Println("Common Issues")
	fmt.Println("-------------")

	for _, issue := range issues {
		if issue.Count == 0 {
			continue
		}

		fmt.Printf("\n%s (< %.1f): %d skills affected\n", issue.Category, issue.Threshold, issue.Count)

		if len(issue.Skills) > 0 {
			skillsStr := strings.Join(issue.Skills, ", ")
			if len(skillsStr) > 80 {
				fmt.Printf("  %s...\n", skillsStr[:80])
			} else {
				fmt.Printf("  %s\n", skillsStr)
			}
		}
	}
}

func printSkillDetails(results []analysisResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	fmt.Println()
	fmt.Println("Skill Details")
	fmt.Println("-------------")
	fmt.Printf("%-20s %-8s %-8s %-8s %-8s %-8s %-8s %s\n",
		"NAME", "TOTAL", "STRUCT", "CONTENT", "EXAM", "TRIG", "CONCISE", "CATEGORY")
	fmt.Println(strings.Repeat("-", 100))

	for _, r := range results {
		desc := r.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}

		fmt.Printf("%-20s %-8.1f %-8.1f %-8.1f %-8.1f %-8.1f %-8.1f %s\n",
			r.Name, r.TotalScore, r.Structure, r.Content,
			r.Examples, r.Triggers, r.Conciseness, desc)
	}
}

func exportJSON(results []analysisResult, distribution distributionReport, issues []commonIssue) error {
	output := map[string]interface{}{
		"distribution": distribution,
		"issues":       issues,
		"skills":       results,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func exportCSV(results []analysisResult) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	headers := []string{
		"Name", "TotalScore", "Structure", "Content", "Examples",
		"Triggers", "Conciseness", "Category",
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("write headers: %w", err)
	}

	for _, r := range results {
		row := []string{
			r.Name,
			fmt.Sprintf("%.1f", r.TotalScore),
			fmt.Sprintf("%.1f", r.Structure),
			fmt.Sprintf("%.1f", r.Content),
			fmt.Sprintf("%.1f", r.Examples),
			fmt.Sprintf("%.1f", r.Triggers),
			fmt.Sprintf("%.1f", r.Conciseness),
			r.Category,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	return nil
}
