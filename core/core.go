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
	dirRenames := make(map[string]string)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if containsIgnoreCase(skipDirs, d.Name()) {
				return filepath.SkipDir
			}

			if path != dir && strings.Contains(d.Name(), oldString) {
				newName := strings.ReplaceAll(d.Name(), oldString, newString)
				dirRenames[path] = filepath.Join(filepath.Dir(path), newName)
			}
		} else {
			*count++
			if *count > maxFiles {
				return fmt.Errorf("exceeded maximum number of files (%d)", maxFiles)
			}

			if err := processFile(path, oldString, newString); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	for oldPath, newPath := range dirRenames {
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return nil
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

	if strings.Contains(filepath.Base(path), oldString) {
		if err := renameFile(path, oldString, newString); err != nil {
			return err
		}
	}

	return nil
}

func renameFile(path, oldString, newString string) error {
	newName := strings.ReplaceAll(filepath.Base(path), oldString, newString)
	newPath := filepath.Join(filepath.Dir(path), newName)

	if err := os.Rename(path, newPath); err != nil {
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
