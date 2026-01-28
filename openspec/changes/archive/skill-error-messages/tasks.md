# Tasks: Enhanced Validation Error Messages

## Status: complete

## 1. Foundation

### 1.1 Extend ValidationError struct
- [x] 1.1.1 Add `Suggestion` field to `ValidationError`
- [x] 1.1.2 Add `Example` field to `ValidationError`
- [x] 1.1.3 Add `Suggestion` field to `ValidationWarning`
- [x] 1.1.4 Add `Example` field to `ValidationWarning`
- [x] 1.1.5 Verify existing tests pass without modification

### 1.2 Update CLI formatter
- [x] 1.2.1 Display suggestions when present in errors
- [x] 1.2.2 Display examples when present in errors
- [x] 1.2.3 Ensure format is readable and well-structured
- [x] 1.2.4 Handle empty suggestion/example fields gracefully

## 2. Enhanced Rules

### 2.1 Update SK001-SK003 (name, format, description)
- [x] 2.1.1 Add suggestion and example to SK001
- [x] 2.1.2 Add suggestion and example to SK002
- [x] 2.1.3 Add suggestion and example to SK003

### 2.2 Update SK004-SK006 (examples, role, instructions)
- [x] 2.2.1 Add suggestion with proper XML example to SK004
- [x] 2.2.2 Add suggestion with role example to SK005
- [x] 2.2.3 Add suggestion with instructions example to SK006

### 2.3 Update SK007-SK009 (constraints, output, edge cases)
- [x] 2.3.1 Add suggestion and example to SK007
- [x] 2.3.2 Add suggestion and example to SK008
- [x] 2.3.3 Add suggestion and example to SK009

## 3. Testing

### 3.1 Unit tests for enhanced errors
- [x] 3.1.1 Test ValidationError with all fields populated
- [x] 3.1.2 Test ValidationError with empty suggestion/example
- [x] 3.1.3 Test CLI formatter output format
- [x] 3.1.4 Verify all existing tests pass

### 3.2 Integration tests for all rules
- [x] 3.2.1 Verify each rule test has suggestion present
- [x] 3.2.2 Verify each rule test has example present
- [x] 3.2.3 Test invalid skills show helpful errors
- [x] 3.2.4 Test valid skills pass without errors
