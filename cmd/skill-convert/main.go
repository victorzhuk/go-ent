package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Input     string
	Output    string
	All       bool
	SkillsDir string
	DryRun    bool
	Backup    bool
	Validate  bool
	Verbose   bool
}

type V3Skill struct {
	Name         string
	Description  string
	Version      string
	Author       string
	License      string
	Tags         []string
	Triggers     V3Triggers
	Category     string
	QualityScore int

	Role         string
	Instructions string
	Constraints  string
	EdgeCases    string
	Examples     string
	OutputFormat string
}

type V3Triggers struct {
	Keywords    []string
	FilePattern string
	Weight      float64
}

type V4Skill struct {
	Name         string
	Description  string
	Triggers     []string
	Role         string
	Instructions string
	Examples     string
	References   []Reference
}

type Reference struct {
	Path    string
	Title   string
	Content string
}

type FileReport struct {
	SourcePath  string
	OutputPath  string
	Status      string
	VersionFrom string
	Changes     []string
	Error       error
}

type MigrationReport struct {
	TotalFiles int
	Successful int
	Failed     int
	Skipped    int
	Details    []FileReport
}

func (r *MigrationReport) Print() {
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Println("║           MIGRATION REPORT                       ║")
	fmt.Println("╠════════════════════════════════════════════════╣")
	fmt.Printf("║ Total:      %6d                                       ║\n", r.TotalFiles)
	fmt.Printf("║ Successful: %6d                                       ║\n", r.Successful)
	fmt.Printf("║ Failed:     %6d                                       ║\n", r.Failed)
	fmt.Printf("║ Skipped:    %6d                                       ║\n", r.Skipped)
	fmt.Println("╚════════════════════════════════════════════════╝")

	if r.Failed > 0 {
		fmt.Println("\n❌ Failed Files:")
		for _, d := range r.Details {
			if d.Status == "failed" {
				fmt.Printf("  - %s: %v\n", d.SourcePath, d.Error)
			}
		}
	}

	if r.Skipped > 0 {
		fmt.Println("\n⏭️  Skipped Files:")
		for _, d := range r.Details {
			if d.Status == "skipped" {
				fmt.Printf("  - %s (reason: %s)\n", d.SourcePath, d.VersionFrom)
			}
		}
	}

	if r.Successful > 0 {
		fmt.Println("\n✅ Successfully Converted:")
		for _, d := range r.Details {
			if d.Status == "success" {
				fmt.Printf("  - %s → %s\n", d.SourcePath, d.OutputPath)
				if len(d.Changes) > 0 {
					for _, c := range d.Changes {
						fmt.Printf("    ✓ %s\n", c)
					}
				}
			}
		}
	}
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		logFatal("parse flags: %v", err)
	}

	if cfg.All {
		if cfg.SkillsDir == "" {
			cfg.SkillsDir = "./pkg/skills"
		}
		report := migrateAll(cfg)
		report.Print()
		os.Exit(0)
	}

	if cfg.Input == "" {
		logFatal("error: --input required (use --all to convert all skills)")
	}

	report := migrateSingle(cfg)
	report.Print()
}

func parseFlags() (*Config, error) {
	cfg := &Config{
		SkillsDir: "./pkg/skills",
	}

	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--input", "-i":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--input requires value")
			}
			cfg.Input = args[i+1]
			i++
		case "--output", "-o":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output requires value")
			}
			cfg.Output = args[i+1]
			i++
		case "--all":
			cfg.All = true
		case "--skills-dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--skills-dir requires value")
			}
			cfg.SkillsDir = args[i+1]
			i++
		case "--dry-run", "-d":
			cfg.DryRun = true
		case "--backup", "-b":
			cfg.Backup = true
		case "--validate":
			cfg.Validate = true
		case "--verbose", "-v":
			cfg.Verbose = true
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return cfg, nil
}

