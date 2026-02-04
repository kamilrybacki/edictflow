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
	TargetLayer           string          `json:"target_layer"`
	Triggers              json.RawMessage `json:"triggers"`
	EnforcementMode       string          `json:"enforcement_mode"`
	TemporaryTimeoutHours int             `json:"temporary_timeout_hours"`
	Version               int             `json:"version"`
	CachedAt              time.Time       `json:"cached_at"`
}

func (s *Storage) SaveRules(rules []CachedRule, version int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing rules
	if _, err := tx.Exec("DELETE FROM cached_rules"); err != nil {
		return err
	}

	query := `INSERT INTO cached_rules (id, name, content, target_layer, triggers, enforcement_mode, temporary_timeout_hours, version, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, r := range rules {
		triggers, _ := json.Marshal(r.Triggers)
		if _, err := tx.Exec(query, r.ID, r.Name, r.Content, r.TargetLayer, string(triggers), r.EnforcementMode, r.TemporaryTimeoutHours, version, time.Now().Unix()); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetRules() ([]CachedRule, error) {
	query := `SELECT id, name, content, target_layer, triggers, enforcement_mode, temporary_timeout_hours, version, cached_at FROM cached_rules`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []CachedRule
	for rows.Next() {
		var r CachedRule
		var triggers string
		var cachedAt int64
		if err := rows.Scan(&r.ID, &r.Name, &r.Content, &r.TargetLayer, &triggers, &r.EnforcementMode, &r.TemporaryTimeoutHours, &r.Version, &cachedAt); err != nil {
			return nil, err
		}
		r.Triggers = json.RawMessage(triggers)
		r.CachedAt = time.Unix(cachedAt, 0)
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *Storage) GetRuleByID(id string) (CachedRule, error) {
	query := `SELECT id, name, content, target_layer, triggers, enforcement_mode, temporary_timeout_hours, version, cached_at FROM cached_rules WHERE id = ?`
	var r CachedRule
	var triggers string
	var cachedAt int64
	err := s.db.QueryRow(query, id).Scan(&r.ID, &r.Name, &r.Content, &r.TargetLayer, &triggers, &r.EnforcementMode, &r.TemporaryTimeoutHours, &r.Version, &cachedAt)
	if err != nil {
		return CachedRule{}, err
	}
	r.Triggers = json.RawMessage(triggers)
	r.CachedAt = time.Unix(cachedAt, 0)
	return r, nil
}

func (s *Storage) GetCachedVersion() int {
	var version int
	s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM cached_rules").Scan(&version)
	return version
}
