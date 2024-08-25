package logger

import (
	"bytes"
	"fmt"
	"os"
	"serverGoChi/config"
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

func Init() {
	logCfg := config.GetLogConfig()

	Logger = &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &myFormatter{logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			ForceColors:            true,
			DisableLevelTruncation: true,
		},
		},
	}

	Logger.SetReportCaller(true)
	// Set Log Output to STDOUT
	Logger.SetOutput(os.Stdout)
	DbLogger = Logger
	// Set Log Level
	setLogLevel(Logger, logCfg.Level)
	setLogLevel(DbLogger, logCfg.DbLevel)
}

func setLogLevel(logger *logrus.Logger, level string) {
	switch strings.ToLower(level) {
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "trace":
		Logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}

// Println Function
func Println(level logLevel, label string, message interface{}) {
	// Make Sure Log Is Not Empty Variable
	if Logger != nil {
		// Set Service Name Log Information
		service := strings.ToLower(config.GetServerConfig().ServerName)

		// Print Log Based On Log Level Type
		switch level {
		case "panic":
			Logger.WithFields(logrus.Fields{
				"service": service,
				"label":   label,
			}).Panicln(message)
		case "fatal":
			Logger.WithFields(logrus.Fields{
				"service": service,
				"label":   label,
			}).Fatalln(message)
		case "error":
			Logger.WithFields(logrus.Fields{
				"service": service,
				"label":   label,
			}).Errorln(message)
		case "warn":
			Logger.WithFields(logrus.Fields{
				"service": service,
				"label":   label,
			}).Warnln(message)
		case "debug":
			Logger.WithFields(logrus.Fields{
				"service": service,
				"label":   label,
			}).Debug(message)
		case "trace":
			Logger.WithFields(logrus.Fields{
				"service": service,
				"label":   label,
			}).Traceln(message)
		default:
			Logger.WithFields(logrus.Fields{
				"service": service,
				"label":   label,
			}).Infoln(message)
		}
	}
}

type myFormatter struct {
	logrus.TextFormatter
}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// this whole mess of dealing with ansi color codes is required if you want the colored output otherwise you will lose colors in the log levels
	var levelColor int
	switch entry.Level {
	case logrus.TraceLevel:
		levelColor = 90 // blue
	case logrus.DebugLevel:
		levelColor = 97 // gray
	case logrus.WarnLevel:
		levelColor = 33 // yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31 // red
	default:
		levelColor = 36 // blue
	}
	fileName := ""
	line := 0
	if entry.HasCaller() {
		strList := strings.Split(entry.Caller.File, "/")
		if len(strList) > 0 {
			fileName = strList[len(strList)-1]
		}
		line = entry.Caller.Line
	}
	file := fmt.Sprintf("%s:%d", fileName, line)
	tag := ""
	t, ok := entry.Data["TAG"]
	if ok {
		tag, ok = t.(string)
		if ok {
			tag = fmt.Sprintf("[\x1b[%dm%-.4s\x1b[0m] ", 33, tag)
		}
	}
	buff := bytes.NewBuffer([]byte{})
	buff.WriteString("[")
	buff.WriteString(entry.Time.Format(f.TimestampFormat))
	buff.WriteString("]")
	buff.WriteString(fmt.Sprintf("[\x1b[%dm%-.4s\x1b[0m]", levelColor, strings.ToUpper(entry.Level.String())))
	buff.WriteString(fmt.Sprintf("[%-30s]", file))

	buff.WriteString(fmt.Sprintf("%-5s", tag))

	buff.WriteString(fmt.Sprintf(" %s\n", entry.Message))
	return buff.Bytes(), nil
}
