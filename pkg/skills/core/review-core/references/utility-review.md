# Simple Utility Function Review

Example showing code review feedback for a straightforward utility function.

## Example

<example>
<input>Review feedback for simple utility function</input>
<output>
## Review Feedback: PR #134 - Add UUID validator

### Code Review

```go
func IsValidUUID(uuid string) bool {
    if len(uuid) != 36 {
        return false
    }
    pattern := `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`
    matched, _ := regexp.MatchString(pattern, uuid)
    return matched
}
```

---

### Issues (P2 - Suggestions)

**💡 1. Error handling ignored**

```go
matched, _ := regexp.MatchString(pattern, uuid)
```

**Fix**:
```go
matched, err := regexp.MatchString(pattern, uuid)
if err != nil {
    return false
}
```

---

**💡 2. Compiles regex on every call**

**Fix**: Use package-level compiled regex:
```go
var uuidRegex = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)

func IsValidUUID(uuid string) bool {
    return uuidRegex.MatchString(uuid)
}
```

---

**💡 3. Testing improvements**

Add `t.Run()` and `t.Parallel()`:
```go
for _, tt := range tests {
    tt := tt
    t.Run(tt.input, func(t *testing.T) {
        t.Parallel()
        if got := IsValidUUID(tt.input); got != tt.want {
            t.Errorf("IsValidUUID(%q) = %v, want %v", tt.input, got, tt.want)
        }
    })
}
```

---

### Summary

**Action**: Approve with suggestions

**Non-blocking**: Handle regex error, compile regex once, add t.Run() and t.Parallel()

**Notes**: Function works correctly. Suggestions are for robustness and performance.
</output>
</example>
