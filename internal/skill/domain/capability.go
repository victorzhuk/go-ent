package domain

// CapabilityType represents the type of skill capability.
type CapabilityType string

const (
	// CapabilityTypeCode indicates code-related capabilities.
	CapabilityTypeCode CapabilityType = "code"
	// CapabilityTypeReview indicates review-related capabilities.
	CapabilityTypeReview CapabilityType = "review"
	// CapabilityTypeTest indicates test-related capabilities.
	CapabilityTypeTest CapabilityType = "test"
	// CapabilityTypeDebug indicates debug-related capabilities.
	CapabilityTypeDebug CapabilityType = "debug"
	// CapabilityTypeRefactor indicates refactor-related capabilities.
	CapabilityTypeRefactor CapabilityType = "refactor"
	// CapabilityTypeDesign indicates design-related capabilities.
	CapabilityTypeDesign CapabilityType = "design"
	// CapabilityTypeDocs indicates documentation-related capabilities.
	CapabilityTypeDocs CapabilityType = "docs"
	// CapabilityTypeBuild indicates build-related capabilities.
	CapabilityTypeBuild CapabilityType = "build"
	// CapabilityTypeDeploy indicates deployment-related capabilities.
	CapabilityTypeDeploy CapabilityType = "deploy"
)

// SkillCapability represents a specific capability a skill provides.
type SkillCapability struct {
	Type        CapabilityType
	Description string
	Confidence  float64
}

// NewSkillCapability creates a new SkillCapability instance.
func NewSkillCapability(cType CapabilityType, description string, confidence float64) (*SkillCapability, error) {
	if description == "" {
		return nil, ErrInvalidCapabilityDesc
	}
	if confidence < 0 || confidence > 1 {
		return nil, ErrInvalidConfidence
	}

	return &SkillCapability{
		Type:        cType,
		Description: description,
		Confidence:  confidence,
	}, nil
}
