package syslogger

import (
	"log"
	"log/syslog"
	"os"
	"path/filepath"
)

var (
	emergLogger   *log.Logger = newLogger()
	alertLogger   *log.Logger = newLogger()
	critLogger    *log.Logger = newLogger()
	errLogger     *log.Logger = newLogger()
	warningLogger *log.Logger = newLogger()
	noticeLogger  *log.Logger = newLogger()
	infoLogger    *log.Logger = newLogger()
	debugLogger   *log.Logger = newLogger()
)

func newLogger() *log.Logger {
	return log.New(os.Stdout, "", 0)
}

func init() {
	prog := filepath.Base(os.Args[0])

	// level 0 emerg
	if logger, err := syslog.New(syslog.LOG_EMERG, prog); err == nil {
		emergLogger.SetOutput(logger)
	}

	// level 1 alert
	if logger, err := syslog.New(syslog.LOG_ALERT, prog); err == nil {
		alertLogger.SetOutput(logger)
	}

	// level 2 crit
	if logger, err := syslog.New(syslog.LOG_CRIT, prog); err == nil {
		critLogger.SetOutput(logger)
	}

	// level 3 err
	if logger, err := syslog.New(syslog.LOG_ERR, prog); err == nil {
		errLogger.SetOutput(logger)
	}

	// level 4 warning
	if logger, err := syslog.New(syslog.LOG_WARNING, prog); err == nil {
		warningLogger.SetOutput(logger)
	}

	// level 5 notice
	if logger, err := syslog.New(syslog.LOG_NOTICE, prog); err == nil {
		noticeLogger.SetOutput(logger)
	}

	// level 6 info
	if logger, err := syslog.New(syslog.LOG_INFO, prog); err == nil {
		infoLogger.SetOutput(logger)
	}

	// level 7 debug
	if logger, err := syslog.New(syslog.LOG_DEBUG, prog); err == nil {
		debugLogger.SetOutput(logger)
	}
}

func logErrors(logger *log.Logger, errs ...error) {
	args := make([]interface{}, len(errs))
	for index, value := range errs {
		args[index] = value
	}
	logger.Println(args...)
}

func Emerg(args ...interface{}) {
	emergLogger.Println(args...)
}

func EmergErrors(errs ...error) {
	logErrors(emergLogger, errs...)
}

func Alert(args ...interface{}) {
	alertLogger.Println(args...)
}

func AlertErrors(errs ...error) {
	logErrors(alertLogger, errs...)
}

func Crit(args ...interface{}) {
	critLogger.Println(args...)
}

func CritErrors(errs ...error) {
	logErrors(critLogger, errs...)
}

func Err(args ...interface{}) {
	errLogger.Println(args...)
}

func ErrErrors(errs ...error) {
	logErrors(errLogger, errs...)
}

func Warning(args ...interface{}) {
	warningLogger.Println(args...)
}

func WarningErrors(errs ...error) {
	logErrors(warningLogger, errs...)
}

func Notice(args ...interface{}) {
	noticeLogger.Println(args...)
}

func NoticeErrors(errs ...error) {
	logErrors(noticeLogger, errs...)
}

func Info(args ...interface{}) {
	infoLogger.Println(args...)
}

func InfoErrors(errs ...error) {
	logErrors(infoLogger, errs...)
}

func Debug(args ...interface{}) {
	debugLogger.Println(args...)
}

func DebugErrors(errs ...error) {
	logErrors(debugLogger, errs...)
}
