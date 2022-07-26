package log

import (
	"fmt"
	"github.com/hb1707/ant-godmin/setting"
	"io"
	"log"
	"os"
	"time"
)

var (
	LPath = ""
	Mode  = setting.App.RUNMODE
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
	_ = log.Output(2, s+"")
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
}

func Info(er ...interface{}) {
	if Mode != "dev" {
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
