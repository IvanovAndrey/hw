package logger

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var levelNames = map[string]Level{
	"error": LevelError,
	"warn":  LevelWarn,
	"info":  LevelInfo,
	"debug": LevelDebug,
}

type Logger struct {
	appName    string
	appVersion string
	module     string
	level      Level
	cid        string
}

type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	App       string `json:"app"`
	Version   string `json:"version"`
	Module    string `json:"module"`
	CID       string `json:"cid"`
	Message   string `json:"message"`
}

func New(appName, appVersion, level string) *Logger {
	lvl, ok := levelNames[strings.ToLower(level)]
	if !ok {
		lvl = LevelInfo
	}
	return &Logger{
		appName:    appName,
		appVersion: appVersion,
		module:     "core",
		level:      lvl,
		cid:        uuid.New().String(),
	}
}

func (l Logger) WithModule(module string) Logger {
	l.module = module
	return l
}

func (l Logger) WithCid(cid string) Logger {
	l.cid = cid
	return l
}

func (l Logger) log(msgLevel Level, levelStr, msg string) {
	if msgLevel > l.level {
		return
	}

	entry := logEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     levelStr,
		App:       l.appName,
		Version:   l.appVersion,
		Module:    l.module,
		CID:       l.cid,
		Message:   msg,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to marshal log entry: " + err.Error() + "\n")
		return
	}

	_, _ = os.Stdout.Write(append(data, '\n'))
}

func (l Logger) Error(msg string) {
	l.log(LevelError, "error", msg)
}

func (l Logger) Warn(msg string) {
	l.log(LevelWarn, "warn", msg)
}

func (l Logger) Info(msg string) {
	l.log(LevelInfo, "info", msg)
}

func (l Logger) Debug(msg string) {
	l.log(LevelDebug, "debug", msg)
}
