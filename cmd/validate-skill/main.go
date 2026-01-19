package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/victorzhuk/go-ent/internal/skill"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: validate-skill <skill-file>")
	}

	parser := skill.NewParser()
	validator := skill.NewValidator()
	scorer := skill.NewQualityScorer()

	meta, err := parser.ParseSkillFile(os.Args[1])
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Read error: %v", err)
	}

	qualityScore := scorer.Score(meta, string(content))
	meta.QualityScore = qualityScore

	result := validator.Validate(meta, string(content))

	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║                    SKILL QUALITY REPORT                    ║")
	fmt.Println("╠════════════════════════════════════════════════════╣")
	fmt.Printf("║ Total Score:     %6.2f / 100                          ║\n", qualityScore.Total)
	fmt.Println("╠════════════════════════════════════════════════════╣")

	fmt.Println("║ Breakdown by Category:                                    ║")
	fmt.Println("╠════════════════════════════════════════════════════╣")

	printCategoryBar("Structure", qualityScore.Structure.Total, 20)
	printCategoryBar("Content", qualityScore.Content.Total, 25)
	printCategoryBar("Examples", qualityScore.Examples.Total, 25)
	printCategoryBar("Triggers", qualityScore.Triggers, 15)
	printCategoryBar("Conciseness", qualityScore.Conciseness, 15)

	fmt.Println("╠════════════════════════════════════════════════════╣")
	fmt.Printf("║ Valid:           %6v                                       ║\n", result.Valid)
	fmt.Printf("║ Errors:           %6d                                       ║\n", result.ErrorCount())
	fmt.Printf("║ Warnings:        %6d                                       ║\n", result.WarningCount())
	fmt.Printf("║ Info:            %6d                                       ║\n", len(result.Issues)-result.ErrorCount()-result.WarningCount())
	fmt.Println("╚════════════════════════════════════════════════════╝")

	if qualityScore.Total < 60 {
		fmt.Println("\n⚠️  LOW QUALITY SCORE - RECOMMENDATIONS:")
		printRecommendations(qualityScore)
	}

	if len(result.Issues) > 0 {
		fmt.Println("\n📋 Validation Issues:")
		for _, issue := range result.Issues {
			fmt.Printf("  %s\n", issue)
		}
	}

	resultStrict := validator.ValidateStrict(meta, string(content))
	fmt.Printf("\nStrict Mode Valid: %v\n", resultStrict.Valid)
	if len(resultStrict.Issues) > 0 {
		fmt.Println("Strict Mode Issues:")
		for _, issue := range resultStrict.Issues {
			fmt.Printf("  %s\n", issue)
		}
	}
}

func printCategoryBar(category string, score, max float64) {
	percentage := (score / max) * 100

	barLength := int(percentage / 10)
	bar := strings.Repeat("█", barLength) + strings.Repeat("░", 10-barLength)

	fmt.Printf("║ %12s:   %5.2f / %-4.0f [%s] %5.0f%%    ║\n",
		category, score, max, bar, percentage)
}

func printRecommendations(score *skill.QualityScore) {
	if score.Structure.Total < 10 {
		fmt.Println("  • Add missing XML sections (role, instructions, constraints, examples, output_format, edge_cases)")
	}
	if score.Content.Total < 15 {
		fmt.Println("  • Improve content quality: clarify role, add actionable instructions, specific constraints")
	}
	if score.Examples.Total < 15 {
		fmt.Println("  • Add more examples (3-5 diverse examples with edge cases)")
	}
	if score.Triggers < 10 {
		fmt.Println("  • Add explicit triggers with weights for better matching")
	}
	if score.Conciseness < 10 {
		fmt.Println("  • Reduce content length to <5000 tokens (move details to references/)")
	}
}
