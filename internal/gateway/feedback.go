package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type FeedbackEntry struct {
	CreatedAt    time.Time `json:"created_at"`
	ArtifactPath string    `json:"artifact_path"`
	Kind         string    `json:"kind"`
	Comment      string    `json:"comment"`
	Href         string    `json:"href"`
	Title        string    `json:"title"`
	UserAgent    string    `json:"user_agent"`
}

type FeedbackStore struct {
	Dir string
}

func (s FeedbackStore) Append(entry FeedbackEntry) (string, error) {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return "", err
	}
	name := entry.CreatedAt.Format("2006-01-02") + "-feedback.jsonl"
	path := filepath.Join(s.Dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	encoded, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return "", err
	}
	return path, nil
}

func DefaultFeedbackDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "codex-artifact-gateway", "feedback"), nil
}
