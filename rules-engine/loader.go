package rulesengine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

func loadRuleLibraryFromDir(root string, rulesetName string, rulesetVersion string) (*ast.KnowledgeLibrary, error) {
	lib := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(lib)
	grlFileCount := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() || strings.ToLower(filepath.Ext(path)) != ".grl" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		if err := ruleBuilder.BuildRuleFromResource(rulesetName, rulesetVersion, pkg.NewBytesResource(content)); err != nil {
			return fmt.Errorf("build %s: %w", path, err)
		}

		grlFileCount++
		return nil
	})
	if err != nil {
		return nil, err
	}

	if grlFileCount == 0 {
		return nil, fmt.Errorf("no .grl files found under %s", root)
	}

	return lib, nil
}
