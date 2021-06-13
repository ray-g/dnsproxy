package logger

import (
	"os"

	"github.com/op/go-logging"
)

var (
	logger    *logging.Logger
	debugMode bool
)

// InitLogger initialize logging instance
// It MUST be called and can only be called 1 time
func InitLogger(name string, debug bool) {
	var format string
	if debug {
		format = `%{color}[%{time:2006-01-02 15:04:05}][%{level:.4s}][%{id:04x}]%{color:reset} %{message} | {%{shortpkg}.%{longfunc}}`
	} else {
		format = `%{color}[%{time:2006-01-02 15:04:05}][%{level:.4s}][%{id:04x}]%{color:reset} %{message}`
	}

	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetBackend(logging.NewLogBackend(os.Stdout, "", 0))

	if debug {
		logging.SetLevel(logging.DEBUG, name)
	} else {
		logging.SetLevel(logging.INFO, name)
	}

	logger = logging.MustGetLogger(name)
	logger.ExtraCalldepth = 1
	debugMode = debug
}

// Logger return the logging instance
func Logger() *logging.Logger {
	return logger
}

// Info wraps logging.Info
func Info(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Infof wraps logging.Infof
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Notice wraps logging.Notice
func Notice(format string, args ...interface{}) {
	logger.Noticef(format, args...)
}

// Noticef wraps logging.Noticef
func Noticef(format string, args ...interface{}) {
	logger.Noticef(format, args...)
}

// Warning wraps logging.Warning
func Warning(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

// Warningf wraps logging.Warningf
func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

// Warn wraps logging.Warning
func Warn(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

// Warnf wraps logging.Warningf
func Warnf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

// Error wraps logging.Error
func Error(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Errorf wraps logging.Errorf
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Panic wraps logging.Panic
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Panicf wraps logging.Panicf
func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

// Fatal wraps logging.Fatal
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf wraps logging.Fatalf
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Debug wraps logging.Debug
func Debug(format string, args ...interface{}) {
	if !debugMode {
		return
	}
	logger.Debugf(format, args...)
}

// Debugf wraps logging.Debugf
func Debugf(format string, args ...interface{}) {
	if !debugMode {
		return
	}
	logger.Debugf(format, args...)
}

// Temp just a temp log...
func Temp(format string, args ...interface{}) {
	// if !debugMode {
	// 	return
	// }
	// logstr := fmt.Sprintf(format, args...)
	// logger.Debugf("[TEMP] " + logstr)
}
