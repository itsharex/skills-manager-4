package operations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// OpLogger records operation logs to a JSONL file.
type OpLogger struct {
	mu   sync.Mutex
	path string
	file *os.File
}

var defaultOpLogger *OpLogger

// InitOpLogger initializes the global operation logger.
func InitOpLogger(poolPath string) error {
	metaDir := filepath.Join(poolPath, ".meta")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return err
	}
	logPath := filepath.Join(poolPath, "operations.log.jsonl")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defaultOpLogger = &OpLogger{path: logPath, file: f}
	return nil
}

// LogOp appends an operation log entry.
func LogOp(op, target, detail, source, storePath, agents string, success bool, errMsg string) {
	if defaultOpLogger == nil {
		return
	}
	entry := models.OpLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Operation: op,
		Target:    target,
		Detail:    detail,
		Source:    source,
		StorePath: storePath,
		Agents:    agents,
		Success:   success,
		Error:     errMsg,
	}
	data, _ := json.Marshal(entry)
	defaultOpLogger.mu.Lock()
	defer defaultOpLogger.mu.Unlock()
	defaultOpLogger.file.Write(append(data, '\n'))
}

// GetOpLogs reads the last N operation log entries.
func GetOpLogs(n int) []models.OpLog {
	if defaultOpLogger == nil {
		return nil
	}
	defaultOpLogger.mu.Lock()
	defer defaultOpLogger.mu.Unlock()

	data, err := os.ReadFile(defaultOpLogger.path)
	if err != nil {
		return nil
	}

	lines := splitLines(data)
	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	var logs []models.OpLog
	for _, line := range lines[start:] {
		if line == "" {
			continue
		}
		var entry models.OpLog
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		logs = append(logs, entry)
	}
	return logs
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := string(data[start:i])
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}
