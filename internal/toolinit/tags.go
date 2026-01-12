package toolinit

// Tag constants for agent categorization

// Role tags
const (
	RolePlanning  = "planning"
	RoleExecution = "execution"
	RoleReview    = "review"
	RoleDebug     = "debug"
	RoleTest      = "test"
)

// Complexity tags
const (
	ComplexityLight    = "light"
	ComplexityStandard = "standard"
	ComplexityHeavy    = "heavy"
)

// ValidRoles returns all valid role tags
func ValidRoles() []string {
	return []string{
		RolePlanning,
		RoleExecution,
		RoleReview,
		RoleDebug,
		RoleTest,
	}
}

// ValidComplexities returns all valid complexity tags
func ValidComplexities() []string {
	return []string{
		ComplexityLight,
		ComplexityStandard,
		ComplexityHeavy,
	}
}

// IsValidRole checks if a role tag is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// IsValidComplexity checks if a complexity tag is valid
func IsValidComplexity(complexity string) bool {
	for _, c := range ValidComplexities() {
		if c == complexity {
			return true
		}
	}
	return false
}
