# Tasks: Skill Dependencies

## Status
complete

## 1. Extend frontmatter
- [x] 1.1 Add `depends_on` []string field
- [x] 1.2 Add `delegates_to` map[string]string field
- [x] 1.3 Update parser

## 2. Dependency resolution
- [x] 2.1 Implement topological sort for load order
- [x] 2.2 Detect circular dependencies
- [x] 2.3 Error on missing dependencies

## 3. Delegation hints
- [x] 3.1 Store delegation metadata
- [x] 3.2 Include in match results when relevant

## 4. Testing
- [x] 4.1 Test dependency resolution
- [x] 4.2 Test circular detection
- [x] 4.3 Test delegation metadata
