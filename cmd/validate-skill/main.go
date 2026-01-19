package main

import (
	"fmt"
	"log"
	"os"

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

	// Calculate quality score
	qualityScore := scorer.Score(meta, string(content))
	meta.QualityScore = qualityScore

	result := validator.Validate(meta, string(content))

	fmt.Printf("Valid: %v\n", result.Valid)
	fmt.Printf("Quality Score: %.2f\n", qualityScore)
	fmt.Printf("Errors: %d, Warnings: %d, Info: %d\n",
		result.ErrorCount(), result.WarningCount(), len(result.Issues)-result.ErrorCount()-result.WarningCount())

	if len(result.Issues) > 0 {
		fmt.Println("\nIssues:")
		for _, issue := range result.Issues {
			fmt.Printf("  %s\n", issue)
		}
	}

	// Test strict mode
	resultStrict := validator.ValidateStrict(meta, string(content))
	fmt.Printf("\nStrict Mode Valid: %v\n", resultStrict.Valid)
	if len(resultStrict.Issues) > 0 {
		fmt.Println("\nStrict Mode Issues:")
		for _, issue := range resultStrict.Issues {
			fmt.Printf("  %s\n", issue)
		}
	}
}
