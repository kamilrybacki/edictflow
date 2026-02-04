// agent/storage/queue.go
package storage

import (
	"time"

	"github.com/google/uuid"
)

type QueuedMessage struct {
	ID        int64
	RefID     string
	MsgType   string
	Payload   string
	CreatedAt time.Time
	Attempts  int
}

func (s *Storage) EnqueueMessage(msgType, payload string) (string, error) {
	refID := uuid.New().String()
	query := `INSERT INTO message_queue (ref_id, msg_type, payload, created_at, attempts) VALUES (?, ?, ?, ?, 0)`
	_, err := s.db.Exec(query, refID, msgType, payload, time.Now().Unix())
	return refID, err
}

func (s *Storage) GetPendingMessages() ([]QueuedMessage, error) {
	query := `SELECT id, ref_id, msg_type, payload, created_at, attempts FROM message_queue WHERE attempts < 3 ORDER BY created_at`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []QueuedMessage
	for rows.Next() {
		var m QueuedMessage
		var createdAt int64
		if err := rows.Scan(&m.ID, &m.RefID, &m.MsgType, &m.Payload, &createdAt, &m.Attempts); err != nil {
			return nil, err
		}
		m.CreatedAt = time.Unix(createdAt, 0)
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *Storage) IncrementAttempts(refID string) error {
	_, err := s.db.Exec("UPDATE message_queue SET attempts = attempts + 1 WHERE ref_id = ?", refID)
	return err
}

func (s *Storage) DeleteMessage(refID string) error {
	_, err := s.db.Exec("DELETE FROM message_queue WHERE ref_id = ?", refID)
	return err
}

func (s *Storage) DeleteOldMessages(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge).Unix()
	_, err := s.db.Exec("DELETE FROM message_queue WHERE created_at < ? OR attempts >= 3", cutoff)
	return err
}
