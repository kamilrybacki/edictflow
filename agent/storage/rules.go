// agent/storage/rules.go
package storage

import (
	"encoding/json"
	"time"
)

type CachedRule struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	Content               string          `json:"content"`
	Description           string          `json:"description"`
	TargetLayer           string          `json:"target_layer"`
	CategoryID            string          `json:"category_id"`
	CategoryName          string          `json:"category_name"`
	Overridable           bool            `json:"overridable"`
	EffectiveStart        *int64          `json:"effective_start,omitempty"`
	EffectiveEnd          *int64          `json:"effective_end,omitempty"`
	Tags                  json.RawMessage `json:"tags"`
	Triggers              json.RawMessage `json:"triggers"`
	EnforcementMode       string          `json:"enforcement_mode"`
	TemporaryTimeoutHours int             `json:"temporary_timeout_hours"`
	Version               int             `json:"version"`
	CachedAt              time.Time       `json:"cached_at"`
}

type CachedCategory struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IsSystem     bool   `json:"is_system"`
	DisplayOrder int    `json:"display_order"`
}

func (s *Storage) SaveRules(rules []CachedRule, version int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Clear existing rules
	if _, err := tx.Exec("DELETE FROM cached_rules"); err != nil {
		return err
	}

	query := `INSERT INTO cached_rules (
		id, name, content, description, target_layer, category_id, category_name,
		overridable, effective_start, effective_end, tags, triggers,
		enforcement_mode, temporary_timeout_hours, version, cached_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, r := range rules {
		triggers, _ := json.Marshal(r.Triggers)
		tags, _ := json.Marshal(r.Tags)
		overridable := 0
		if r.Overridable {
			overridable = 1
		}
		if _, err := tx.Exec(query,
			r.ID, r.Name, r.Content, r.Description, r.TargetLayer, r.CategoryID, r.CategoryName,
			overridable, r.EffectiveStart, r.EffectiveEnd, string(tags), string(triggers),
			r.EnforcementMode, r.TemporaryTimeoutHours, version, time.Now().Unix(),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetRules() ([]CachedRule, error) {
	query := `SELECT id, name, content, description, target_layer, category_id, category_name,
		overridable, effective_start, effective_end, tags, triggers,
		enforcement_mode, temporary_timeout_hours, version, cached_at FROM cached_rules`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []CachedRule
	for rows.Next() {
		var r CachedRule
		var triggers, tags string
		var cachedAt int64
		var overridable int
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Content, &r.Description, &r.TargetLayer, &r.CategoryID, &r.CategoryName,
			&overridable, &r.EffectiveStart, &r.EffectiveEnd, &tags, &triggers,
			&r.EnforcementMode, &r.TemporaryTimeoutHours, &r.Version, &cachedAt,
		); err != nil {
			return nil, err
		}
		r.Triggers = json.RawMessage(triggers)
		r.Tags = json.RawMessage(tags)
		r.Overridable = overridable == 1
		r.CachedAt = time.Unix(cachedAt, 0)
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *Storage) GetRulesByLayer(targetLayer string) ([]CachedRule, error) {
	query := `SELECT id, name, content, description, target_layer, category_id, category_name,
		overridable, effective_start, effective_end, tags, triggers,
		enforcement_mode, temporary_timeout_hours, version, cached_at
		FROM cached_rules WHERE target_layer = ?`
	rows, err := s.db.Query(query, targetLayer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []CachedRule
	for rows.Next() {
		var r CachedRule
		var triggers, tags string
		var cachedAt int64
		var overridable int
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Content, &r.Description, &r.TargetLayer, &r.CategoryID, &r.CategoryName,
			&overridable, &r.EffectiveStart, &r.EffectiveEnd, &tags, &triggers,
			&r.EnforcementMode, &r.TemporaryTimeoutHours, &r.Version, &cachedAt,
		); err != nil {
			return nil, err
		}
		r.Triggers = json.RawMessage(triggers)
		r.Tags = json.RawMessage(tags)
		r.Overridable = overridable == 1
		r.CachedAt = time.Unix(cachedAt, 0)
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *Storage) GetRuleByID(id string) (CachedRule, error) {
	query := `SELECT id, name, content, description, target_layer, category_id, category_name,
		overridable, effective_start, effective_end, tags, triggers,
		enforcement_mode, temporary_timeout_hours, version, cached_at FROM cached_rules WHERE id = ?`
	var r CachedRule
	var triggers, tags string
	var cachedAt int64
	var overridable int
	err := s.db.QueryRow(query, id).Scan(
		&r.ID, &r.Name, &r.Content, &r.Description, &r.TargetLayer, &r.CategoryID, &r.CategoryName,
		&overridable, &r.EffectiveStart, &r.EffectiveEnd, &tags, &triggers,
		&r.EnforcementMode, &r.TemporaryTimeoutHours, &r.Version, &cachedAt,
	)
	if err != nil {
		return CachedRule{}, err
	}
	r.Triggers = json.RawMessage(triggers)
	r.Tags = json.RawMessage(tags)
	r.Overridable = overridable == 1
	r.CachedAt = time.Unix(cachedAt, 0)
	return r, nil
}

func (s *Storage) SaveCategories(categories []CachedCategory) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("DELETE FROM cached_categories"); err != nil {
		return err
	}

	query := `INSERT INTO cached_categories (id, name, is_system, display_order) VALUES (?, ?, ?, ?)`
	for _, c := range categories {
		isSystem := 0
		if c.IsSystem {
			isSystem = 1
		}
		if _, err := tx.Exec(query, c.ID, c.Name, isSystem, c.DisplayOrder); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetCategories() ([]CachedCategory, error) {
	query := `SELECT id, name, is_system, display_order FROM cached_categories ORDER BY display_order, name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []CachedCategory
	for rows.Next() {
		var c CachedCategory
		var isSystem int
		if err := rows.Scan(&c.ID, &c.Name, &isSystem, &c.DisplayOrder); err != nil {
			return nil, err
		}
		c.IsSystem = isSystem == 1
		categories = append(categories, c)
	}
	return categories, nil
}

func (s *Storage) GetCachedVersion() int {
	var version int
	_ = s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM cached_rules").Scan(&version)
	return version
}
