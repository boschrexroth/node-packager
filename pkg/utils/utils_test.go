/*
 * SPDX-FileCopyrightText: Bosch Rexroth AG
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */
package utils

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ()

func TestPrintfln(t *testing.T) {
	Printfln("%v is a number", 42)
}

func TestInfofln(t *testing.T) {
	Infofln("%v is a number", 42)
}

func TestWarnfln(t *testing.T) {
	Warnfln("%v is a number", 42)
}

func TestErrorfln(t *testing.T) {
	Errorfln("%v is a number", 42)
}

func TestCreateFile(t *testing.T) {
	// Arrange
	file := path.Join(t.TempDir(), "test")
	value := t.Name()

	// Act
	assert.NoError(t, CreateFile(file, []byte(value), os.ModePerm))

	// Assert
	assert.FileExists(t, file)
	bytes, err := os.ReadFile(file)
	assert.NoError(t, err)
	assert.Equal(t, value, string(bytes))
}

func TestCreateFileInvalidPath(t *testing.T) {
	// Arrange
	file := ""
	value := t.Name()

	// Act
	assert.Error(t, CreateFile(file, []byte(value), os.ModePerm))
}

func TestCopyFile(t *testing.T) {
	// Arrange
	value := t.Name()
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := path.Join(tmpDir, "B")

	assert.NoError(t, CreateFile(src, []byte(value), os.ModePerm))
	assert.FileExists(t, src)

	// Act
	assert.NoError(t, CopyFile(src, dst))

	// Assert
	srcBytes, err := os.ReadFile(src)
	assert.NoError(t, err)
	assert.Equal(t, value, string(srcBytes))

	assert.FileExists(t, dst)
	dstBytes, err := os.ReadFile(dst)
	assert.NoError(t, err)

	assert.EqualValues(t, srcBytes, dstBytes)
}

func TestCopyFileInvalidSrc(t *testing.T) {
	// Arrange
	src := ""
	dst := path.Join(t.TempDir(), "B")

	// Act
	assert.Error(t, CopyFile(src, dst))
}

func TestCopyFileInvalidDst(t *testing.T) {
	// Arrange
	value := t.Name()
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := ""

	assert.NoError(t, CreateFile(src, []byte(value), os.ModePerm))
	assert.FileExists(t, src)

	// Act
	assert.Error(t, CopyFile(src, dst))
}

func TestCopyFileSrcEqualsDst(t *testing.T) {
	// Arrange
	value := t.Name()
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := src

	assert.NoError(t, CreateFile(src, []byte(value), os.ModePerm))
	assert.FileExists(t, src)

	// Act
	assert.Error(t, CopyFile(src, dst))
}

func TestCopyDir(t *testing.T) {
	// Arrange
	value := t.Name()
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := path.Join(tmpDir, "B")

	// Create some dir structure in src dir
	file0 := "myfile"
	file1 := path.Join("folder1", "myfile")
	file2 := path.Join("folder1", "folder2", "myfile")
	file3 := path.Join("folder1", "folder2", "folder3", "myfile")

	assert.NoError(t, CreateFile(path.Join(src, file0), []byte(value+"0"), os.ModePerm))
	assert.NoError(t, CreateFile(path.Join(src, file1), []byte(value+"1"), os.ModePerm))
	assert.NoError(t, CreateFile(path.Join(src, file2), []byte(value+"2"), os.ModePerm))
	assert.NoError(t, CreateFile(path.Join(src, file3), []byte(value+"3"), os.ModePerm))
	assert.FileExists(t, path.Join(src, file0))
	assert.FileExists(t, path.Join(src, file1))
	assert.FileExists(t, path.Join(src, file2))
	assert.FileExists(t, path.Join(src, file3))

	// Act
	assert.NoError(t, CopyDir(src, dst))

	// Assert
	assert.DirExists(t, dst)
	assert.FileExists(t, path.Join(dst, file0))
	assert.FileExists(t, path.Join(dst, file1))
	assert.FileExists(t, path.Join(dst, file2))
	assert.FileExists(t, path.Join(dst, file3))

	// Assert
	srcBytes0, err := os.ReadFile(path.Join(src, file0))
	assert.NotEmpty(t, srcBytes0)
	assert.NoError(t, err)
	srcBytes1, err := os.ReadFile(path.Join(src, file1))
	assert.NotEmpty(t, srcBytes1)
	assert.NoError(t, err)
	srcBytes2, err := os.ReadFile(path.Join(src, file2))
	assert.NotEmpty(t, srcBytes2)
	assert.NoError(t, err)
	srcBytes3, err := os.ReadFile(path.Join(src, file3))
	assert.NotEmpty(t, srcBytes3)
	assert.NoError(t, err)

	dstBytes0, err := os.ReadFile(path.Join(dst, file0))
	assert.NoError(t, err)
	dstBytes1, err := os.ReadFile(path.Join(dst, file1))
	assert.NoError(t, err)
	dstBytes2, err := os.ReadFile(path.Join(dst, file2))
	assert.NoError(t, err)
	dstBytes3, err := os.ReadFile(path.Join(dst, file3))
	assert.NoError(t, err)

	assert.EqualValues(t, srcBytes0, dstBytes0)
	assert.EqualValues(t, srcBytes1, dstBytes1)
	assert.EqualValues(t, srcBytes2, dstBytes2)
	assert.EqualValues(t, srcBytes3, dstBytes3)
}

