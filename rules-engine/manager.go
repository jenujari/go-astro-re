package rulesengine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"
)

const (
	DefaultTenantID = "default"
)

type Config struct {
	RuleRootDir    string
	RuleSetName    string
	DefaultVersion string
}

type Manager struct {
	cfg Config

	mu      sync.RWMutex
	kbCache map[string]*ast.KnowledgeBase
	libByKV map[string]*ast.KnowledgeLibrary

	ruleEngine *engine.GruleEngine
	execMu     sync.Mutex
}

func NewManager(cfg Config) (*Manager, error) {
	if cfg.RuleRootDir == "" {
		return nil, fmt.Errorf("rule root dir required")
	}
	if cfg.RuleSetName == "" {
		return nil, fmt.Errorf("rule set name required")
	}
	if cfg.DefaultVersion == "" {
		return nil, fmt.Errorf("default version required")
	}

	mgr := &Manager{
		cfg:        cfg,
		kbCache:    make(map[string]*ast.KnowledgeBase),
		libByKV:    make(map[string]*ast.KnowledgeLibrary),
		ruleEngine: engine.NewGruleEngine(),
	}
	mgr.ruleEngine.MaxCycle = 5

	// Prime default tenant/version at startup.
	if _, err := mgr.KnowledgeBase(DefaultTenantID, cfg.DefaultVersion); err != nil {
		return nil, fmt.Errorf("prime default knowledge base: %w", err)
	}

	return mgr, nil
}

func (m *Manager) Execute(dataCtx ast.IDataContext, kb *ast.KnowledgeBase) error {
	if dataCtx == nil {
		return fmt.Errorf("data context is nil")
	}
	if kb == nil {
		return fmt.Errorf("knowledge base is nil")
	}

	// Shared engine instance guarded for concurrent requests.
	m.execMu.Lock()
	defer m.execMu.Unlock()

	if err := m.ruleEngine.Execute(dataCtx, kb); err != nil {
		return fmt.Errorf("execute rules: %w", err)
	}

	return nil
}

func (m *Manager) KnowledgeBase(tenantID string, version string) (*ast.KnowledgeBase, error) {
	tenantID = normalizeTenant(tenantID)
	if version == "" {
		version = m.cfg.DefaultVersion
	}

	cacheKey := keyFromTenantVersion(tenantID, version)

	m.mu.RLock()
	if kb, ok := m.kbCache[cacheKey]; ok {
		m.mu.RUnlock()
		return kb, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if kb, ok := m.kbCache[cacheKey]; ok {
		return kb, nil
	}

	lib, err := m.loadLibraryLocked(tenantID, version)
	if err != nil {
		return nil, err
	}

	kb, err := lib.NewKnowledgeBaseInstance(m.cfg.RuleSetName, version)
	if err != nil {
		return nil, fmt.Errorf("new knowledge base instance tenant=%s version=%s: %w", tenantID, version, err)
	}

	m.kbCache[cacheKey] = kb
	return kb, nil
}

func (m *Manager) StartWatcher(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("new watcher: %w", err)
	}

	if err := addRecursiveWatch(watcher, m.cfg.RuleRootDir); err != nil {
		watcher.Close()
		return fmt.Errorf("watch rules dir: %w", err)
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if isRuleChangeEvent(event) {
					m.invalidateCaches()
					if event.Has(fsnotify.Create) {
						if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
							_ = addRecursiveWatch(watcher, event.Name)
						}
					}
				}
			case <-watcher.Errors:
				// no-op: keep watcher alive
			}
		}
	}()

	return nil
}

func (m *Manager) invalidateCaches() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kbCache = make(map[string]*ast.KnowledgeBase)
	m.libByKV = make(map[string]*ast.KnowledgeLibrary)
}

func (m *Manager) loadLibraryLocked(tenantID string, version string) (*ast.KnowledgeLibrary, error) {
	libKey := keyFromTenantVersion(tenantID, version)
	if lib, ok := m.libByKV[libKey]; ok {
		return lib, nil
	}

	ruleDir, err := m.resolveRuleDir(tenantID, version)
	if err != nil {
		return nil, err
	}

	lib, err := loadRuleLibraryFromDir(ruleDir, m.cfg.RuleSetName, version)
	if err != nil {
		return nil, fmt.Errorf("load rules tenant=%s version=%s dir=%s: %w", tenantID, version, ruleDir, err)
	}

	m.libByKV[libKey] = lib
	return lib, nil
}

func (m *Manager) resolveRuleDir(tenantID string, version string) (string, error) {
	candidates := []string{
		filepath.Join(m.cfg.RuleRootDir, tenantID, version),
		filepath.Join(m.cfg.RuleRootDir, tenantID),
		filepath.Join(m.cfg.RuleRootDir, version),
		m.cfg.RuleRootDir,
	}

	for _, candidate := range candidates {
		if hasGRLFiles(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no rules found for tenant=%s version=%s under %s", tenantID, version, m.cfg.RuleRootDir)
}

func addRecursiveWatch(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		if err := w.Add(path); err != nil {
			return err
		}
		return nil
	})
}

func hasGRLFiles(root string) bool {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return false
	}

	errFound := errors.New("found grl")
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".grl" {
			return errFound
		}
		return nil
	})

	return errors.Is(err, errFound)
}

func isRuleChangeEvent(event fsnotify.Event) bool {
	if filepath.Ext(event.Name) != ".grl" {
		return false
	}
	return event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)
}

func keyFromTenantVersion(tenantID string, version string) string {
	return tenantID + "::" + version
}

func normalizeTenant(tenantID string) string {
	if tenantID == "" {
		return DefaultTenantID
	}
	return tenantID
}
