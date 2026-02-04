// agent/storage/projects.go
package storage

import (
	"encoding/json"
	"time"
)

type WatchedProject struct {
	Path            string    `json:"path"`
	DetectedContext []string  `json:"detected_context"`
	DetectedTags    []string  `json:"detected_tags"`
	LastSyncAt      time.Time `json:"last_sync_at"`
}

func (s *Storage) AddProject(path string) error {
	query := `INSERT OR REPLACE INTO watched_projects (path, last_sync_at) VALUES (?, ?)`
	_, err := s.db.Exec(query, path, time.Now().Unix())
	return err
}

func (s *Storage) RemoveProject(path string) error {
	_, err := s.db.Exec("DELETE FROM watched_projects WHERE path = ?", path)
	return err
}

func (s *Storage) GetProjects() ([]WatchedProject, error) {
	query := `SELECT path, detected_context, detected_tags, last_sync_at FROM watched_projects`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []WatchedProject
	for rows.Next() {
		var p WatchedProject
		var context, tags *string
		var lastSyncAt int64
		if err := rows.Scan(&p.Path, &context, &tags, &lastSyncAt); err != nil {
			return nil, err
		}
		if context != nil {
			json.Unmarshal([]byte(*context), &p.DetectedContext)
		}
		if tags != nil {
			json.Unmarshal([]byte(*tags), &p.DetectedTags)
		}
		p.LastSyncAt = time.Unix(lastSyncAt, 0)
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *Storage) UpdateProjectContext(path string, context, tags []string) error {
	contextJSON, _ := json.Marshal(context)
	tagsJSON, _ := json.Marshal(tags)
	query := `UPDATE watched_projects SET detected_context = ?, detected_tags = ?, last_sync_at = ? WHERE path = ?`
	_, err := s.db.Exec(query, string(contextJSON), string(tagsJSON), time.Now().Unix(), path)
	return err
}
