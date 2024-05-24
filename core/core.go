package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func walkDir(dir string, maxFiles int, oldString, newString string, skipDirs []string, count *int) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && containsIgnoreCase(skipDirs, d.Name()) {
			return filepath.SkipDir
		}

		if d.IsDir() {
			if strings.Contains(d.Name(), oldString) {
				newPath := filepath.Join(filepath.Dir(path), strings.ReplaceAll(d.Name(), oldString, newString))
				if err := renameDir(path, newPath); err != nil {
					return err
				}
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			*count++
			if *count > maxFiles {
				return fmt.Errorf("exceeded maximum number of files (%d)", maxFiles)
			}

			if err := processFile(path, oldString, newString); err != nil {
				return err
			}

			if strings.Contains(d.Name(), oldString) {
				newPath := filepath.Join(filepath.Dir(path), strings.ReplaceAll(d.Name(), oldString, newString))
				if err := renameFile(path, newPath); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func containsIgnoreCase(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func processFile(path, oldString, newString string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	newContent := strings.ReplaceAll(content, oldString, newString)

	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		return err
	}

	return nil
}

func renameFile(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func renameDir(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func Main(dir string, maxFiles int, replace string, skipDirs []string) {
	if replace == "" {
		fmt.Println("Error: --replace parameter is required")
		os.Exit(1)
	}

	parts := strings.Split(replace, "=")
	if len(parts) != 2 {
		fmt.Println("Error: --replace parameter should be in the format 'string1=string2'")
		os.Exit(1)
	}
	oldString, newString := parts[0], parts[1]

	if _, err := os.Stat(filepath.Join(dir, ".git")); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Error: directory is not under version control")
		os.Exit(1)
	}

	var count int
	err := walkDir(dir, maxFiles, oldString, newString, skipDirs, &count)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processed %d files\n", count)
}
