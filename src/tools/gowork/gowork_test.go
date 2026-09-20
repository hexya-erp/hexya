// Copyright 2025 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package gowork

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/modfile"
)

// writeFile writes the given content in the given file, creating its
// directory if needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	assert.Nil(t, os.MkdirAll(filepath.Dir(path), 0755))
	assert.Nil(t, os.WriteFile(path, []byte(content), 0644))
}

func TestGoVersion(t *testing.T) {
	t.Run("Testing Go version format", func(t *testing.T) {
		version := GoVersion()
		assert.NotEmpty(t, version)
		assert.True(t, modfile.GoVersionRE.MatchString(version), "invalid go version: %s", version)
	})
}

func TestFindGoWork(t *testing.T) {
	t.Run("Testing go.work lookup in hierarchy", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		sub := filepath.Join(root, "sub", "subsub")
		assert.Nil(t, os.MkdirAll(sub, 0755))
		gwPath := filepath.Join(root, GoWorkFileName)
		writeFile(t, gwPath, "go 1.22\n\nuse .\n")
		assert.Equal(t, gwPath, FindGoWork(sub))
		assert.Equal(t, gwPath, FindGoWork(root))
	})
	t.Run("Testing go.work not found", func(t *testing.T) {
		t.Setenv("GOWORK", "off")
		root := t.TempDir()
		writeFile(t, filepath.Join(root, GoWorkFileName), "go 1.22\n")
		assert.Equal(t, "", FindGoWork(root))
	})
	t.Run("Testing GOWORK env variable", func(t *testing.T) {
		root := t.TempDir()
		gwPath := filepath.Join(root, "custom.work")
		t.Setenv("GOWORK", gwPath)
		assert.Equal(t, gwPath, FindGoWork(root))
	})
}

func TestFindModuleDir(t *testing.T) {
	t.Run("Testing module dir lookup", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "pkg", "sub")
		assert.Nil(t, os.MkdirAll(sub, 0755))
		writeFile(t, filepath.Join(root, GoModFileName), "module example.com/foo\n\ngo 1.22\n")
		res, err := filepath.EvalSymlinks(FindModuleDir(sub))
		assert.Nil(t, err)
		expected, err := filepath.EvalSymlinks(root)
		assert.Nil(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestUsePath(t *testing.T) {
	t.Run("Testing use paths computation", func(t *testing.T) {
		root := t.TempDir()
		gwPath := filepath.Join(root, GoWorkFileName)
		assert.Equal(t, ".", UsePath(gwPath, root))
		assert.Equal(t, "./pool", UsePath(gwPath, filepath.Join(root, "pool")))
		assert.Equal(t, "./a/b", UsePath(gwPath, filepath.Join(root, "a", "b")))
		assert.Equal(t, "..", UsePath(gwPath, filepath.Dir(root)))
	})
}

func TestOpenAndSaveGoWork(t *testing.T) {
	t.Run("Testing creation of a new go.work", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		gwPath, wf, err := OpenGoWork(root, true)
		assert.Nil(t, err)
		assert.Equal(t, filepath.Join(root, GoWorkFileName), gwPath)
		assert.NotNil(t, wf)
		assert.Equal(t, GoVersion(), wf.Go.Version)
		// Nothing written until we save
		_, err = os.Stat(gwPath)
		assert.True(t, os.IsNotExist(err))
		assert.Nil(t, SaveGoWork(gwPath, wf))
		data, err := os.ReadFile(gwPath)
		assert.Nil(t, err)
		assert.Contains(t, string(data), "go "+GoVersion())
	})
	t.Run("Testing no creation when create is false", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		gwPath, wf, err := OpenGoWork(root, false)
		assert.Nil(t, err)
		assert.Equal(t, "", gwPath)
		assert.Nil(t, wf)
	})
	t.Run("Testing parsing of an existing go.work", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		gwPath := filepath.Join(root, GoWorkFileName)
		writeFile(t, gwPath, "go 1.22\n\nuse ./foo\n")
		path, wf, err := OpenGoWork(root, false)
		assert.Nil(t, err)
		assert.Equal(t, gwPath, path)
		assert.Equal(t, 1, len(wf.Use))
		assert.Equal(t, "./foo", wf.Use[0].Path)
	})
	t.Run("Testing parsing of an invalid go.work", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		writeFile(t, filepath.Join(root, GoWorkFileName), "this is not a go.work file\n")
		_, _, err := OpenGoWork(root, false)
		assert.NotNil(t, err)
	})
}

func TestAddUse(t *testing.T) {
	t.Run("Testing adding use directives to a new go.work", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		poolDir := filepath.Join(root, "pool")
		assert.Nil(t, os.MkdirAll(poolDir, 0755))
		assert.Nil(t, AddUse(root, root, poolDir))
		_, wf, err := OpenGoWork(root, false)
		assert.Nil(t, err)
		var paths []string
		for _, u := range wf.Use {
			paths = append(paths, u.Path)
		}
		assert.Equal(t, []string{".", "./pool"}, paths)
	})
	t.Run("Testing adding an existing use directive is idempotent", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		assert.Nil(t, AddUse(root, root))
		assert.Nil(t, AddUse(root, root))
		_, wf, err := OpenGoWork(root, false)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(wf.Use))
	})
	t.Run("Testing updating a go.work of a parent directory", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		writeFile(t, filepath.Join(root, GoWorkFileName), "go 1.22\n\nuse .\n")
		sub := filepath.Join(root, "sub")
		assert.Nil(t, os.MkdirAll(sub, 0755))
		assert.Nil(t, AddUse(sub, sub))
		_, err := os.Stat(filepath.Join(sub, GoWorkFileName))
		assert.True(t, os.IsNotExist(err))
		_, wf, err := OpenGoWork(root, false)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(wf.Use))
		assert.Equal(t, "./sub", wf.Use[1].Path)
	})
}

func TestDropUse(t *testing.T) {
	t.Run("Testing dropping a use directive", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		poolDir := filepath.Join(root, "pool")
		assert.Nil(t, AddUse(root, root, poolDir))
		assert.Nil(t, DropUse(root, poolDir))
		_, wf, err := OpenGoWork(root, false)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(wf.Use))
		assert.Equal(t, ".", wf.Use[0].Path)
	})
	t.Run("Testing dropping an unknown use directive", func(t *testing.T) {
		t.Setenv("GOWORK", "")
		root := t.TempDir()
		assert.Nil(t, AddUse(root, root))
		assert.Nil(t, DropUse(root, filepath.Join(root, "unknown")))
		_, wf, err := OpenGoWork(root, false)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(wf.Use))
	})
	t.Run("Testing dropping without any go.work is a no-op", func(t *testing.T) {
		t.Setenv("GOWORK", "off")
		root := t.TempDir()
		assert.Nil(t, DropUse(root, root))
		_, err := os.Stat(filepath.Join(root, GoWorkFileName))
		assert.True(t, os.IsNotExist(err))
	})
}
