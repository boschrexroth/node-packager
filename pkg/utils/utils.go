/*
 * SPDX-FileCopyrightText: Bosch Rexroth AG
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultFilePermissionDirectories is the default permisson for directories.
	DefaultFilePermissionDirectories = 0755

	// DefaultFilePermissionsFiles are the default file permission for files.
	DefaultFilePermissionsFiles = 0644
)

// Printfln prints a formated error, prefixed and postfixed by a new line.
func Errorfln(format string, a ...interface{}) {
	//fmt.Print(time.Now().Format(time.StampMilli) + " ") //time.RFC3339=ISO 8601
	fmt.Println()
	fmt.Printf("ERROR: "+format, a...)
	fmt.Println()
}

// Printfln prints a formated warning, prefixed and postfixed by a new line.
func Warnfln(format string, a ...interface{}) {
	fmt.Println()
	//fmt.Print(time.Now().Format(time.StampMilli) + " ") //time.RFC3339=ISO 8601
	fmt.Printf("WARNING: "+format, a...)
	fmt.Println()
}

// Printfln prints a formated information, prefixed and postfixed by a new line.
func Infofln(format string, a ...interface{}) {
	fmt.Println()
	//fmt.Print(time.Now().Format(time.StampMilli) + " ") //time.RFC3339=ISO 8601
	fmt.Printf("INFO: "+format, a...)
	fmt.Println()
}

// Printfln prints a formated message, followed by a new line.
func Printfln(format string, a ...interface{}) {
	//fmt.Print(time.Now().Format(time.StampMilli) + " ") //time.RFC3339=ISO 8601
	fmt.Printf(format, a...)
	fmt.Println()
}

// SetAllFilePermissions sets all file permissions (777).
func SetAllFilePermissions(path string) error {
	return os.Chmod(path, os.ModePerm)
}

// ReplaceFileExtension replaces the file extension on the given file path.
func ReplaceFileExtension(file, extension string) string {
	return strings.TrimSuffix(file, filepath.Ext(file)) + extension
}

// FileExists returns true if the file or folder exists, otherwise false.
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateFile creates the file. If the directory doesn't exist it will be created.
func CreateFile(path string, buffer []byte, perm os.FileMode) (err error) {
	dirpath := filepath.Dir(path)
	err = os.MkdirAll(dirpath, DefaultFilePermissionDirectories)
	if err != nil {
		return err
	}

	return os.WriteFile(path, buffer, perm)
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	if absSrc == absDst {
		return fmt.Errorf("%s and %s are the same file", src, dst)
	}

	in, err := os.Open(absSrc)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(absDst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	si, err := os.Stat(absSrc)
	if err != nil {
		return err
	}

	if err = os.Chmod(absDst, si.Mode()); err != nil {
		return err
	}

	return nil
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Existing files on dst will be removed.
// Symlinks are ignored and skipped.
func CopyDir(src, dst string) error {

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	if absSrc == absDst {
		return fmt.Errorf("%s and %s are the same directory", src, dst)
	}

	si, err := os.Stat(absSrc)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(absDst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		// clean the directory, if it already exists
		if err = CleanDir(absDst); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(absDst, si.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(absSrc)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		from := filepath.Join(absSrc, entry.Name())
		to := filepath.Join(absDst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(from, to); err != nil {
				return err
			}
		} else {
			// Skip symlinks
			info, infoErr := entry.Info()
			if infoErr != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				continue
			}

			if err := CopyFile(from, to); err != nil {
				return err
			}
		}
	}

	return nil
}

// CleanDir cleans the dir.
func CleanDir(dir string) error {

	// Remove
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	// Create
	return os.MkdirAll(dir, DefaultFilePermissionDirectories)
}

// HumanizedFileSize returns the human readable file size in the range of Bytes ... MB e.g. '10 MB'.
func HumanizedFileSize(file string) (string, error) {

	size, err := FileSize(file)
	if err != nil {
		return "unknown size", err
	}

	switch {
	case size < 1_024:
		// Bytes
		return fmt.Sprintf("%d Bytes", size), nil
	case size < 1048576:
		// KB
		return fmt.Sprintf("%.1f KB", float32(size)/1_024), nil
	default:
		// MB
		return fmt.Sprintf("%.1f MB", float32(size)/1_048_576), nil
	}
}

// FileSize returns the size of the gien file in bytes.
func FileSize(file string) (int64, error) {
	fi, err := os.Stat(file)

	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
