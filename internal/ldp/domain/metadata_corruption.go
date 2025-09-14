package domain

import (
	"fmt"
	"time"
)

// MetadataCorruptionDetector detects and repairs corrupted container metadata
type MetadataCorruptionDetector struct {
	timestampManager *TimestampManager
}

// NewMetadataCorruptionDetector creates a new metadata corruption detector
func NewMetadataCorruptionDetector(timestampManager *TimestampManager) *MetadataCorruptionDetector {
	return &MetadataCorruptionDetector{
		timestampManager: timestampManager,
	}
}

// CorruptionType represents different types of metadata corruption
type CorruptionType string

const (
	CorruptionTypeInvalidTimestamp     CorruptionType = "INVALID_TIMESTAMP"
	CorruptionTypeInvalidContainerType CorruptionType = "INVALID_CONTAINER_TYPE"
	CorruptionTypeInvalidMemberList    CorruptionType = "INVALID_MEMBER_LIST"
	CorruptionTypeInvalidParentID      CorruptionType = "INVALID_PARENT_ID"
	CorruptionTypeMissingRequiredField CorruptionType = "MISSING_REQUIRED_FIELD"
	CorruptionTypeInvalidMetadataType  CorruptionType = "INVALID_METADATA_TYPE"
)

// CorruptionIssue represents a detected corruption issue
type CorruptionIssue struct {
	Type        CorruptionType `json:"type"`
	Field       string         `json:"field"`
	Description string         `json:"description"`
	Severity    string         `json:"severity"` // "low", "medium", "high", "critical"
}

// CorruptionReport contains all detected corruption issues
type CorruptionReport struct {
	ContainerID string            `json:"containerId"`
	Issues      []CorruptionIssue `json:"issues"`
	IsCorrupted bool              `json:"isCorrupted"`
	Timestamp   time.Time         `json:"timestamp"`
}

// DetectCorruption analyzes a container for metadata corruption
func (mcd *MetadataCorruptionDetector) DetectCorruption(container *Container) *CorruptionReport {
	report := &CorruptionReport{
		ContainerID: container.ID(),
		Issues:      make([]CorruptionIssue, 0),
		IsCorrupted: false,
		Timestamp:   time.Now(),
	}

	// Check timestamp corruption
	mcd.checkTimestampCorruption(container, report)

	// Check container type corruption
	mcd.checkContainerTypeCorruption(container, report)

	// Check member list corruption
	mcd.checkMemberListCorruption(container, report)

	// Check parent ID corruption
	mcd.checkParentIDCorruption(container, report)

	// Check required fields
	mcd.checkRequiredFields(container, report)

	// Check metadata type corruption
	mcd.checkMetadataTypeCorruption(container, report)

	report.IsCorrupted = len(report.Issues) > 0
	return report
}

// checkTimestampCorruption checks for timestamp-related corruption
func (mcd *MetadataCorruptionDetector) checkTimestampCorruption(container *Container, report *CorruptionReport) {
	metadata := container.GetMetadata()

	// Check if createdAt exists and is valid
	if createdAt, exists := metadata["createdAt"]; exists {
		if _, ok := createdAt.(time.Time); !ok {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeInvalidTimestamp,
				Field:       "createdAt",
				Description: fmt.Sprintf("createdAt field has invalid type: %T", createdAt),
				Severity:    "high",
			})
		}
	} else {
		report.Issues = append(report.Issues, CorruptionIssue{
			Type:        CorruptionTypeMissingRequiredField,
			Field:       "createdAt",
			Description: "createdAt field is missing",
			Severity:    "high",
		})
	}

	// Check if updatedAt exists and is valid
	if updatedAt, exists := metadata["updatedAt"]; exists {
		if _, ok := updatedAt.(time.Time); !ok {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeInvalidTimestamp,
				Field:       "updatedAt",
				Description: fmt.Sprintf("updatedAt field has invalid type: %T", updatedAt),
				Severity:    "high",
			})
		}
	} else {
		report.Issues = append(report.Issues, CorruptionIssue{
			Type:        CorruptionTypeMissingRequiredField,
			Field:       "updatedAt",
			Description: "updatedAt field is missing",
			Severity:    "high",
		})
	}

	// Check timestamp order if both exist and are valid
	if err := mcd.timestampManager.ValidateTimestamps(container); err != nil {
		report.Issues = append(report.Issues, CorruptionIssue{
			Type:        CorruptionTypeInvalidTimestamp,
			Field:       "timestamps",
			Description: err.Error(),
			Severity:    "medium",
		})
	}
}

