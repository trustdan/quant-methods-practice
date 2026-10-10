package bank

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// Bank provides query access to approved active question templates.
type Bank struct {
	mu        sync.RWMutex
	templates map[string]*domain.QuestionTemplate
	byModule  map[string][]*domain.QuestionTemplate
	byFamily  map[string][]*domain.QuestionTemplate
}

// NewBank creates a new empty bank instance.
func NewBank() *Bank {
	return &Bank{
		templates: make(map[string]*domain.QuestionTemplate),
		byModule:  make(map[string][]*domain.QuestionTemplate),
		byFamily:  make(map[string][]*domain.QuestionTemplate),
	}
}

// Add inserts a template into the bank.
func (b *Bank) Add(tmpl *domain.QuestionTemplate) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.removeLocked(tmpl.ID)
	b.templates[tmpl.ID] = tmpl
	b.byModule[tmpl.ModuleID] = append(b.byModule[tmpl.ModuleID], tmpl)
	b.byFamily[tmpl.FamilyID] = append(b.byFamily[tmpl.FamilyID], tmpl)
}

// Get retrieves a template by ID.
func (b *Bank) Get(id string) (*domain.QuestionTemplate, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	t, ok := b.templates[id]
	return t, ok
}

// Count returns the number of templates in the bank.
func (b *Bank) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.templates)
}

// List returns all templates in stable sorted order by ID.
func (b *Bank) List() []*domain.QuestionTemplate {
	b.mu.RLock()
	defer b.mu.RUnlock()

	list := make([]*domain.QuestionTemplate, 0, len(b.templates))
	for _, t := range b.templates {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

// ListByModule returns templates belonging to a module in stable sorted order.
func (b *Bank) ListByModule(moduleID string) []*domain.QuestionTemplate {
	b.mu.RLock()
	defer b.mu.RUnlock()

	src := b.byModule[moduleID]
	list := make([]*domain.QuestionTemplate, len(src))
	copy(list, src)
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

// ListByFamily returns templates belonging to a family in stable sorted order.
func (b *Bank) ListByFamily(familyID string) []*domain.QuestionTemplate {
	b.mu.RLock()
	defer b.mu.RUnlock()

	src := b.byFamily[familyID]
	list := make([]*domain.QuestionTemplate, len(src))
	copy(list, src)
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

// LoadActiveBank loads all approved templates from an approved curriculum directory.
// It fails if any template has unapproved/draft/retired status, unknown fields, or invalid schemas.
func LoadActiveBank(dir string, reg *Registry) (*Bank, error) {
	b := NewBank()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return b, nil
		}
		return nil, fmt.Errorf("reading active bank directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		tmpl, err := ValidateTemplateFile(filePath, reg)
		if err != nil {
			return nil, fmt.Errorf("failed to validate %s: %w", entry.Name(), err)
		}

		if tmpl.Status != domain.StatusActive {
			return nil, fmt.Errorf("active bank only admits approved active records; file %s has status %q", entry.Name(), tmpl.Status)
		}

		if _, exists := b.Get(tmpl.ID); exists {
			return nil, fmt.Errorf("duplicate template id %q in active bank (found in %s)", tmpl.ID, entry.Name())
		}

		b.Add(tmpl)
	}

	return b, nil
}

// ValidateTemplateFile reads and strictly validates a template file from disk.
func ValidateTemplateFile(path string, reg *Registry) (*domain.QuestionTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading template file %s: %w", path, err)
	}

	tmpl, err := StrictDecodeTemplate(data)
	if err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}

	if err := ValidateTemplate(tmpl, reg); err != nil {
		return nil, fmt.Errorf("validating %s: %w", path, err)
	}

	return tmpl, nil
}

// ValidationResult records the outcome of validating a single template file.
type ValidationResult struct {
	Path       string
	TemplateID string
	Status     domain.TemplateStatus
	Valid      bool
	Error      error
}

// ValidateBankDir scans a directory and validates all contained JSON template files.
func ValidateBankDir(dir string, reg *Registry) ([]ValidationResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading bank directory: %w", err)
	}

	var results []ValidationResult
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		tmpl, valErr := ValidateTemplateFile(filePath, reg)
		res := ValidationResult{
			Path: filePath,
		}
		if valErr != nil {
			res.Valid = false
			res.Error = valErr
		} else {
			res.Valid = true
			res.TemplateID = tmpl.ID
			res.Status = tmpl.Status
		}
		results = append(results, res)
	}

	return results, nil
}

// Remove excludes content from future selection. Existing session snapshots are separate.
func (b *Bank) Remove(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.removeLocked(id)
}
func (b *Bank) removeLocked(id string) {
	old, ok := b.templates[id]
	if !ok {
		return
	}
	delete(b.templates, id)
	filter := func(src []*domain.QuestionTemplate) []*domain.QuestionTemplate {
		dst := make([]*domain.QuestionTemplate, 0, len(src))
		for _, t := range src {
			if t.ID != id {
				dst = append(dst, t)
			}
		}
		return dst
	}
	b.byModule[old.ModuleID] = filter(b.byModule[old.ModuleID])
	b.byFamily[old.FamilyID] = filter(b.byFamily[old.FamilyID])
}