func printHelp() {
	fmt.Println("Skill Migration Tool: v3 → v4")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  skill-convert --input <path> --output <dir>    Convert single skill")
	fmt.Println("  skill-convert --all --skills-dir <dir>         Convert all skills")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --input, -i       Input skill file path")
	fmt.Println("  --output, -o      Output directory")
	fmt.Println("  --all             Convert all skills in directory")
	fmt.Println("  --skills-dir      Skills directory (default: ./pkg/skills)")
	fmt.Println("  --dry-run, -d     Show changes without writing")
	fmt.Println("  --backup, -b      Create backups before converting")
	fmt.Println("  --validate        Validate output is v4 compliant")
	fmt.Println("  --verbose, -v     Verbose output")
	fmt.Println("  --help, -h        Show this help")
}

func detectVersion(content string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	frontmatterLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				break
			}
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		}
	}

	frontmatter := strings.Join(frontmatterLines, "\n")

	hasTriggersObject := strings.Contains(frontmatter, "triggers:") &&
		strings.Contains(frontmatter, "keywords:")
	hasVersion := strings.Contains(frontmatter, "version:")
	hasXMLTags := strings.Contains(content, "<role>") ||
		strings.Contains(content, "<instructions>") ||
		strings.Contains(content, "<examples>")
	hasMarkdownSections := regexp.MustCompile(`^##\s+Role`).MatchString(content) ||
		regexp.MustCompile(`^##\s+Instructions`).MatchString(content)

	if hasTriggersObject || hasVersion {
		if hasMarkdownSections {
			return "v3"
		}
	}

	if hasXMLTags {
		return "v2"
	}

	if strings.Contains(frontmatter, "name:") && strings.Contains(frontmatter, "description:") {
		return "v1"
	}

	return "unknown"
}

func parseV3(content string) (*V3Skill, error) {
	skill := &V3Skill{}

	lines := strings.Split(content, "\n")
	inFrontmatter := false
	currentSection := ""
	var sectionBuilder strings.Builder
	var frontmatterLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				inFrontmatter = false
				parseFrontmatter(frontmatterLines, skill)
				continue
			}
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
			continue
		}

		if regexp.MustCompile(`^##\s+`).MatchString(trimmed) {
			sectionName := strings.TrimPrefix(trimmed, "## ")
			sectionName = strings.TrimSpace(sectionName)

			if currentSection != "" {
				saveSection(skill, currentSection, sectionBuilder.String())
				sectionBuilder.Reset()
			}
			currentSection = sectionName
			continue
		}

		if currentSection != "" {
			sectionBuilder.WriteString(line)
			sectionBuilder.WriteString("\n")
		}
	}

	if currentSection != "" {
		saveSection(skill, currentSection, sectionBuilder.String())
	}

	return skill, nil
}

func parseFrontmatter(lines []string, skill *V3Skill) {
	skill.Triggers = V3Triggers{Keywords: []string{}}

	inKeywordsSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inKeywordsSection {
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
				item := strings.TrimSpace(trimmed[2:])
				if item != "" {
					skill.Triggers.Keywords = append(skill.Triggers.Keywords, item)
				}
			} else if trimmed == "" || !strings.HasPrefix(line, "    ") {
				inKeywordsSection = false
			}
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			skill.Name = value
		case "description":
			skill.Description = value
		case "version":
			skill.Version = value
		case "author":
			skill.Author = value
		case "license":
			skill.License = value
		case "quality_score":
			var score int
			_, _ = fmt.Sscanf(value, "%d", &score)
			skill.QualityScore = score
		case "category":
			skill.Category = value
		case "keywords":
			switch {
			case strings.HasPrefix(value, "["):
				value = strings.Trim(value, "[]")
				items := strings.Split(value, ",")
				for _, item := range items {
					item = strings.TrimSpace(strings.Trim(item, `'"`))
					if item != "" {
						skill.Triggers.Keywords = append(skill.Triggers.Keywords, item)
					}
				}
			case value == "-" || value == "":
				inKeywordsSection = true
			default:
				skill.Triggers.Keywords = append(skill.Triggers.Keywords, value)
			}
		case "file_pattern":
			skill.Triggers.FilePattern = value
		case "weight":
			_, _ = fmt.Sscanf(value, "%f", &skill.Triggers.Weight)
		}
	}
}

