package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"serverGoChi/src/server"
	"strings"
	"time"
)

// Log Variable
var Logger *logrus.Logger

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

func init() {
	// Initialize Log as New Logrus Logger
	Logger = logrus.New()

	// Set Log Format to JSON Format
	Logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: false,
		TimestampFormat:  time.RFC3339Nano,
	})

	// Set Log Output to STDOUT
	Logger.SetOutput(os.Stdout)

	// Set Log Level
	switch strings.ToLower(server.Config.GetString("SERVER_LOG_LEVEL")) {
	case "panic":
		Logger.SetLevel(logrus.PanicLevel)
	case "fatal":
		Logger.SetLevel(logrus.FatalLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "trace":
		Logger.SetLevel(logrus.TraceLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}
}

// Println Function
func Println(level logLevel, label string, message interface{}) {
	// Make Sure Log Is Not Empty Variable
	if Logger != nil {
		// Set Service Name Log Information
		service := strings.ToLower(server.Config.GetString("SERVER_NAME"))

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
