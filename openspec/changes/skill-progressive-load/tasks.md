# Tasks: Progressive Loading

## 1. Define LoadLevel enum
- [ ] 1.1 Add LoadLevel type (Metadata, Core, Extended)
- [ ] 1.2 Update Skill struct with load level tracking

## 2. Implement level parsing
- [ ] 2.1 Parse frontmatter + triggers (Level 1)
- [ ] 2.2 Parse core sections (Level 2)
- [ ] 2.3 Parse full body + references (Level 3)

## 3. Registry load management
- [ ] 3.1 Load Level 1 by default
- [ ] 3.2 Upgrade to Level 2 on match
- [ ] 3.3 Upgrade to Level 3 on execution

## 4. Testing
- [ ] 4.1 Test each load level
- [ ] 4.2 Measure token usage per level
- [ ] 4.3 Verify lazy loading works
