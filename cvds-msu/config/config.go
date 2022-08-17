package config

import (
	"github.com/CVDS2020/CVDS2020/common/config"
	"github.com/CVDS2020/CVDS2020/cvds-msu/args"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Config struct {
	Http     Http              `yaml:"http" json:"http"`
	Storage  Storage           `yaml:"storage" json:"storage"`
	Log      Log               `yaml:"log" json:"log"`
	Service  Service           `yaml:"service" json:"service"`
	Figure   Figure            `yaml:"figure" json:"figure"`
	Localize map[string]string `yaml:"localize" json:"localize"`
	Version  string            `yaml:"version" json:"version"`
}

func (c *Config) PreHandle() config.PreHandlerConfig {
	if c == nil {
		c = new(Config)
	}
	c.Version = "v1.0.0"
	return c
}

var globalConfig unsafe.Pointer

func loadGlobalConfig() *Config {
	return (*Config)(atomic.LoadPointer(&globalConfig))
}

func storeGlobalConfig(c *Config) {
	atomic.StorePointer(&globalConfig, unsafe.Pointer(c))
}

var configInitializer sync.Once
var parser config.Parser

func initConfig() {
	file := args.GetArgsConfig().ConfigFile
	if file == "" {
		switch {
		case config.Exist("config.yaml"):
			parser.SetConfigFile("config.yaml", &config.TypeYaml)
		case config.Exist("config.yml"):
			parser.SetConfigFile("config.yml", &config.TypeYaml)
		case config.Exist("config.json"):
			parser.SetConfigFile("config.json", &config.TypeJson)
		case config.Exist("config"):
			parser.SetConfigFile("config", &config.TypeUnknown)
		}
	} else {
		parser.SetConfigFile(file, nil)
	}
	c := new(Config)
	if err := parser.Unmarshal(c); err != nil {
		println(err.Error())
		os.Exit(1)
	}
	storeGlobalConfig(c)
}

func ReloadConfig() {
	GlobalConfig()
	nc := new(Config)
	if err := parser.Unmarshal(nc); err != nil {
		println(err.Error())
		return
	}
	storeGlobalConfig(nc)
	configReloaded()
}

var (
	configReloadedCallbacks       []func()
	configReloadedCallbacksLocker sync.Mutex
)

func RegisterConfigReloadedCallback(callback func()) {
	configReloadedCallbacksLocker.Lock()
	configReloadedCallbacks = append(configReloadedCallbacks, callback)
	configReloadedCallbacksLocker.Unlock()
}

func configReloaded() {
	configReloadedCallbacksLocker.Lock()
	for _, callback := range configReloadedCallbacks {
		callback()
	}
	configReloadedCallbacksLocker.Unlock()
}

func GlobalConfig() *Config {
	if cp := loadGlobalConfig(); cp != nil {
		return cp
	}
	configInitializer.Do(initConfig)
	return loadGlobalConfig()
}

func HttpConfig() *Http {
	return &GlobalConfig().Http
}

func StorageConfig() *Storage {
	return &GlobalConfig().Storage
}

func LogConfig() *Log {
	return &GlobalConfig().Log
}

func ServiceConfig() *Service {
	return &GlobalConfig().Service
}

func FigureConfig() *Figure {
	return &GlobalConfig().Figure
}
