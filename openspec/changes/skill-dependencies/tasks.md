# Tasks: Skill Dependencies

## 1. Extend frontmatter
- [ ] 1.1 Add `depends_on` []string field
- [ ] 1.2 Add `delegates_to` map[string]string field
- [ ] 1.3 Update parser

## 2. Dependency resolution
- [ ] 2.1 Implement topological sort for load order
- [ ] 2.2 Detect circular dependencies
- [ ] 2.3 Error on missing dependencies

## 3. Delegation hints
- [ ] 3.1 Store delegation metadata
- [ ] 3.2 Include in match results when relevant

## 4. Testing
- [ ] 4.1 Test dependency resolution
- [ ] 4.2 Test circular detection
- [ ] 4.3 Test delegation metadata
