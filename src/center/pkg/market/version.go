package market

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// VersionManager manages skill versions
type VersionManager struct {
	market *CapabilityMarket

	mu sync.RWMutex
}

// NewVersionManager creates a new version manager
func NewVersionManager(market *CapabilityMarket) *VersionManager {
	return &VersionManager{
		market: market,
	}
}

// VersionInfo contains version comparison info
type VersionInfo struct {
	Major   int
	Minor   int
	Patch   int
	Pre     string // Pre-release (e.g., "alpha", "beta")
	Build   string // Build metadata
}

// ParseVersion parses a version string
func ParseVersion(version string) (*VersionInfo, error) {
	v := &VersionInfo{}

	// Handle semantic versioning: major.minor.patch[-pre][+build]
	parts := splitVersion(version)

	if len(parts) < 1 {
		return nil, errors.New("invalid version format")
	}

	// Parse major
	major, err := parseInt(parts[0])
	if err != nil {
		return nil, err
	}
	v.Major = major

	if len(parts) > 1 {
		minor, err := parseInt(parts[1])
		if err != nil {
			return nil, err
		}
		v.Minor = minor
	}

	if len(parts) > 2 {
		patch, err := parseInt(parts[2])
		if err != nil {
			return nil, err
		}
		v.Patch = patch
	}

	return v, nil
}

// Compare compares two versions
func (v *VersionInfo) Compare(other *VersionInfo) int {
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	// Pre-release versions have lower precedence
	if v.Pre != "" && other.Pre == "" {
		return -1
	}
	if v.Pre == "" && other.Pre != "" {
		return 1
	}

	return 0
}

// String returns the version string
func (v *VersionInfo) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// IsNewer checks if version is newer than another
func (v *VersionInfo) IsNewer(other *VersionInfo) bool {
	return v.Compare(other) > 0
}

// IsCompatible checks if versions are compatible (same major)
func (v *VersionInfo) IsCompatible(other *VersionInfo) bool {
	return v.Major == other.Major
}

// GetLatestVersion returns the latest version of a skill
func (vm *VersionManager) GetLatestVersion(skillID string) (string, error) {
	versions, err := vm.market.GetVersions(skillID)
	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", errors.New("no versions available")
	}

	latest := versions[0]
	latestInfo, err := ParseVersion(latest)
	if err != nil {
		return "", err
	}

	for _, v := range versions[1:] {
		info, err := ParseVersion(v)
		if err != nil {
			continue
		}

		if info.IsNewer(latestInfo) {
			latest = v
			latestInfo = info
		}
	}

	return latest, nil
}

// GetCompatibleVersions returns versions compatible with given version
func (vm *VersionManager) GetCompatibleVersions(skillID, baseVersion string) ([]string, error) {
	versions, err := vm.market.GetVersions(skillID)
	if err != nil {
		return nil, err
	}

	baseInfo, err := ParseVersion(baseVersion)
	if err != nil {
		return nil, err
	}

	var compatible []string
	for _, v := range versions {
		info, err := ParseVersion(v)
		if err != nil {
			continue
		}

		if info.IsCompatible(baseInfo) {
			compatible = append(compatible, v)
		}
	}

	return compatible, nil
}

// CheckForUpdates checks if there's a newer version available
func (vm *VersionManager) CheckForUpdates(skillID, currentVersion string) (string, bool, error) {
	latest, err := vm.GetLatestVersion(skillID)
	if err != nil {
		return "", false, err
	}

	currentInfo, err := ParseVersion(currentVersion)
	if err != nil {
		return "", false, err
	}

	latestInfo, err := ParseVersion(latest)
	if err != nil {
		return "", false, err
	}

	return latest, latestInfo.IsNewer(currentInfo), nil
}

// VersionHistory represents version history
type VersionHistory struct {
	SkillID   string
	Version   string
	Changes   []ChangeRecord
	Published time.Time
}

// ChangeRecord represents a change in a version
type ChangeRecord struct {
	Type     ChangeType // added, changed, deprecated, removed, fixed, security
	Area     string     // feature area affected
	Description string
}

// ChangeType represents type of change
type ChangeType string

const (
	ChangeAdded     ChangeType = "added"
	ChangeChanged   ChangeType = "changed"
	ChangeDeprecated ChangeType = "deprecated"
	ChangeRemoved   ChangeType = "removed"
	ChangeFixed     ChangeType = "fixed"
	ChangeSecurity  ChangeType = "security"
)

// GetVersionHistory returns the history of a skill version
func (vm *VersionManager) GetVersionHistory(skillID, version string) (*VersionHistory, error) {
	pkg, err := vm.market.GetSkillVersion(skillID, version)
	if err != nil {
		return nil, err
	}

	// Check if history is stored in metadata
	if pkg.Metadata.Metadata != nil {
		if historyJSON, ok := pkg.Metadata.Metadata["version_history"]; ok {
			var history VersionHistory
			if err := json.Unmarshal([]byte(historyJSON), &history); err == nil {
				return &history, nil
			}
		}
	}

	// Return basic history
	return &VersionHistory{
		SkillID:   skillID,
		Version:   version,
		Published: pkg.Metadata.PublishedAt,
	}, nil
}

// splitVersion splits version string into parts
func splitVersion(version string) []string {
	// Remove build metadata
	buildIdx := -1
	for i, c := range version {
		if c == '+' {
			buildIdx = i
			break
		}
	}

	if buildIdx > 0 {
		version = version[:buildIdx]
	}

	// Remove pre-release
	preIdx := -1
	for i, c := range version {
		if c == '-' {
			preIdx = i
			break
		}
	}

	var pre string
	if preIdx > 0 {
		pre = version[preIdx+1:]
		version = version[:preIdx]
	}

	// Split by dots
	var parts []string
	start := 0
	for i, c := range version {
		if c == '.' {
			parts = append(parts, version[start:i])
			start = i + 1
		}
	}
	if start < len(version) {
		parts = append(parts, version[start:])
	}

	return parts
}

// parseInt safely parses an integer
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		result = result * 10 + int(c - '0')
	}
	return result, nil
}