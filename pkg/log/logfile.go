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

// logSpec 固定一个日志级别的落盘与通知参数。
//
// fileLayout 是日志文件名里的时间格式：error/info 按月切（200601），
// warning/debug 按天切（20060102）。这是既有行为，不要顺手统一。
type logSpec struct {
	prefix     string
	fileTag    string
	fileLayout string
	lev        DebugLevel
	notify     bool // Info 不触发 Notice 回调
}

var (
	specFatal   = logSpec{"[ FATAL ]", "error", "200601", DebugLevelFatal, true}
	specError   = logSpec{"[ ERROR ]", "error", "200601", DebugLevelError, true}
	specWarning = logSpec{"[ WARNING ]", "warning", "20060102", DebugLevelWarning, true}
	specInfo    = logSpec{"[ INFO ]", "info", "200601", DebugLevelInfo, false}
	specDebug   = logSpec{"[ DEBUG ]", "debug", "20060102", DebugLevelDebug, true}
)

// emit 是所有日志函数的唯一出口。
//
// outputDepth / callerDepth 沿用各导出函数**原本直接传给 log.Output / runtime.Caller
// 的数值**，由 emit 内部各加 1 抵消自己这一层调用栈。这是本函数唯一的不变量。
//
// 传错的后果是日志里的 file:line 指向本文件而不是业务代码——不会报错，
// 日志文本也完全正常，只有定位能力悄悄没了。因此新增导出函数时，
// 必须与同级的非格式化版本传**相同**的深度，尤其不要写成「格式化版本调用非格式化版本」，
// 那会凭空多出一层栈。
//
// outputSuffix 只追加到落盘文本，不进 Notice 回调——这是 Fatal 的既有行为。
func emit(spec logSpec, outputDepth, callerDepth int, s, outputSuffix string) {
	log.SetPrefix(spec.prefix)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile(LPath+spec.fileTag+"_"+time.Now().Format(spec.fileLayout)+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	_ = log.Output(outputDepth+1, s+outputSuffix)
	if !spec.notify {
		return
	}
	funcName, _, line, ok := runtime.Caller(callerDepth + 1)
	if ok {
		s = fmt.Sprintf("%s %s %s:%d %s", spec.prefix, time.Now().Local().Format("2006/01/02 15:04:05"), runtime.FuncForPC(funcName).Name(), line, s)
	}
	go Notice(s, spec.lev)
}

// formatted 供格式化变体使用，复现 fmt.Sprint 作用在 []interface{} 上的既有格式
// （含外层方括号）。
//
// 目的是让 log.Warningf(f, a...) 与调用点自行 log.Warning(fmt.Sprintf(f, a...))
// 输出**逐字节一致**——两种写法会长期混在同一份日志文件里，格式不能有差别。
func formatted(format string, a ...interface{}) string {
	return fmt.Sprint([]interface{}{fmt.Sprintf(format, a...)})
}

// 注意 Fatal 的参数是单个 interface{} 而非变长切片，所以它的输出**没有**外层方括号，
// 与其他级别不同。Fatalf 保持同样的无方括号形态。
func Fatal(er interface{}, lev ...int) {
	var depth = 2
	if len(lev) > 0 {
		depth = lev[0]
	}
	emit(specFatal, depth, 1, fmt.Sprint(er), " [程序被终止]")
	os.Exit(1)
}

func Fatalf(format string, a ...interface{}) {
	emit(specFatal, 2, 1, fmt.Sprintf(format, a...), " [程序被终止]")
	os.Exit(1)
}

func Error(er ...interface{}) {
	emit(specError, 2, 1, fmt.Sprint(er), "")
}

func Errorf(format string, a ...interface{}) {
	emit(specError, 2, 1, formatted(format, a...), "")
}

func ErrorLev(lev int, er ...interface{}) {
	emit(specError, lev+1, lev, fmt.Sprint(er), "")
}

func Warning(er ...interface{}) {
	emit(specWarning, 2, 1, fmt.Sprint(er), "")
}

func Warningf(format string, a ...interface{}) {
	emit(specWarning, 2, 1, formatted(format, a...), "")
}

func Info(er ...interface{}) {
	if setting.App.RUNMODE != "dev" {
		return
	}
	emit(specInfo, 2, 1, fmt.Sprint(er), "")
}

func Infof(format string, a ...interface{}) {
	if setting.App.RUNMODE != "dev" {
		return
	}
	emit(specInfo, 2, 1, formatted(format, a...), "")
}

func Debug(er ...interface{}) {
	emit(specDebug, 2, 1, fmt.Sprint(er), "")
}

func Debugf(format string, a ...interface{}) {
	emit(specDebug, 2, 1, formatted(format, a...), "")
}
