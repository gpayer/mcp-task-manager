package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"gopkg.in/yaml.v3"
)

// MarkdownStorage handles reading/writing task markdown files
type MarkdownStorage struct {
	dir string
}

// NewMarkdownStorage creates a new markdown storage
func NewMarkdownStorage(dir string) *MarkdownStorage {
	return &MarkdownStorage{dir: dir}
}

// EnsureDir creates the tasks directory if it doesn't exist
func (s *MarkdownStorage) EnsureDir() error {
	return os.MkdirAll(s.dir, 0755)
}

// taskPath returns the file path for a task ID
func (s *MarkdownStorage) taskPath(id int) string {
	return filepath.Join(s.dir, fmt.Sprintf("%03d.md", id))
}

// Save writes a task to a markdown file
func (s *MarkdownStorage) Save(t *task.Task) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}

	// Build frontmatter
	frontmatter := struct {
		ID        int           `yaml:"id"`
		ParentID  *int          `yaml:"parent_id,omitempty"`
		Title     string        `yaml:"title"`
		Status    task.Status   `yaml:"status"`
		Priority  task.Priority `yaml:"priority"`
		Type      string        `yaml:"type"`
		CreatedAt string        `yaml:"created_at"`
		UpdatedAt string        `yaml:"updated_at"`
	}{
		ID:        t.ID,
		ParentID:  t.ParentID,
		Title:     t.Title,
		Status:    t.Status,
		Priority:  t.Priority,
		Type:      t.Type,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(frontmatter); err != nil {
		return err
	}
	buf.WriteString("---\n\n")
	buf.WriteString(t.Description)

	// Atomic write: write to temp, then rename
	tmpPath := s.taskPath(t.ID) + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.taskPath(t.ID))
}

// Load reads a task from a markdown file
func (s *MarkdownStorage) Load(id int) (*task.Task, error) {
	data, err := os.ReadFile(s.taskPath(id))
	if err != nil {
		return nil, err
	}
	return s.parse(data)
}

// Delete removes a task file
func (s *MarkdownStorage) Delete(id int) error {
	return os.Remove(s.taskPath(id))
}

// LoadAll reads all tasks from the directory
func (s *MarkdownStorage) LoadAll() ([]*task.Task, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []*task.Task
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Skip index file
		if entry.Name() == ".index.json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}

		t, err := s.parse(data)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// parse extracts task from markdown with frontmatter
func (s *MarkdownStorage) parse(data []byte) (*task.Task, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Find opening ---
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("invalid frontmatter: missing opening ---")
	}

	// Collect frontmatter
	var frontmatterBuf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		frontmatterBuf.WriteString(line)
		frontmatterBuf.WriteString("\n")
	}

	// Parse frontmatter
	var fm struct {
		ID        int    `yaml:"id"`
		ParentID  *int   `yaml:"parent_id"`
		Title     string `yaml:"title"`
		Status    string `yaml:"status"`
		Priority  string `yaml:"priority"`
		Type      string `yaml:"type"`
		CreatedAt string `yaml:"created_at"`
		UpdatedAt string `yaml:"updated_at"`
	}
	if err := yaml.Unmarshal(frontmatterBuf.Bytes(), &fm); err != nil {
		return nil, err
	}

	// Collect body (description)
	var bodyBuf bytes.Buffer
	for scanner.Scan() {
		if bodyBuf.Len() > 0 {
			bodyBuf.WriteString("\n")
		}
		bodyBuf.WriteString(scanner.Text())
	}

	// Parse timestamps
	createdAt, _ := parseTime(fm.CreatedAt)
	updatedAt, _ := parseTime(fm.UpdatedAt)

	return &task.Task{
		ID:          fm.ID,
		ParentID:    fm.ParentID,
		Title:       fm.Title,
		Description: strings.TrimSpace(bodyBuf.String()),
		Status:      task.Status(fm.Status),
		Priority:    task.Priority(fm.Priority),
		Type:        fm.Type,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// parseTime tries multiple time formats
func parseTime(s string) (t time.Time, err error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err = time.Parse(f, s); err == nil {
			return
		}
	}
	return
}

// NextID returns the next available task ID
func (s *MarkdownStorage) NextID() (int, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxID := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		if id, err := strconv.Atoi(name); err == nil && id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}
