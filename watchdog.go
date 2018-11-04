// Author me@bytejedi.com
// 监控模块入口

package watchdog

var DogConfig *Config

type Config struct {
	PprofConfig *pprofConfig `json:"pprof" yaml:"pprof"` // pprof的配置
}

func Watch() {
	// pprof监控
	go watchPprof()
}
