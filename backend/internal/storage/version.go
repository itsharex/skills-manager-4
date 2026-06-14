package storage

import (
	"sort"
	"strconv"
	"strings"
)

type semVer struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

// parseSemVer attempts to parse a semantic version string.
// Returns false if the string is not a valid semver.
func parseSemVer(v string) (semVer, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return semVer{}, false
	}

	var prerelease string
	if idx := strings.Index(v, "-"); idx != -1 {
		prerelease = v[idx+1:]
		v = v[:idx]
	}

	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return semVer{}, false
	}

	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return semVer{}, false
	}

	return semVer{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: prerelease,
	}, true
}

func intCompare(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// CompareVersions compares two semantic versions.
// Returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2.
func CompareVersions(v1, v2 string) int {
	p1, ok1 := parseSemVer(v1)
	p2, ok2 := parseSemVer(v2)

	// Invalid/empty versions sort before valid ones
	if !ok1 && !ok2 {
		return strings.Compare(v1, v2)
	}
	if !ok1 {
		return -1
	}
	if !ok2 {
		return 1
	}

	// Compare major.minor.patch
	if p1.major != p2.major {
		return intCompare(p1.major, p2.major)
	}
	if p1.minor != p2.minor {
		return intCompare(p1.minor, p2.minor)
	}
	if p1.patch != p2.patch {
		return intCompare(p1.patch, p2.patch)
	}

	// Same base version, compare pre-release
	// Non-prerelease > prerelease (e.g., 1.0.0 > 1.0.0-beta)
	if p1.prerelease == "" && p2.prerelease != "" {
		return 1
	}
	if p1.prerelease != "" && p2.prerelease == "" {
		return -1
	}
	return strings.Compare(p1.prerelease, p2.prerelease)
}

// SortVersions sorts a slice of version strings in ascending order.
func SortVersions(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		return CompareVersions(versions[i], versions[j]) < 0
	})
}

// LatestVersion returns the latest version from a slice.
// Returns empty string if slice is empty.
func LatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	latest := versions[0]
	for _, v := range versions[1:] {
		if CompareVersions(v, latest) > 0 {
			latest = v
		}
	}
	return latest
}