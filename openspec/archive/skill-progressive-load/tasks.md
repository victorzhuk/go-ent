# Tasks: Progressive Loading

## 1. Define LoadLevel enum
- [x] 1.1 Add LoadLevel type (Metadata, Core, Extended)
- [x] 1.2 Update Skill struct with load level tracking

## 2. Implement level parsing
- [x] 2.1 Parse frontmatter + triggers (Level 1)
- [x] 2.2 Parse core sections (Level 2)
- [x] 2.3 Parse full body + references (Level 3)

## 3. Registry load management
- [x] 3.1 Load Level 1 by default
- [x] 3.2 Upgrade to Level 2 on match
- [x] 3.3 Upgrade to Level 3 on execution

## 4. Testing
- [x] 4.1 Test each load level
- [x] 4.2 Measure token usage per level
- [x] 4.3 Verify lazy loading works
