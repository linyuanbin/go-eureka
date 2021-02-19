package eureka

import (
	"fmt"
	"sync"
	"time"
)

var (
	log  Log
	once sync.Once
)

func init() {
	log = &LogImpl{}
}

func InitLog(l Log) {
	once.Do(func() {
		if l != nil {
			log = l
		}
	})

}

type Log interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

func Debug(args ...interface{}) {
	log.Debug(fmt.Sprint(args...))
}

func Debugf(format string, args ...interface{}) {
	log.Debug(fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	log.Info(fmt.Sprint(args...))
}

func Infof(format string, args ...interface{}) {
	log.Info(fmt.Sprintf(format, args...))
}

func Warn(args ...interface{}) {
	log.Warn(fmt.Sprint(args...))
}

func Warnf(format string, args ...interface{}) {
	log.Warn(fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	log.Error(fmt.Sprint(args...))
}

func Errorf(format string, args ...interface{}) {
	log.Error(fmt.Sprintf(format, args...))
}

//impl
type LogImpl struct{}

func (log *LogImpl) Sprint(level string, args ...interface{}) {
	msg := fmt.Sprintf(`time="%s" level=%s msg="%s"`,
		time.Now().Format("2006/01/02 15:04:05.000000"),
		level,
		fmt.Sprint(args...),
	)
	println(msg)
}

func (log *LogImpl) Debug(args ...interface{}) {
	log.Sprint("debug", args...)
}

func (log *LogImpl) Info(args ...interface{}) {
	log.Sprint("info", args...)
}

func (log *LogImpl) Warn(args ...interface{}) {
	log.Sprint("warn", args...)
}

func (log *LogImpl) Error(args ...interface{}) {
	log.Sprint("error", args...)
}
