// Package config handles Runefile discovery and settings resolution.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFound is returned when no Runefile can be located by upward discovery.
var ErrNotFound = errors.New("no Runefile found in this directory or any parent")

// candidateNames are the accepted Runefile base names, matched case-insensitively.
var candidateNames = []string{"runefile", ".runefile"}

// Resolve picks the Runefile to use. If fileFlag is non-empty it is used
// directly (and must exist); otherwise discovery walks up from startDir.
func Resolve(fileFlag, startDir string) (string, error) {
	if fileFlag != "" {
		abs, err := filepath.Abs(fileFlag)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return "", errors.New("--file points at a directory, not a Runefile")
		}
		return abs, nil
	}
	return Discover(startDir)
}

// Discover walks startDir and its ancestors, returning the absolute path of the
// nearest Runefile / .runefile (case-insensitive). It returns ErrNotFound if
// none exists up to the filesystem root.
func Discover(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		if path, ok := runefileIn(dir); ok {
			return path, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// runefileIn returns the path of a Runefile in dir, if one exists.
func runefileIn(dir string) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		for _, cand := range candidateNames {
			if lower == cand {
				return filepath.Join(dir, e.Name()), true
			}
		}
	}
	return "", false
}
