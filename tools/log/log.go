package log

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// levels
const (
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
)

const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printErrorLevel   = "[error  ] "
	printFatalLevel   = "[fatal  ] "
)

type Logger struct {
	level      int
	baseLogger *log.Logger
	baseFile   *os.File

	//自定义新增 走chan记
	msgChanLen int64
	msgChan    chan string
	//signalChan chan string
	//wg         sync.WaitGroup
	async bool
	path  string
	date  string
	flag  int
}

func New(strLevel string, pathname string, flag int) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return nil, errors.New("unknown level: " + strLevel)
	}

	// logger
	var baseLogger *log.Logger
	var baseFile *os.File
	now := time.Now()
	if pathname != "" {

		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute())

		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return nil, err
		}

		baseLogger = log.New(file, "", flag)
		baseFile = file
	} else {
		baseLogger = log.New(os.Stdout, "", flag)
	}

	// new
	logger := new(Logger)
	logger.level = level
	logger.baseLogger = baseLogger
	logger.baseFile = baseFile

	//下面自定义新增 走chan
	logger.flag = flag

	if pathname != "" {
		logger.async = true
		logger.path = pathname
		logger.date = fmt.Sprintf("%d%02d%02d", now.Year(), now.Month(), now.Day())
	}

	if logger.async {
		//logger.signalChan = make(chan string, 1)
		logger.msgChanLen = 100
		logger.msgChan = make(chan string, logger.msgChanLen)

		go logger.Start()
	}

	return logger, nil
}

//自定义
func (logger *Logger) Start() {

	for {
		select {
		case msg := <-logger.msgChan:
			logger.baseLogger.Output(3, msg)
			//case <-logger.signalChan:
			//	for msg := range logger.msgChan {
			//		logger.baseLogger.Output(3, msg)
			//	}
			//	logger.wg.Done()
		}
	}

}

// It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close()
	}

	//自定义新增
	if logger.async {
		//close(logger.signalChan)
		//logger.wg.Wait()
		close(logger.msgChan)
	}

	logger.baseLogger = nil
	logger.baseFile = nil
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}

	format = printLevel + format
	if logger.async {
		now := time.Now()
		date := fmt.Sprintf("%d%02d%02d", now.Year(), now.Month(), now.Day())
		if logger.date != date {
			logger.date = date
			logger.baseFile.Close()
			filename := fmt.Sprintf("%d%02d%02d_%02d_%02d.log",
				now.Year(),
				now.Month(),
				now.Day(),
				now.Hour(),
				now.Minute())
			file, err := os.Create(path.Join(logger.path, filename))
			if err != nil {
				log.Println("err crate log file :", err)
			}
			logger.baseLogger = log.New(file, "", logger.flag)
			logger.baseFile = file
		}
		logger.msgChan <- fmt.Sprintf(format, a...)
	} else {
		//源生
		logger.baseLogger.Output(3, fmt.Sprintf(format, a...))
	}

	if level == fatalLevel {
		fmt.Println("log fatal level err, content : ", format)
		os.Exit(1)
	}
}

func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

var gLogger, _ = New("debug", "", log.LstdFlags)

// It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		gLogger = logger
	}
}

func Debug(format string, a ...interface{}) {
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func Release(format string, a ...interface{}) {
	gLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func Error(format string, a ...interface{}) {
	gLogger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func Fatal(format string, a ...interface{}) {
	gLogger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

func Close() {
	gLogger.Close()
}