func saveSection(skill *V3Skill, name, content string) {
	name = strings.TrimSpace(name)
	content = strings.TrimSpace(content)

	switch name {
	case "Role":
		skill.Role = content
	case "Instructions":
		skill.Instructions = content
	case "Constraints":
		skill.Constraints = content
	case "Edge Cases", "EdgeCases":
		skill.EdgeCases = content
	case "Examples":
		fmt.Fprintf(os.Stderr, "saveSection(Examples): content len=%d, startsWith=<example>: %v\n", len(content), strings.HasPrefix(content, "<example>"))
		skill.Examples = convertExamplesToMarkdown(content)
	case "Output Format", "OutputFormat":
		skill.OutputFormat = content
	}
}

func convertExamplesToMarkdown(content string) string {
	// Simple regex-based replacement
	// Match <example><input>...</input><output>...</output></example>
	// Use (?s) for DOTALL mode to match across newlines
	examplePattern := regexp.MustCompile(`(?s)<example>\s*<input>(.*?)</input>\s*<output>(.*?)</output>\s*</example>`)

	matches := examplePattern.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return content
	}

	var result strings.Builder
	for i, match := range matches {
		exampleNum := i + 1

		input := strings.TrimSpace(match[1])
		output := strings.TrimSpace(match[2])

		result.WriteString(fmt.Sprintf("### Example %d\n\n", exampleNum))
		result.WriteString(fmt.Sprintf("**Input**: %s\n\n", input))

		result.WriteString("**Output**:\n")
		result.WriteString(output)
		result.WriteString("\n\n")
	}

	return result.String()
}

func cleanDescription(desc string) string {
	desc = strings.TrimSpace(desc)

	if strings.HasPrefix(desc, "'") && strings.HasSuffix(desc, "'") {
		desc = desc[1 : len(desc)-1]
	}
	if strings.HasPrefix(desc, `"`) && strings.HasSuffix(desc, `"`) {
		desc = desc[1 : len(desc)-1]
	}

	if strings.Contains(desc, "Auto-activates for:") {
		idx := strings.Index(desc, "Auto-activates for:")
		if idx > 0 {
			desc = strings.TrimSpace(desc[:idx])
		} else {
			parts := strings.SplitN(desc, "Auto-activates for:", 2)
			if len(parts) > 1 && strings.TrimSpace(parts[0]) != "" {
				desc = strings.TrimSpace(parts[0])
			} else {
				desc = strings.TrimSpace(parts[1])
			}
		}
		desc = strings.TrimSuffix(desc, ".")
	}

	if len(desc) > 256 {
		desc = desc[:256]
		desc = strings.TrimSpace(desc)
	}

	return desc
}

func extractReferences(instructions, role, examples string) []Reference {
	var refs []Reference
	seen := make(map[string]bool)

	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	for _, content := range []string{instructions, role, examples} {
		matches := linkPattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			title := match[1]
			path := match[2]

			if seen[path] {
				continue
			}
			seen[path] = true

			ref := Reference{
				Title:   title,
				Path:    path,
				Content: fmt.Sprintf("Reference content for %s\n\nOriginal link: %s", title, path),
			}
			refs = append(refs, ref)
		}
	}

	return refs
}

