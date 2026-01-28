# Tasks: Explicit Skill Triggers

## Status: complete

## 1. Add Trigger struct to parser
- [x] 1.1 Define Trigger struct with Pattern, Keywords, FilePattern, Weight fields
- [x] 1.2 Add Triggers []Trigger field to Frontmatter struct
- [x] 1.3 Add YAML parsing for triggers section
- [x] 1.4 Default weight 0.7 if not specified
- [x] 1.5 Validate weight range 0.0-1.0

## 2. Backward-compatible trigger extraction
- [x] 2.1 Implement getTriggers() to return explicit triggers if present
- [x] 2.2 Add fallback to description-based extraction if no explicit triggers
- [x] 2.3 Set fallback triggers weight to 0.5
- [x] 2.4 Verify existing skills continue working

## 3. Add SK012 validation rule
- [x] 3.1 Add SK012 rule in rules.go to check for explicit triggers
- [x] 3.2 Return info-level warning if using description-based triggers
- [x] 3.3 Include example of explicit trigger format in warning
- [x] 3.4 No warning if explicit triggers present

## 4. Update tests
- [x] 4.1 Test parsing explicit triggers
- [x] 4.2 Test weight validation
- [x] 4.3 Test backward compatibility fallback
- [x] 4.4 Test SK012 validation rule
