package rulesengine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

type Runtime struct {
	lib            *ast.KnowledgeLibrary
	ruleSetName    string
	ruleSetVersion string
}

func NewRuntimeFromDir(root string, ruleSetName string, ruleSetVersion string) (*Runtime, error) {
	lib, err := loadRuleLibraryFromDir(root, ruleSetName, ruleSetVersion)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		lib:            lib,
		ruleSetName:    ruleSetName,
		ruleSetVersion: ruleSetVersion,
	}, nil
}

func (rt *Runtime) NewKnowledgeBaseInstance() (*ast.KnowledgeBase, error) {
	if rt == nil || rt.lib == nil {
		return nil, fmt.Errorf("runtime not initialized")
	}

	kb, err := rt.lib.NewKnowledgeBaseInstance(rt.ruleSetName, rt.ruleSetVersion)
	if err != nil {
		return nil, fmt.Errorf("new knowledge base instance: %w", err)
	}

	return kb, nil
}

func (rt *Runtime) ExecuteCustomerRules(customer *Customer) error {
	if customer == nil {
		return fmt.Errorf("customer is nil")
	}

	kb, err := rt.NewKnowledgeBaseInstance()
	if err != nil {
		return err
	}

	dataCtx := ast.NewDataContext()
	if err := dataCtx.Add("Customer", customer); err != nil {
		return fmt.Errorf("bind customer: %w", err)
	}

	gruleEngine := engine.NewGruleEngine()
	gruleEngine.MaxCycle = 5

	if err := gruleEngine.Execute(dataCtx, kb); err != nil {
		return fmt.Errorf("execute rules: %w", err)
	}

	return nil
}

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