func TestCopyDirDstExists(t *testing.T) {
	// Arrange
	value := t.Name()
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := path.Join(tmpDir, "B")

	// Create a file in src and (existing) dst dir
	file := "myfile"
	assert.NoError(t, CreateFile(path.Join(src, file), []byte(value), os.ModePerm))
	assert.NoError(t, CreateFile(path.Join(dst, file), []byte(value), os.ModePerm))
	assert.FileExists(t, path.Join(src, file))
	assert.FileExists(t, path.Join(dst, file))

	// Act
	assert.NoError(t, CopyDir(src, dst))

	// Assert
	assert.DirExists(t, dst)
	assert.FileExists(t, path.Join(dst, file))

	// Assert
	srcBytes, err := os.ReadFile(path.Join(src, file))
	assert.NotEmpty(t, srcBytes)
	assert.NoError(t, err)

	dstBytes, err := os.ReadFile(path.Join(dst, file))
	assert.NoError(t, err)
	assert.EqualValues(t, srcBytes, dstBytes)
}

func TestCopyDirSrcNotaDir(t *testing.T) {
	// Arrange
	value := t.Name()
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := path.Join(tmpDir, "B")

	//Src is a file
	assert.NoError(t, CreateFile(src, []byte(value), os.ModePerm))
	assert.FileExists(t, src)

	// Act
	assert.Error(t, CopyDir(src, dst))
}

func TestCopyDirSrcNotExists(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := path.Join(tmpDir, "B")

	//Src is a file
	assert.NoDirExists(t, src)

	// Act
	assert.Error(t, CopyDir(src, dst))
}

func TestCopyDirSrcEqualsDst(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	src := path.Join(tmpDir, "A")
	dst := src

	// Act
	assert.Error(t, CopyDir(src, dst))
}

func TestCleanDir(t *testing.T) {
	// Arrange
	value := t.Name()
	dir := path.Join(t.TempDir(), t.Name())
	assert.NoError(t, os.MkdirAll(dir, DefaultFilePermissionDirectories))

	// Create the directory to be cleaned
	assert.DirExists(t, dir)

	file0 := "myfile"
	file1 := path.Join("folder1", "myfile")
	file2 := path.Join("folder1", "folder2", "myfile")
	file3 := path.Join("folder1", "folder2", "folder3", "myfile")

	assert.NoError(t, CreateFile(path.Join(dir, file0), []byte(value+"0"), os.ModePerm))
	assert.NoError(t, CreateFile(path.Join(dir, file1), []byte(value+"1"), os.ModePerm))
	assert.NoError(t, CreateFile(path.Join(dir, file2), []byte(value+"2"), os.ModePerm))
	assert.NoError(t, CreateFile(path.Join(dir, file3), []byte(value+"3"), os.ModePerm))
	assert.FileExists(t, path.Join(dir, file0))
	assert.FileExists(t, path.Join(dir, file1))
	assert.FileExists(t, path.Join(dir, file2))
	assert.FileExists(t, path.Join(dir, file3))

	// Act
	assert.NoError(t, CleanDir(dir))

	// Assert
	assert.DirExists(t, dir)
}

func TestFileExists(t *testing.T) {
	// Arrange
	file := path.Join(t.TempDir(), t.Name())
	assert.NoFileExists(t, file)

	// Act
	assert.False(t, FileExists(file))
	assert.NoError(t, CreateFile(file, []byte(t.Name()), os.ModePerm))
	assert.FileExists(t, file)

	// Assert
	assert.True(t, FileExists(file))
}

func TestSetAllFilePermissions(t *testing.T) {
	// Arrange
	file := path.Join(t.TempDir(), t.Name())
	assert.NoError(t, CreateFile(file, []byte(t.Name()), os.ModePerm))
	assert.FileExists(t, file)

	// Act
	assert.NoError(t, SetAllFilePermissions(file))

	// Assert
	fi, err := os.Stat(file)
	assert.NoError(t, err)
	assert.Equal(t, fi.Mode(), os.ModePerm)
}

func TestReplaceFileExtension(t *testing.T) {
	// Arrange
	ext := ".txt"
	file := path.Join(t.TempDir(), t.Name()+ext)
	assert.Equal(t, ".txt", filepath.Ext(file))

	// Act
	file = ReplaceFileExtension(file, ".png")

	// Assert
	assert.Equal(t, ".png", filepath.Ext(file))
}

func TestHumanizedFileSize(t *testing.T) {
	// Arrange
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	// Collect some dirs, having files of different size we can look for.
	dirsToRead := []string{cwd, os.ExpandEnv("$HOME")}

	for _, dir := range dirsToRead {
		entries, err := os.ReadDir(dir)
		assert.NoError(t, err)

		// Act
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			size, err := HumanizedFileSize(path.Join(dir, entry.Name()))
			assert.NoError(t, err)
			assert.NotEmpty(t, size)
		}
	}
}

func TestHumanizedFileSizeNotExists(t *testing.T) {
	// Arrange
	file := path.Join(t.TempDir(), t.Name())

	// Act
	size, err := HumanizedFileSize(file)
	assert.Error(t, err)
	assert.Equal(t, size, "unknown size")
}

func TestFileSize(t *testing.T) {
	// Arrange
	file := path.Join(t.TempDir(), t.Name())
	assert.NoError(t, CreateFile(file, []byte(t.Name()), os.ModePerm))

	// Act
	size, err := FileSize(file)
	assert.NoError(t, err)
	assert.Greater(t, size, int64(0))
}

func TestFileSizeNotExists(t *testing.T) {
	// Arrange
	file := path.Join(t.TempDir(), t.Name())

	// Act
	size, err := FileSize(file)
	assert.Error(t, err)
	assert.Equal(t, size, int64(0))
}