func transformToV4(v3 *V3Skill, category string) V4Skill {
	v4 := V4Skill{
		Name:         v3.Name,
		Description:  cleanDescription(v3.Description),
		Role:         v3.Role,
		Instructions: v3.Instructions,
		Examples:     convertExamplesToMarkdown(v3.Examples),
	}

	triggers := v3.Triggers.Keywords
	if v3.Triggers.FilePattern != "" {
		triggers = append(triggers, v3.Triggers.FilePattern)
	}
	v4.Triggers = triggers

	if v3.Constraints != "" {
		if v4.Instructions != "" {
			v4.Instructions += "\n\n"
		}
		v4.Instructions += "## Constraints\n\n" + v3.Constraints
	}

	if v3.EdgeCases != "" {
		if v4.Instructions != "" {
			v4.Instructions += "\n\n"
		}
		v4.Instructions += "## Edge Cases\n\n" + v3.EdgeCases
	}

	if v3.OutputFormat != "" {
		if v4.Instructions != "" {
			v4.Instructions += "\n\n"
		}
		v4.Instructions += "## Output Format\n\n" + v3.OutputFormat
	}

	v4.References = extractReferences(v3.Instructions, v3.Role, v3.Examples)

	return v4
}

func generateV4Output(v4 *V4Skill) map[string]string {
	files := make(map[string]string)

	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", v4.Name))
	sb.WriteString(fmt.Sprintf("description: %s\n", v4.Description))
	sb.WriteString("triggers:\n")
	for _, t := range v4.Triggers {
		sb.WriteString(fmt.Sprintf("  - %s\n", t))
	}
	sb.WriteString("---\n\n")

	sb.WriteString("## Role\n\n")
	sb.WriteString(v4.Role)
	sb.WriteString("\n\n")

	sb.WriteString("## Instructions\n\n")
	sb.WriteString(v4.Instructions)
	sb.WriteString("\n\n")

	sb.WriteString("## Examples\n\n")
	sb.WriteString(v4.Examples)

	if len(v4.References) > 0 {
		sb.WriteString("\n\n## References\n\n")
		for _, ref := range v4.References {
			sb.WriteString(fmt.Sprintf("- [%s](%s)\n", ref.Title, ref.Path))
		}
	}

	files["SKILL.md"] = sb.String()

	for _, ref := range v4.References {
		var refContent strings.Builder
		refContent.WriteString(fmt.Sprintf("# %s\n\n", ref.Title))
		refContent.WriteString(ref.Content)
		files[ref.Path] = refContent.String()
	}

	return files
}

func writeFiles(outputDir string, files map[string]string, backup, dryRun bool) error {
	if dryRun {
		fmt.Println("\n📄 Dry Run - Files that would be written:")
		for path, content := range files {
			fmt.Printf("\n=== %s ===\n", filepath.Join(outputDir, path))
			preview := content
			if len(preview) > 500 {
				preview = preview[:500] + "\n... (truncated)"
			}
			fmt.Println(preview)
		}
		return nil
	}

	if backup {
		timestamp := time.Now().Unix()
		backupDir := fmt.Sprintf("%s.backup.%d", outputDir, timestamp)
		if err := os.MkdirAll(backupDir, 0o750); err != nil {
			return fmt.Errorf("create backup dir: %w", err)
		}

		err := filepath.Walk(outputDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || path == outputDir {
				return nil
			}
			relPath, err := filepath.Rel(outputDir, path)
			if err != nil {
				return err
			}
			destPath := filepath.Join(backupDir, relPath)
			destDir := filepath.Dir(destPath)
			if err := os.MkdirAll(destDir, 0o750); err != nil {
				return err
			}
			if err := copyFile(path, destPath); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Printf("⚠️  Warning: backup failed: %v\n", err)
		} else {
			fmt.Printf("📦 Backup created: %s\n", backupDir)
		}
	}

	for path, content := range files {
		fullPath := filepath.Join(outputDir, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	// #nosec G304 - src is validated by caller
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)
		destDir := filepath.Dir(destPath)

		if err := os.MkdirAll(destDir, 0o750); err != nil {
			return err
		}

		// #nosec G304 - path is validated by filepath.Walk
		return copyFile(path, destPath)
	})
}

