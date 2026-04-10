package logger

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// Log Variable
var (
	Logger   *logrus.Logger
	DbLogger *logrus.Logger
)

// Log Level Data Type
type logLevel string

// Log Level Data Type Constant
const (
	LogLevelPanic logLevel = "panic"
	LogLevelFatal logLevel = "fatal"
	LogLevelError logLevel = "error"
	LogLevelWarn  logLevel = "warn"
	LogLevelDebug logLevel = "debug"
	LogLevelTrace logLevel = "trace"
	LogLevelInfo  logLevel = "info"
)

func Init(logLevel, dbLogLevel string) {
	Logger = &logrus.Logger{
		Out:          os.Stdout,
		Level:        logrus.DebugLevel,
		ReportCaller: true,
		Formatter:    &appFormatter{},
	}
	DbLogger = &logrus.Logger{
		Out:          os.Stdout,
		Level:        logrus.DebugLevel,
		ReportCaller: false,
		Formatter:    &appFormatter{},
	}
	setLogLevel(Logger, logLevel)
	setLogLevel(DbLogger, dbLogLevel)
}

func setLogLevel(l *logrus.Logger, level string) {
	switch strings.ToLower(level) {
	case "panic":
		l.SetLevel(logrus.PanicLevel)
	case "fatal":
		l.SetLevel(logrus.FatalLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	case "warn":
		l.SetLevel(logrus.WarnLevel)
	case "debug":
		l.SetLevel(logrus.DebugLevel)
	case "trace":
		l.SetLevel(logrus.TraceLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}
}

// Println is used by response helpers for HTTP-access structured logging.
func Println(level logLevel, label string, message interface{}) {
	if Logger == nil {
		return
	}
	entry := Logger.WithField("label", label)
	switch level {
	case LogLevelPanic:
		entry.Panic(message)
	case LogLevelFatal:
		entry.Fatal(message)
	case LogLevelError:
		entry.Error(message)
	case LogLevelWarn:
		entry.Warn(message)
	case LogLevelDebug:
		entry.Debug(message)
	case LogLevelTrace:
		entry.Trace(message)
	default:
		entry.Info(message)
	}
}

// ── Formatter ─────────────────────────────────────────────────────────────────

type appFormatter struct{}

func (f *appFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	const (
		reset  = "\x1b[0m"
		red    = "\x1b[31m"
		yellow = "\x1b[33m"
		cyan   = "\x1b[36m"
		gray   = "\x1b[90m"
		white  = "\x1b[97m"
	)

	// Level label + color
	var lvlColor, lvlLabel string
	switch entry.Level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		lvlColor, lvlLabel = red, "ERROR"
	case logrus.WarnLevel:
		lvlColor, lvlLabel = yellow, "WARN "
	case logrus.DebugLevel:
		lvlColor, lvlLabel = gray, "DEBUG"
	case logrus.TraceLevel:
		lvlColor, lvlLabel = gray, "TRACE"
	default:
		lvlColor, lvlLabel = cyan, "INFO "
	}

	// Caller: parent_dir/file.go:line  (e.g. handler/auth.go:46)
	caller := ""
	if entry.HasCaller() {
		parts := strings.Split(entry.Caller.File, "/")
		file := parts[len(parts)-1]
		dir := ""
		if len(parts) >= 2 {
			dir = parts[len(parts)-2] + "/"
		}
		caller = fmt.Sprintf("%s%s:%d", dir, file, entry.Caller.Line)
	}

	// Extra fields (skip internal TAG key), sorted for determinism
	var fieldParts []string
	for k, v := range entry.Data {
		if k == "TAG" {
			continue
		}
		fieldParts = append(fieldParts, fmt.Sprintf("%s%s=%s%q%s", gray, k, white, fmt.Sprintf("%v", v), reset))
	}
	sort.Strings(fieldParts)
	fields := ""
	if len(fieldParts) > 0 {
		fields = "  " + strings.Join(fieldParts, "  ")
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf,
		"%s%s%s  %s%s%s  %s%-40s%s  %s%s\n",
		gray, entry.Time.Format("2006-01-02 15:04:05.000"), reset,
		lvlColor, lvlLabel, reset,
		gray, caller, reset,
		entry.Message,
		fields,
	)
	return buf.Bytes(), nil
}
