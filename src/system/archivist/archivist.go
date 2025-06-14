package archivist

import (
	"fmt"
	"github.com/voodooEntity/go-cyberbrain/src/system/interfaces"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	LEVEL_DEBUG   = 1
	LEVEL_INFO    = 2
	LEVEL_WARNING = 3
	LEVEL_ERROR   = 4
	LEVEL_FATAL   = 5
)

type Archivist struct {
	logLevel map[string]int
	logFlags [5]bool
	logger   interfaces.LoggerInterface
}

type Config struct {
	Logger   interfaces.LoggerInterface
	LogLevel int
}

func New(conf *Config) *Archivist {
	// init archivist with default log flag set
	archivist := &Archivist{
		logFlags: [5]bool{false, true, true, true, true},
	}

	// in case no logger is given we gonne default
	// to logger to stdout
	archivist.SetLogger(conf.Logger)

	// set the provided loglevel
	archivist.SetLogLevel(conf.LogLevel)

	return archivist
}

func (a *Archivist) store(message string, stype string, dump bool, formatted bool, params []interface{}) {
	// dispatch the caller file+line number
	_, file, line, _ := runtime.Caller(2)
	arrPackagePath := strings.Split(file, "/")
	packageFile := arrPackagePath[len(arrPackagePath)-1]

	logLine := time.Now().Format("2006-01-02 15:04:05") + "|" + stype + "|" + packageFile + "#" + strconv.Itoa(line) + "|"
	if true == dump {
		if true == formatted {
			logLine = logLine + fmt.Sprintf(message, params...)
		} else {
			logLine = logLine + message + "|" + fmt.Sprintf("%+v", params)
		}
	} else {
		logLine = logLine + message
	}

	a.logger.Println(logLine)
}

func (a *Archivist) Error(message string, params ...interface{}) {
	if a.logFlags[LEVEL_ERROR-1] {
		if 0 == len(params) {
			a.store(message, "error", false, false, nil)
		} else {
			a.store(message, "error", true, false, params)
		}
	}
}

func (a *Archivist) ErrorF(message string, params ...interface{}) {
	if a.logFlags[LEVEL_ERROR-1] {
		a.store(message, "error", true, true, params)
	}
}

func (a *Archivist) Fatal(message string, params ...interface{}) {
	if a.logFlags[LEVEL_FATAL-1] {
		if 0 == len(params) {
			a.store(message, "fatal", false, false, nil)
		} else {
			a.store(message, "fatal", true, false, params)
		}
	}
}

func (a *Archivist) FatalF(message string, params ...interface{}) {
	if a.logFlags[LEVEL_FATAL-1] {
		a.store(message, "fatal", true, true, params)
	}
}

func (a *Archivist) Info(message string, params ...interface{}) {
	if a.logFlags[LEVEL_INFO-1] {
		if 0 == len(params) {
			a.store(message, "info", false, false, nil)
		} else {
			a.store(message, "info", true, false, params)
		}
	}
}

func (a *Archivist) InfoF(message string, params ...interface{}) {
	if a.logFlags[LEVEL_INFO-1] {
		a.store(message, "info", true, true, params)
	}
}

func (a *Archivist) Warning(message string, params ...interface{}) {
	if a.logFlags[LEVEL_WARNING-1] {
		if 0 == len(params) {
			a.store(message, "warning", false, false, nil)
		} else {
			a.store(message, "warning", true, false, params)
		}
	}
}

func (a *Archivist) WarningF(message string, params ...interface{}) {
	if a.logFlags[LEVEL_WARNING-1] {
		a.store(message, "warning", true, true, params)
	}
}

func (a *Archivist) Debug(message string, params ...interface{}) {
	if a.logFlags[LEVEL_DEBUG-1] {
		if 0 == len(params) {
			a.store(message, "debug", false, false, nil)
		} else {
			a.store(message, "debug", true, false, params)
		}
	}
}

func (a *Archivist) DebugF(message string, params ...interface{}) {
	if a.logFlags[LEVEL_DEBUG-1] {
		a.store(message, "debug", true, true, params)
	}
}

func (a *Archivist) SetLogLevel(logLevel int) {
	// check for non initialized log level first
	if 0 == logLevel {
		logLevel = LEVEL_WARNING
	}

	if logLevel >= LEVEL_DEBUG && logLevel <= LEVEL_FATAL {
		for index, _ := range a.logFlags {
			if logLevel-1 <= index {
				a.logFlags[index] = true
			} else {
				a.logFlags[index] = false
			}
		}
	} else {
		a.Error("Given LOG_LEVEL is unknown, defaulting to LEVEL_WARNING provided was: ", logLevel)
		a.SetLogLevel(LEVEL_WARNING)
	}
}

func (a *Archivist) SetLogger(logger interfaces.LoggerInterface) {
	// if logger is nil
	if nil == logger {
		logger = log.New(os.Stdout, "", 0)
	}
	//
	a.logger = logger
}
