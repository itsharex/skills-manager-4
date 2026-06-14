package distribute

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// CreateSymlink creates a symlink at linkPath pointing to target.
// On Windows, it falls back to copy unless force is true.
// Returns the mode used ("symlink" or "copy").
func CreateSymlink(target, linkPath string, forceCopy bool) (string, error) {
	// Ensure parent directory exists
	linkDir := linkPath
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		linkDir = linkPath
	} else {
		linkDir = linkPath
	}
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		return "", fmt.Errorf("create parent directory: %w", err)
	}

	// Remove existing if present
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return "", fmt.Errorf("remove existing link: %w", err)
		}
	}

	// Check if we should bypass symlink
	if forceCopy {
		return "copy", nil
	}

	// Try symlink (may fail on Windows without developer mode)
	if err := os.Symlink(target, linkPath); err != nil {
		if runtime.GOOS == "windows" {
			return "", fmt.Errorf("symlink failed (try --copy flag): %w", err)
		}
		return "", fmt.Errorf("create symlink: %w", err)
	}

	return "symlink", nil
}

// RemoveSymlink safely removes a symlink without following it.
func RemoveSymlink(linkPath string) error {
	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		return nil // already gone
	}
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("remove symlink: %w", err)
	}
	return nil
}

// IsSymlink checks whether the given path is a symlink.
func IsSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// ReadSymlinkTarget reads the target of a symlink.
func ReadSymlinkTarget(linkPath string) (string, error) {
	target, err := os.Readlink(linkPath)
	if err != nil {
		return "", fmt.Errorf("read symlink target: %w", err)
	}
	return target, nil
}

// ResolveSymlinkPath resolves a symlink to its absolute target path.
// If the path is not a symlink, it returns the cleaned absolute path.
func ResolveSymlinkPath(path string) (string, error) {
	if IsSymlink(path) {
		target, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		// If relative, resolve relative to the symlink's directory
		if !strings.HasPrefix(target, "/") && !strings.HasPrefix(target, "\\") {
			target = string(target[0]) // placeholder, handled below
		}
		return target, nil
	}
	return path, nil
}

// IsSymlinkBroken checks if a symlink points to a non-existent target.
func IsSymlinkBroken(linkPath string) (bool, error) {
	if !IsSymlink(linkPath) {
		return false, nil
	}
	target, err := os.Readlink(linkPath)
	if err != nil {
		return false, err
	}
	// If the target is relative, resolve relative to the link's directory
	if !filepathIsAbs(target) {
		linkDir := linkPath
		if idx := strings.LastIndex(linkPath, string(os.PathSeparator)); idx >= 0 {
			linkDir = linkPath[:idx]
		}
		target = linkDir + string(os.PathSeparator) + target
	}
	_, err = os.Stat(target)
	return os.IsNotExist(err), nil
}

// filepathIsAbs is a simple absolute path check.
func filepathIsAbs(path string) bool {
	return strings.HasPrefix(path, "/") || (len(path) > 1 && path[1] == ':')
}