// checkContainerTypeCorruption checks for container type corruption
func (mcd *MetadataCorruptionDetector) checkContainerTypeCorruption(container *Container, report *CorruptionReport) {
	if !container.ContainerType.IsValid() {
		report.Issues = append(report.Issues, CorruptionIssue{
			Type:        CorruptionTypeInvalidContainerType,
			Field:       "containerType",
			Description: fmt.Sprintf("invalid container type: %s", container.ContainerType),
			Severity:    "critical",
		})
	}

	// Check metadata consistency
	metadata := container.GetMetadata()
	if containerType, exists := metadata["containerType"]; exists {
		if typeStr, ok := containerType.(string); ok {
			if typeStr != container.ContainerType.String() {
				report.Issues = append(report.Issues, CorruptionIssue{
					Type:        CorruptionTypeInvalidContainerType,
					Field:       "containerType",
					Description: fmt.Sprintf("container type mismatch: field=%s, metadata=%s", container.ContainerType, typeStr),
					Severity:    "medium",
				})
			}
		} else {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeInvalidMetadataType,
				Field:       "containerType",
				Description: fmt.Sprintf("containerType metadata has invalid type: %T", containerType),
				Severity:    "medium",
			})
		}
	}
}

// checkMemberListCorruption checks for member list corruption
func (mcd *MetadataCorruptionDetector) checkMemberListCorruption(container *Container, report *CorruptionReport) {
	if container.Members == nil {
		report.Issues = append(report.Issues, CorruptionIssue{
			Type:        CorruptionTypeInvalidMemberList,
			Field:       "members",
			Description: "members list is nil",
			Severity:    "high",
		})
		return
	}

	// Check for duplicate members
	memberSet := make(map[string]bool)
	for _, member := range container.Members {
		if member == "" {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeInvalidMemberList,
				Field:       "members",
				Description: "empty member ID found in members list",
				Severity:    "medium",
			})
		}

		if memberSet[member] {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeInvalidMemberList,
				Field:       "members",
				Description: fmt.Sprintf("duplicate member ID found: %s", member),
				Severity:    "medium",
			})
		}
		memberSet[member] = true
	}
}

// checkParentIDCorruption checks for parent ID corruption
func (mcd *MetadataCorruptionDetector) checkParentIDCorruption(container *Container, report *CorruptionReport) {
	// Check for self-reference
	if container.ParentID == container.ID() {
		report.Issues = append(report.Issues, CorruptionIssue{
			Type:        CorruptionTypeInvalidParentID,
			Field:       "parentID",
			Description: "container cannot be its own parent",
			Severity:    "critical",
		})
	}

	// Check metadata consistency
	metadata := container.GetMetadata()
	if parentID, exists := metadata["parentID"]; exists {
		if parentIDStr, ok := parentID.(string); ok {
			if parentIDStr != container.ParentID {
				report.Issues = append(report.Issues, CorruptionIssue{
					Type:        CorruptionTypeInvalidParentID,
					Field:       "parentID",
					Description: fmt.Sprintf("parent ID mismatch: field=%s, metadata=%s", container.ParentID, parentIDStr),
					Severity:    "medium",
				})
			}
		} else {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeInvalidMetadataType,
				Field:       "parentID",
				Description: fmt.Sprintf("parentID metadata has invalid type: %T", parentID),
				Severity:    "medium",
			})
		}
	}
}

// checkRequiredFields checks for missing required fields
func (mcd *MetadataCorruptionDetector) checkRequiredFields(container *Container, report *CorruptionReport) {
	metadata := container.GetMetadata()

	requiredFields := []string{"type", "containerType"}
	for _, field := range requiredFields {
		if _, exists := metadata[field]; !exists {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeMissingRequiredField,
				Field:       field,
				Description: fmt.Sprintf("required field %s is missing", field),
				Severity:    "high",
			})
		}
	}
}