func inferCategory(filePath string) string {
	parts := strings.Split(filepath.Clean(filePath), string(filepath.Separator))

	for i, part := range parts {
		if part == "skills" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return "other"
}

func migrateSingle(cfg *Config) *MigrationReport {
	report := &MigrationReport{
		TotalFiles: 1,
		Details:    []FileReport{},
	}

	content, err := os.ReadFile(cfg.Input)
	if err != nil {
		report.Failed = 1
		report.Details = append(report.Details, FileReport{
			SourcePath: cfg.Input,
			Status:     "failed",
			Error:      fmt.Errorf("read file: %w", err),
		})
		return report
	}

	version := detectVersion(string(content))

	if version == "unknown" {
		report.Skipped = 1
		report.Details = append(report.Details, FileReport{
			SourcePath:  cfg.Input,
			Status:      "skipped",
			VersionFrom: "unknown format",
		})
		return report
	}

	v3, err := parseV3(string(content))
	if err != nil {
		report.Failed = 1
		report.Details = append(report.Details, FileReport{
			SourcePath:  cfg.Input,
			Status:      "failed",
			VersionFrom: version,
			Error:       fmt.Errorf("parse v3: %w", err),
		})
		return report
	}

	category := inferCategory(cfg.Input)
	v4 := transformToV4(v3, category)

	changes := []string{
		"Flattened triggers array",
		"Removed version, author, license, category, quality_score",
		"Cleaned description",
	}
	if len(v4.References) > 0 {
		for _, ref := range v4.References {
			changes = append(changes, fmt.Sprintf("Created reference: %s", ref.Path))
		}
	}

	outputDir := cfg.Output
	if outputDir == "" {
		outputDir = filepath.Dir(cfg.Input)
	}

	files := generateV4Output(&v4)

	if err := writeFiles(outputDir, files, cfg.Backup, cfg.DryRun); err != nil {
		report.Failed = 1
		report.Details = append(report.Details, FileReport{
			SourcePath:  cfg.Input,
			OutputPath:  outputDir,
			Status:      "failed",
			VersionFrom: version,
			Changes:     changes,
			Error:       fmt.Errorf("write files: %w", err),
		})
		return report
	}

	report.Successful = 1
	report.Details = append(report.Details, FileReport{
		SourcePath:  cfg.Input,
		OutputPath:  outputDir,
		Status:      "success",
		VersionFrom: version,
		Changes:     changes,
	})

	fmt.Printf("\n✅ Successfully migrated %s → v4\n", cfg.Input)

	return report
}

func migrateAll(cfg *Config) *MigrationReport {
	report := &MigrationReport{
		Details: []FileReport{},
	}

	if cfg.Backup {
		timestamp := time.Now().Unix()
		backupDir := fmt.Sprintf("%s.backup.%d", cfg.SkillsDir, timestamp)
		if err := copyDir(cfg.SkillsDir, backupDir); err != nil {
			fmt.Printf("⚠️  Warning: backup failed: %v\n", err)
		} else {
			fmt.Printf("📦 Backup created: %s\n\n", backupDir)
		}
	}

	var skillFiles []string

	err := filepath.Walk(cfg.SkillsDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && strings.Contains(filepath.Base(path), "backup") {
			return filepath.SkipDir
		}

		if !info.IsDir() && filepath.Base(path) == "SKILL.md" {
			skillFiles = append(skillFiles, path)
		}
		return nil
	})
	if err != nil {
		logFatal("walk skills dir: %v", err)
	}

	report.TotalFiles = len(skillFiles)

	fmt.Printf("Found %d skill files in %s\n\n", len(skillFiles), cfg.SkillsDir)

	for _, skillFile := range skillFiles {
		singleCfg := *cfg
		singleCfg.Input = skillFile
		singleCfg.Output = filepath.Dir(skillFile)
		singleCfg.All = false
		singleCfg.Backup = false

		singleReport := migrateSingle(&singleCfg)
		report.Successful += singleReport.Successful
		report.Failed += singleReport.Failed
		report.Skipped += singleReport.Skipped
		report.Details = append(report.Details, singleReport.Details...)
	}

	return report
}

func logFatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
