// Author me@bytejedi.com

package watchdog

import (
	"os"
	"path"
	"runtime/pprof"
	"runtime/trace"
	"time"
	"treasure/core/zlog"
	"treasure/util"
)

var profiles struct {
	cpu       prof // cpu profile
	heap      prof // 内存 profile
	goroutine prof // 协程 profile
	trace     prof // trace profile
}

type prof struct {
	status bool     // 是否开启此profile
	f      *os.File // 此profile的文件句柄
}

type pprofConfig struct {
	Seconds              int    `json:"seconds" yaml:"seconds"`                               // 生成profile的时间周期
	CPUProfilePath       string `json:"cpu_profile_path" yaml:"cpu_profile_path"`             // cpu profile的存储目录
	HeapProfilePath      string `json:"heap_profile_path" yaml:"heap_profile_path"`           // heap profile的存储目录
	GoroutineProfilePath string `json:"goroutine_profile_path" yaml:"goroutine_profile_path"` // goroutine profile的存储目录
	TraceProfilePath     string `json:"trace_profile_path" yaml:"trace_profile_path"`         // trace profile的存储目录
}

// 获取profile的入口
func watchPprof() {
	counter := 0 // profile种类的计数器，如果计数器为0表示不开启profile
	// 创建cpu profile目录，makedirs内置了检测path是否为空的功能，如果为空，表示不开启此项profile，将返回false
	if ok, _ := util.Makedirs(DogConfig.PprofConfig.CPUProfilePath, 0755); ok {
		profiles.cpu.status = true
		counter++
		zlog.Info("watchdog::开启 CPU 监控...")
	}
	// 创建内存 profile目录，makedirs内置了检测path是否为空的功能，如果为空，表示不开启此项profile，将返回false
	if ok, _ := util.Makedirs(DogConfig.PprofConfig.HeapProfilePath, 0755); ok {
		profiles.heap.status = true
		counter++
		zlog.Info("watchdog::开启 内存 监控...")
	}
	// 创建协程 profile目录，makedirs内置了检测path是否为空的功能，如果为空，表示不开启此项profile，将返回false
	if ok, _ := util.Makedirs(DogConfig.PprofConfig.GoroutineProfilePath, 0755); ok {
		profiles.goroutine.status = true
		counter++
		zlog.Info("watchdog::开启 协程 监控...")
	}
	// 创建trace profile目录，makedirs内置了检测path是否为空的功能，如果为空，表示不开启此项profile，将返回false
	if ok, _ := util.Makedirs(DogConfig.PprofConfig.TraceProfilePath, 0755); ok {
		profiles.trace.status = true
		counter++
		zlog.Info("watchdog::开启 Trace 监控...")
	}
	if counter > 0 {
		duration := time.Second * time.Duration(DogConfig.PprofConfig.Seconds)
		if duration <= 0 {
			duration = 60 // 默认的生成profile的时间周期是60
		}
		zlog.Info("watchdog::watchdog服务已启动")
		for {
			startProfile()
			time.Sleep(duration)
			stopProfile()
		}
	}
}

// 开始profiling
func startProfile() {
	// 使用ISO8601标准格式化时间字符串作为文件名
	filename := time.Now().Format("2006-01-02T15:04:05.000Z0700") + ".prof"
	// cpu
	if profiles.cpu.status {
		f, err := os.OpenFile(path.Join(DogConfig.PprofConfig.CPUProfilePath, filename), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			zlog.Error(err)
		}
		profiles.cpu.f = f
		pprof.StartCPUProfile(profiles.cpu.f)
	}
	// 内存
	if profiles.heap.status {
		f, err := os.OpenFile(path.Join(DogConfig.PprofConfig.HeapProfilePath, filename), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			zlog.Error(err)
		}
		profiles.heap.f = f
	}
	// 协程
	if profiles.goroutine.status {
		f, err := os.OpenFile(path.Join(DogConfig.PprofConfig.GoroutineProfilePath, filename), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			zlog.Error(err)
		}
		profiles.goroutine.f = f
	}
	// trace
	if profiles.trace.status {
		f, err := os.OpenFile(path.Join(DogConfig.PprofConfig.TraceProfilePath, filename), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			zlog.Error(err)
		}
		profiles.trace.f = f
		trace.Start(profiles.trace.f)
	}
}

// 结束profiling
func stopProfile() {
	// cpu
	if profiles.cpu.f != nil {
		pprof.StopCPUProfile()
		profiles.cpu.f.Close()
	}
	// 内存
	if profiles.heap.f != nil {
		pprof.Lookup("heap").WriteTo(profiles.heap.f, 0)
		profiles.heap.f.Close()
	}
	// 协程
	if profiles.goroutine.f != nil {
		pprof.Lookup("goroutine").WriteTo(profiles.goroutine.f, 0)
		profiles.goroutine.f.Close()
	}
	// trace
	if profiles.trace.f != nil {
		trace.Stop()
		profiles.trace.f.Close()
	}
}