// checkMetadataTypeCorruption checks for invalid metadata types
func (mcd *MetadataCorruptionDetector) checkMetadataTypeCorruption(container *Container, report *CorruptionReport) {
	metadata := container.GetMetadata()

	// Check Dublin Core fields for correct types
	dublinCoreFields := map[string]string{
		"dc:title":       "string",
		"dc:description": "string",
		"dc:creator":     "string",
		"dc:subject":     "string",
		"dc:publisher":   "string",
		"dc:contributor": "string",
		"dc:type":        "string",
		"dc:format":      "string",
		"dc:identifier":  "string",
		"dc:source":      "string",
		"dc:language":    "string",
		"dc:relation":    "string",
		"dc:coverage":    "string",
		"dc:rights":      "string",
	}

	for field, expectedType := range dublinCoreFields {
		if value, exists := metadata[field]; exists {
			switch expectedType {
			case "string":
				if _, ok := value.(string); !ok {
					report.Issues = append(report.Issues, CorruptionIssue{
						Type:        CorruptionTypeInvalidMetadataType,
						Field:       field,
						Description: fmt.Sprintf("%s should be string but is %T", field, value),
						Severity:    "low",
					})
				}
			}
		}
	}

	// Check dc:date separately as it should be time.Time
	if dcDate, exists := metadata["dc:date"]; exists {
		if _, ok := dcDate.(time.Time); !ok {
			report.Issues = append(report.Issues, CorruptionIssue{
				Type:        CorruptionTypeInvalidMetadataType,
				Field:       "dc:date",
				Description: fmt.Sprintf("dc:date should be time.Time but is %T", dcDate),
				Severity:    "low",
			})
		}
	}
}

// RepairCorruption attempts to repair detected corruption issues
func (mcd *MetadataCorruptionDetector) RepairCorruption(container *Container, report *CorruptionReport) (bool, error) {
	if !report.IsCorrupted {
		return false, nil // Nothing to repair
	}

	repaired := false

	for _, issue := range report.Issues {
		switch issue.Type {
		case CorruptionTypeInvalidTimestamp:
			if issue.Field == "createdAt" || issue.Field == "updatedAt" || issue.Field == "timestamps" {
				if mcd.timestampManager.RepairTimestamps(container) {
					repaired = true
				}
			}

		case CorruptionTypeInvalidContainerType:
			if issue.Field == "containerType" {
				// If container type is invalid, default to BasicContainer
				if !container.ContainerType.IsValid() {
					container.ContainerType = BasicContainer
					container.SetMetadata("containerType", BasicContainer.String())
					repaired = true
				}
				// Fix metadata consistency
				container.SetMetadata("containerType", container.ContainerType.String())
				repaired = true
			}

		case CorruptionTypeInvalidMemberList:
			if issue.Field == "members" {
				// Remove duplicates and empty members
				uniqueMembers := make([]string, 0)
				memberSet := make(map[string]bool)
				for _, member := range container.Members {
					if member != "" && !memberSet[member] {
						uniqueMembers = append(uniqueMembers, member)
						memberSet[member] = true
					}
				}
				if len(uniqueMembers) != len(container.Members) {
					container.Members = uniqueMembers
					repaired = true
				}
			}

		case CorruptionTypeInvalidParentID:
			if issue.Field == "parentID" {
				// If container is its own parent, clear parent ID
				if container.ParentID == container.ID() {
					container.ParentID = ""
					container.SetMetadata("parentID", "")
					repaired = true
				}
				// Fix metadata consistency
				container.SetMetadata("parentID", container.ParentID)
				repaired = true
			}

		case CorruptionTypeMissingRequiredField:
			// Add missing required fields
			if issue.Field == "type" {
				container.SetMetadata("type", "Container")
				repaired = true
			} else if issue.Field == "containerType" {
				container.SetMetadata("containerType", container.ContainerType.String())
				repaired = true
			}

		case CorruptionTypeInvalidMetadataType:
			// Remove invalid metadata types (they can be re-set correctly later)
			metadata := container.GetMetadata()
			delete(metadata, issue.Field)
			repaired = true
		}
	}

	if repaired {
		// Update timestamp after repair
		mcd.timestampManager.UpdateTimestamp(container)

		// Emit repair event
		event := NewContainerUpdatedEvent(container.ID(), map[string]interface{}{
			"repaired":  true,
			"issues":    len(report.Issues),
			"updatedAt": time.Now(),
		})
		container.AddEvent(event)
	}

	return repaired, nil
}
