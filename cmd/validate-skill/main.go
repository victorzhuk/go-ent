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

	meta, err := parser.ParseSkillFile(os.Args[1])
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Read error: %v", err)
	}

	result := validator.Validate(meta, string(content))

	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Println("║                  SKILL VALIDATION REPORT                  ║")
	fmt.Println("╠════════════════════════════════════════════════╣")

	fmt.Printf("║ Valid:           %6v                                       ║\n", result.Valid)
	fmt.Printf("║ Errors:           %6d                                       ║\n", result.ErrorCount())
	fmt.Printf("║ Warnings:        %6d                                       ║\n", result.WarningCount())
	fmt.Println("╚══════════════════════════════════════════════════╝")

	if len(result.Issues) > 0 {
		fmt.Println("\n📋 Validation Issues:")
		for _, issue := range result.Issues {
			fmt.Printf("  %s\n", issue.String())
		}
	}

	resultStrict := validator.ValidateStrict(meta, string(content))
	fmt.Printf("\nStrict Mode Valid: %v\n", resultStrict.Valid)
	if len(resultStrict.Issues) > 0 {
		fmt.Println("Strict Mode Issues:")
		for _, issue := range resultStrict.Issues {
			fmt.Printf("  %s\n", issue.String())
		}
	}
}
