// Copyright 2025 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package gowork provides helpers to read and modify go.work and go.mod files.
package gowork

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/modfile"
)

const (
	// GoWorkFileName is the name of the Go workspace file
	GoWorkFileName = "go.work"
	// GoModFileName is the name of the Go module file
	GoModFileName = "go.mod"
)

// GoVersion returns the version of the Go toolchain running this command,
// in a format suitable for a 'go' directive (e.g. "1.26.0").
func GoVersion() string {
	version := strings.TrimPrefix(runtime.Version(), "go")
	// Strip any suffix such as 'rc1', 'beta1' or 'devel ...'
	for i, c := range version {
		if (c < '0' || c > '9') && c != '.' {
			version = version[:i]
			break
		}
	}
	version = strings.TrimSuffix(version, ".")
	if version == "" {
		version = "1.22"
	}
	return version
}

// FindGoWork returns the path of the go.work file applicable to the given
// directory, looking up in the directory hierarchy.
// It returns an empty string if no go.work file could be found or if
// workspaces have been explicitly disabled with GOWORK=off.
func FindGoWork(dir string) string {
	switch gw := os.Getenv("GOWORK"); gw {
	case "":
	case "off":
		return ""
	default:
		return gw
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		gwPath := filepath.Join(dir, GoWorkFileName)
		if fi, err := os.Stat(gwPath); err == nil && !fi.IsDir() {
			return gwPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// FindModuleDir returns the root directory of the Go module the given
// directory belongs to, or an empty string if it could not be found.
func FindModuleDir(dir string) string {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, GoModFileName)); err == nil && !fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// OpenGoWork parses the go.work file applicable to the given directory.
// If no such file exists and create is true, then a new go.work file is
// created in the given directory.
// It returns the path of the go.work file and its parsed content.
func OpenGoWork(dir string, create bool) (string, *modfile.WorkFile, error) {
	gwPath := FindGoWork(dir)
	if gwPath == "" {
		if !create {
			return "", nil, nil
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return "", nil, err
		}
		gwPath = filepath.Join(absDir, GoWorkFileName)
		wf := new(modfile.WorkFile)
		wf.Syntax = new(modfile.FileSyntax)
		if err = wf.AddGoStmt(GoVersion()); err != nil {
			return "", nil, err
		}
		return gwPath, wf, nil
	}
	data, err := os.ReadFile(gwPath)
	if err != nil {
		return "", nil, err
	}
	wf, err := modfile.ParseWork(gwPath, data, nil)
	if err != nil {
		return "", nil, err
	}
	return gwPath, wf, nil
}

// SaveGoWork formats and writes the given go.work file at the given path.
func SaveGoWork(gwPath string, wf *modfile.WorkFile) error {
	wf.Cleanup()
	return os.WriteFile(gwPath, modfile.Format(wf.Syntax), 0644)
}

// UsePath returns the path to use in a 'use' directive of the go.work
// file located at gwPath for the module rooted at modDir.
func UsePath(gwPath, modDir string) string {
	absModDir, err := filepath.Abs(modDir)
	if err != nil {
		absModDir = modDir
	}
	relPath, err := filepath.Rel(filepath.Dir(gwPath), absModDir)
	if err != nil {
		return filepath.ToSlash(absModDir)
	}
	relPath = filepath.ToSlash(relPath)
	if !strings.HasPrefix(relPath, ".") {
		relPath = "./" + relPath
	}
	return relPath
}

// AddUse adds a 'use' directive for each of the given module
// directories to the go.work file applicable to dir.
// If no go.work file can be found, a new one is created in dir.
func AddUse(dir string, modDirs ...string) error {
	gwPath, wf, err := OpenGoWork(dir, true)
	if err != nil {
		return err
	}
	for _, modDir := range modDirs {
		if err = wf.AddUse(UsePath(gwPath, modDir), ""); err != nil {
			return fmt.Errorf("unable to add use directive for %s: %s", modDir, err)
		}
	}
	return SaveGoWork(gwPath, wf)
}

// DropUse removes the 'use' directive of each of the given module
// directories from the go.work file applicable to dir.
// It is a no-op if no go.work file can be found.
func DropUse(dir string, modDirs ...string) error {
	gwPath, wf, err := OpenGoWork(dir, false)
	if err != nil {
		return err
	}
	if wf == nil {
		return nil
	}
	for _, modDir := range modDirs {
		if err = wf.DropUse(UsePath(gwPath, modDir)); err != nil {
			return fmt.Errorf("unable to drop use directive for %s: %s", modDir, err)
		}
	}
	return SaveGoWork(gwPath, wf)
}
