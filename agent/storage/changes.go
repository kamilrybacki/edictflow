// agent/storage/changes.go
package storage

import "time"

type PendingChange struct {
	ID              string    `json:"id"`
	RuleID          string    `json:"rule_id"`
	FilePath        string    `json:"file_path"`
	OriginalContent string    `json:"original_content"`
	ModifiedContent string    `json:"modified_content"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

func (s *Storage) SavePendingChange(change PendingChange) error {
	query := `INSERT OR REPLACE INTO pending_changes (id, rule_id, file_path, original_content, modified_content, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, change.ID, change.RuleID, change.FilePath, change.OriginalContent, change.ModifiedContent, change.Status, change.CreatedAt.Unix())
	return err
}

func (s *Storage) GetPendingChange(id string) (PendingChange, error) {
	query := `SELECT id, rule_id, file_path, original_content, modified_content, status, created_at FROM pending_changes WHERE id = ?`
	var c PendingChange
	var createdAt int64
	err := s.db.QueryRow(query, id).Scan(&c.ID, &c.RuleID, &c.FilePath, &c.OriginalContent, &c.ModifiedContent, &c.Status, &createdAt)
	if err != nil {
		return PendingChange{}, err
	}
	c.CreatedAt = time.Unix(createdAt, 0)
	return c, nil
}

func (s *Storage) GetPendingChanges() ([]PendingChange, error) {
	query := `SELECT id, rule_id, file_path, original_content, modified_content, status, created_at FROM pending_changes ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []PendingChange
	for rows.Next() {
		var c PendingChange
		var createdAt int64
		if err := rows.Scan(&c.ID, &c.RuleID, &c.FilePath, &c.OriginalContent, &c.ModifiedContent, &c.Status, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt = time.Unix(createdAt, 0)
		changes = append(changes, c)
	}
	return changes, nil
}

func (s *Storage) UpdateChangeStatus(id, status string) error {
	_, err := s.db.Exec("UPDATE pending_changes SET status = ? WHERE id = ?", status, id)
	return err
}

func (s *Storage) DeleteChange(id string) error {
	_, err := s.db.Exec("DELETE FROM pending_changes WHERE id = ?", id)
	return err
}
