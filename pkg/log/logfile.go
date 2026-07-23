package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/hb1707/ant-godmin/setting"
)

type DebugLevel int

const (
	DebugLevelFatal DebugLevel = iota
	DebugLevelError
	DebugLevelWarning
	DebugLevelInfo
	DebugLevelDebug
)

var (
	LPath  = ""
	Mode   = setting.App.RUNMODE
	Notice = func(er string, lev DebugLevel) {}
)

func Fatal(er interface{}, lev ...int) {
	var depth = 2
	if len(lev) > 0 {
		depth = lev[0]
	}
	log.SetPrefix("[ FATAL ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile(LPath+"error_"+time.Now().Format("200601")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	s := fmt.Sprint(er)
	_ = log.Output(depth, s+" [程序被终止]")
	funcName, _, line, ok := runtime.Caller(1)
	if ok {
		s = fmt.Sprintf("[ FATAL ] %s %s:%d %s", time.Now().Local().Format("2006/01/02 15:04:05"), runtime.FuncForPC(funcName).Name(), line, s)
	}
	go Notice(s, DebugLevelFatal)
	os.Exit(1)
}
func Error(er ...interface{}) {
	log.SetPrefix("[ ERROR ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	file, err := os.OpenFile(LPath+"error_"+time.Now().Format("200601")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	s := fmt.Sprint(er)
	_ = log.Output(2, s)
	funcName, _, line, ok := runtime.Caller(1)
	if ok {
		s = fmt.Sprintf("[ ERROR ] %s %s:%d %s", time.Now().Local().Format("2006/01/02 15:04:05"), runtime.FuncForPC(funcName).Name(), line, s)
	}
	go Notice(s, DebugLevelError)
}
func ErrorLev(lev int, er ...interface{}) {
	log.SetPrefix("[ ERROR ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	file, err := os.OpenFile(LPath+"error_"+time.Now().Format("200601")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	s := fmt.Sprint(er)
	_ = log.Output(lev+1, s)
	funcName, _, line, ok := runtime.Caller(lev)
	if ok {
		s = fmt.Sprintf("[ ERROR ] %s %s:%d %s", time.Now().Local().Format("2006/01/02 15:04:05"), runtime.FuncForPC(funcName).Name(), line, s)
	}
	go Notice(s, DebugLevelError)
}
func Warning(er ...interface{}) {

	log.SetPrefix("[ WARNING ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile(LPath+"warning_"+time.Now().Format("20060102")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))

	s := fmt.Sprint(er)
	_ = log.Output(2, s)
	funcName, _, line, ok := runtime.Caller(1)
	if ok {
		s = fmt.Sprintf("[ WARNING ] %s %s:%d %s", time.Now().Local().Format("2006/01/02 15:04:05"), runtime.FuncForPC(funcName).Name(), line, s)
	}
	go Notice(s, DebugLevelWarning)
}

func Info(er ...interface{}) {
	if setting.App.RUNMODE != "dev" {
		return
	}
	log.SetPrefix("[ INFO ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	file, err := os.OpenFile(LPath+"info_"+time.Now().Format("200601")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	s := fmt.Sprint(er)
	_ = log.Output(2, s)
}

func Debug(er ...interface{}) {
	log.SetPrefix("[ DEBUG ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile(LPath+"debug_"+time.Now().Format("20060102")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))

	s := fmt.Sprint(er)
	_ = log.Output(2, s)
	funcName, _, line, ok := runtime.Caller(1)
	if ok {
		s = fmt.Sprintf("[ DEBUG ] %s %s:%d %s", time.Now().Local().Format("2006/01/02 15:04:05"), runtime.FuncForPC(funcName).Name(), line, s)
	}
	go Notice(s, DebugLevelDebug)
}
