package storage

import (
	"testing"
)

func TestCompareVersions_Equal(t *testing.T) {
	if c := CompareVersions("1.0.0", "1.0.0"); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
}

func TestCompareVersions_Major(t *testing.T) {
	if c := CompareVersions("2.0.0", "1.0.0"); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
	if c := CompareVersions("1.0.0", "2.0.0"); c != -1 {
		t.Errorf("expected -1, got %d", c)
	}
}

func TestCompareVersions_Minor(t *testing.T) {
	if c := CompareVersions("1.2.0", "1.1.0"); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
	if c := CompareVersions("1.1.0", "1.2.0"); c != -1 {
		t.Errorf("expected -1, got %d", c)
	}
}

func TestCompareVersions_Patch(t *testing.T) {
	if c := CompareVersions("1.0.3", "1.0.2"); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
	if c := CompareVersions("1.0.2", "1.0.3"); c != -1 {
		t.Errorf("expected -1, got %d", c)
	}
}

func TestCompareVersions_Prerelease(t *testing.T) {
	// Pre-release is lower than release
	if c := CompareVersions("1.0.0", "1.0.0-beta"); c != 1 {
		t.Errorf("expected 1 (release > prerelease), got %d", c)
	}
	if c := CompareVersions("1.0.0-beta", "1.0.0"); c != -1 {
		t.Errorf("expected -1 (prerelease < release), got %d", c)
	}
}

func TestCompareVersions_PrereleaseOrder(t *testing.T) {
	if c := CompareVersions("1.0.0-alpha", "1.0.0-beta"); c != -1 {
		t.Errorf("expected -1 (alpha < beta), got %d", c)
	}
	if c := CompareVersions("1.0.0-beta", "1.0.0-alpha"); c != 1 {
		t.Errorf("expected 1 (beta > alpha), got %d", c)
	}
}

func TestCompareVersions_EmptyString(t *testing.T) {
	// Empty sorts before valid
	if c := CompareVersions("", "1.0.0"); c != -1 {
		t.Errorf("expected -1, got %d", c)
	}
	if c := CompareVersions("1.0.0", ""); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
	if c := CompareVersions("", ""); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
}

func TestCompareVersions_InvalidSemver(t *testing.T) {
	// Invalid versions sort before valid ones
	if c := CompareVersions("not-a-version", "1.0.0"); c != -1 {
		t.Errorf("expected -1, got %d", c)
	}
	if c := CompareVersions("1.0.0", "not-a-version"); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
}

func TestCompareVersions_BothInvalid(t *testing.T) {
	// Two invalid versions compare lexicographically
	if c := CompareVersions("aaa", "bbb"); c != -1 {
		t.Errorf("expected -1 (aaa < bbb lexicographically), got %d", c)
	}
	if c := CompareVersions("zzz", "aaa"); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
}

func TestCompareVersions_PartialSemver(t *testing.T) {
	// "1.0" or "1" are not valid semver (need 3 parts)
	if c := CompareVersions("1.0", "1.0.0"); c != -1 {
		t.Errorf("expected -1 (invalid < valid), got %d", c)
	}
}

func TestCompareVersions_PrereleaseSameVersion(t *testing.T) {
	if c := CompareVersions("1.0.0-rc.1", "1.0.0-rc.1"); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
}

func TestCompareVersions_ZeroVersions(t *testing.T) {
	if c := CompareVersions("0.0.0", "0.0.0"); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
	if c := CompareVersions("0.0.1", "0.0.0"); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
}

func TestSortVersions(t *testing.T) {
	versions := []string{"1.0.0", "0.5.0", "2.0.0", "1.2.0", "1.0.0-beta"}
	SortVersions(versions)

	expected := []string{"0.5.0", "1.0.0-beta", "1.0.0", "1.2.0", "2.0.0"}
	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestSortVersions_WithInvalid(t *testing.T) {
	versions := []string{"2.0.0", "invalid", "", "1.0.0"}
	SortVersions(versions)

	// Invalid/empty sort before valid ones
	if versions[0] != "" && versions[0] != "invalid" {
		t.Errorf("expected empty or invalid first, got %q", versions[0])
	}
	if versions[len(versions)-1] != "2.0.0" {
		t.Errorf("expected '2.0.0' last, got %q", versions[len(versions)-1])
	}
}

func TestSortVersions_InPlace(t *testing.T) {
	original := []string{"3.0.0", "1.0.0", "2.0.0"}
	before := make([]string, len(original))
	copy(before, original)
	SortVersions(original)

	// Verify it was sorted in-place (same slice, not new allocation)
	if &original[0] == &before[0] {
		// It's the same backing array, which means in-place
	}

	expected := []string{"1.0.0", "2.0.0", "3.0.0"}
	for i, v := range original {
		if v != expected[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestSortVersions_Empty(t *testing.T) {
	// Should not panic
	SortVersions([]string{})
}

func TestSortVersions_Single(t *testing.T) {
	versions := []string{"1.0.0"}
	SortVersions(versions)
	if versions[0] != "1.0.0" {
		t.Errorf("expected '1.0.0', got %q", versions[0])
	}
}

func TestSortVersions_Nil(t *testing.T) {
	// Should not panic
	SortVersions(nil)
}

func TestLatestVersion(t *testing.T) {
	versions := []string{"1.0.0", "2.0.0", "0.5.0", "1.5.0"}
	latest := LatestVersion(versions)
	if latest != "2.0.0" {
		t.Errorf("expected '2.0.0', got %q", latest)
	}
}

func TestLatestVersion_WithPrerelease(t *testing.T) {
	versions := []string{"1.0.0", "2.0.0-beta", "1.5.0"}
	latest := LatestVersion(versions)
	// 2.x has higher major version than 1.x even with pre-release
	if latest != "2.0.0-beta" {
		t.Errorf("expected '2.0.0-beta', got %q", latest)
	}
}

func TestLatestVersion_Empty(t *testing.T) {
	latest := LatestVersion([]string{})
	if latest != "" {
		t.Errorf("expected empty string, got %q", latest)
	}
}

func TestLatestVersion_Single(t *testing.T) {
	latest := LatestVersion([]string{"1.0.0"})
	if latest != "1.0.0" {
		t.Errorf("expected '1.0.0', got %q", latest)
	}
}

func TestLatestVersion_AllInvalid(t *testing.T) {
	versions := []string{"invalid", "also-invalid", "zzz"}
	latest := LatestVersion(versions)
	// All invalid, lexicographic order: "also-invalid" < "invalid" < "zzz"
	if latest != "zzz" {
		t.Errorf("expected 'zzz' (lexicographically largest), got %q", latest)
	}
}

func TestLatestVersion_Mixed(t *testing.T) {
	versions := []string{"invalid", "2.0.0", "also-invalid"}
	latest := LatestVersion(versions)
	if latest != "2.0.0" {
		t.Errorf("expected '2.0.0', got %q", latest)
	}
}

func TestLatestVersion_Nil(t *testing.T) {
	latest := LatestVersion(nil)
	if latest != "" {
		t.Errorf("expected empty string, got %q", latest)
	}
}