package bridge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildInventory(root string) (Inventory, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Inventory{}, err
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return Inventory{}, fmt.Errorf("inventory root %s is not a directory", root)
	}
	inventory := Inventory{Root: filepath.ToSlash(root), RootREADMEExcluded: true}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}
				inventory.DescendantFolders++
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "README.md" {
			return nil
		}
		inventory.DescendantFiles++
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		switch {
		case strings.HasSuffix(path, ".go"):
			inventory.GoFiles++
			inventory.GoPhysicalLines += lineCount(raw)
		case strings.HasSuffix(path, ".gooo"):
			inventory.GoooFiles++
			inventory.GoooPhysicalLines += lineCount(raw)
		case strings.HasSuffix(path, ".tf") || strings.HasSuffix(path, ".tofu"):
			inventory.HCLFiles++
			inventory.HCLPhysicalLines += lineCount(raw)
		case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".json") && strings.Contains(strings.ToLower(filepath.Base(path)), "openapi"):
			inventory.OpenAPIFiles++
			inventory.OpenAPIPhysicalLines += lineCount(raw)
		}
		return nil
	})
	if err != nil {
		return Inventory{}, err
	}
	return inventory, nil
}
