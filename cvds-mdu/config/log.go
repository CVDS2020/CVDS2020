package config

import (
	"github.com/CVDS2020/CVDS2020/common/config"
	"github.com/CVDS2020/CVDS2020/common/def"
	"github.com/CVDS2020/CVDS2020/common/errors"
	"github.com/CVDS2020/CVDS2020/common/log"
)

const DefaultTimeLayout = "2006-01-02 15:04:05.999999999"

var (
	LogConfigNotFoundError  = errors.New("log config not found")
	InvalidLogEncodingError = errors.New("invalid log encoding")
)

func newTrue() *bool {
	b := true
	return &b
}

type JsonEncoder struct {
	// Set the keys used for each log entry. If any key is empty, that portion
	// of the entry is omitted.
	MessageKey      string `json:"message-key" yaml:"message-key"`
	LevelKey        string `json:"level-key" yaml:"level-key"`
	TimeKey         string `json:"time-key" yaml:"time-key"`
	NameKey         string `json:"name-key" yaml:"name-key"`
	CallerKey       string `json:"caller-key" yaml:"caller-key"`
	FunctionKey     string `json:"function-key" yaml:"function-key"`
	StacktraceKey   string `json:"stacktrace-key" yaml:"stacktrace-key"`
	SkipLineEndingP *bool  `json:"skip-line-ending" yaml:"skip-line-ending"`
	LineEnding      string `json:"line-ending" yaml:"line-ending"`
	EscapeESCP      *bool  `json:"escape-esc" yaml:"escape-esc"`
	// Configure the primitive representations of common complex types. For
	// example, some users may want all time.Times serialized as floating-point
	// seconds since epoch, while others may prefer ISO8601 strings.
	EncodeLevel    log.LevelEncoder    `json:"level-encoder" yaml:"level-encoder"`
	EncodeTime     log.TimeEncoder     `json:"time-encoder" yaml:"time-encoder"`
	EncodeDuration log.DurationEncoder `json:"duration-encoder" yaml:"duration-encoder"`
	EncodeCaller   log.CallerEncoder   `json:"caller-encoder" yaml:"caller-encoder"`
	// Unlike the other primitive type encoders, EncodeName is optional. The
	// zero value falls back to FullNameEncoder.
	EncodeName log.NameEncoder `json:"name-encoder" yaml:"name-encoder"`
}

func (e *JsonEncoder) PreHandle() config.PreHandlerConfig {
	if e == nil {
		e = new(JsonEncoder)
	}
	e.MessageKey = "msg"
	e.LevelKey = "level"
	e.TimeKey = "time"
	e.NameKey = "logger"
	e.CallerKey = "caller"
	e.FunctionKey = "func"

	e.EncodeLevel = log.CapitalLevelEncoder
	e.EncodeTime = log.TimeEncoderOfLayout(DefaultTimeLayout)
	e.EncodeDuration = log.StringDurationEncoder
	e.EncodeCaller = log.ShortCallerEncoder
	e.EncodeName = log.FullNameEncoder
	return e
}

func (e *JsonEncoder) SkipLineEnding() bool {
	if e.SkipLineEndingP == nil {
		return false
	}
	return *e.SkipLineEndingP
}

func (e *JsonEncoder) EscapeESC() bool {
	if e.EscapeESCP == nil {
		return false
	}
	return *e.EscapeESCP
}

type ConsoleEncoder struct {
	// Set the keys used for each log entry. If any key is empty, that portion
	// of the entry is omitted.
	DisableLevelP      *bool  `json:"disable-level" yaml:"disable-level"`
	DisableTimeP       *bool  `json:"disable-time" yaml:"disable-time"`
	DisableNameP       *bool  `json:"disable-name" yaml:"disable-name"`
	DisableCallerP     *bool  `json:"disable-caller" yaml:"disable-caller"`
	DisableFunctionP   *bool  `json:"disable-function" yaml:"disable-function"`
	DisableStacktraceP *bool  `json:"disable-stacktrace" yaml:"disable-stacktrace"`
	SkipLineEndingP    *bool  `json:"skip-line-ending" yaml:"skip-line-ending"`
	LineEnding         string `json:"line-ending" yaml:"line-ending"`
	// Configure the primitive representations of common complex types. For
	// example, some users may want all time.Times serialized as floating-point
	// seconds since epoch, while others may prefer ISO8601 strings.
	EncodeLevel    log.LevelEncoder    `json:"level-encoder" yaml:"level-encoder"`
	EncodeTime     log.TimeEncoder     `json:"time-encoder" yaml:"time-encoder"`
	EncodeDuration log.DurationEncoder `json:"duration-encoder" yaml:"duration-encoder"`
	EncodeCaller   log.CallerEncoder   `json:"caller-encoder" yaml:"caller-encoder"`
	// Unlike the other primitive type encoders, EncodeName is optional. The
	// zero value falls back to FullNameEncoder.
	EncodeName log.NameEncoder `json:"name-encoder" yaml:"name-encoder"`
	// Configures the field separator used by the console encoder. Defaults
	// to tab.
	ConsoleSeparator string `json:"console-separator" yaml:"console-separator"`
}

func (e *ConsoleEncoder) PreHandle() config.PreHandlerConfig {
	if e == nil {
		e = new(ConsoleEncoder)
	}
	e.DisableStacktraceP = newTrue()

	e.EncodeLevel = log.CapitalLevelEncoder
	e.EncodeTime = log.TimeEncoderOfLayout(DefaultTimeLayout)
	e.EncodeDuration = log.StringDurationEncoder
	e.EncodeCaller = log.ShortCallerEncoder
	e.EncodeName = log.FullNameEncoder
	return e
}

func (e *ConsoleEncoder) DisableLevel() bool {
	if e.DisableLevelP == nil {
		return false
	}
	return *e.DisableLevelP
}

func (e *ConsoleEncoder) DisableTime() bool {
	if e.DisableTimeP == nil {
		return false
	}
	return *e.DisableTimeP
}

func (e *ConsoleEncoder) DisableName() bool {
	if e.DisableNameP == nil {
		return false
	}
	return *e.DisableNameP
}

func (e *ConsoleEncoder) DisableCaller() bool {
	if e.DisableCallerP == nil {
		return false
	}
	return *e.DisableCallerP
}

func (e *ConsoleEncoder) DisableFunction() bool {
	if e.DisableFunctionP == nil {
		return false
	}
	return *e.DisableFunctionP
}

func (e *ConsoleEncoder) DisableStacktrace() bool {
	if e.DisableStacktraceP == nil {
		return false
	}
	return *e.DisableStacktraceP
}

func (e *ConsoleEncoder) SkipLineEnding() bool {
	if e.SkipLineEndingP == nil {
		return false
	}
	return *e.SkipLineEndingP
}

type LoggerConfig struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling Config.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config.
	Level log.AtomicLevel `json:"level" yaml:"level"`
	// DevelopmentP puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	DevelopmentP *bool `json:"development" yaml:"development"`
	// DisableCallerP stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCallerP *bool `json:"disable-caller" yaml:"disable-caller"`
	// DisableStacktraceP completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktraceP *bool `json:"disable-stacktrace" yaml:"disable-stacktrace"`
	// Sampling sets a sampling policy. A nil SamplingConfig disables sampling.
	Sampling *log.SamplingConfig `json:"sampling" yaml:"sampling"`
	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `json:"encoding" yaml:"encoding"`
	// OutputPaths is a list of URLs or file paths to write logging output to.
	// See Open for details.
	OutputPaths []string `json:"output-paths" yaml:"output-paths"`
	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	// The default is standard error.
	//
	// Note that this setting only affects internal errors; for sample code that
	// sends error-level logs to a different location from info- and debug-level
	// logs, see the package-level AdvancedConfiguration example.
	ErrorOutputPaths []string `json:"error-output-paths" yaml:"error-output-paths"`
	// InitialFields is a collection of fields to add to the root logger.
	InitialFields map[string]interface{} `json:"initial-fields" yaml:"initial-fields"`

	ConsoleEncoder ConsoleEncoder `yaml:"console-encoder" json:"console-encoder"`
	JsonEncoder    JsonEncoder    `yaml:"json-encoder" json:"json-encoder"`
}

func (c *LoggerConfig) PreHandle() config.PreHandlerConfig {
	if c == nil {
		c = new(LoggerConfig)
	}
	c.Level = log.NewAtomicLevelAt(log.InfoLevel)
	c.DisableStacktraceP = newTrue()
	c.Encoding = "json"
	c.OutputPaths = []string{"stdout"}
	c.ErrorOutputPaths = []string{"stderr"}
	return c
}

func (c *LoggerConfig) Development() bool {
	if c.DevelopmentP == nil {
		return false
	}
	return *c.DevelopmentP
}

func (c *LoggerConfig) DisableCaller() bool {
	if c.DisableCallerP == nil {
		return false
	}
	return *c.DisableCallerP
}

func (c *LoggerConfig) DisableStacktrace() bool {
	if c.DisableStacktraceP == nil {
		return false
	}
	return *c.DisableStacktraceP
}

type Log struct {
	LoggerConfig `yaml:",inline"`
	Configs      map[string]*LoggerConfig `yaml:",inline" json:"configs"`
	Modules      map[string]string        `yaml:"modules" json:"modules"`
}

func (l *Log) PreHandle() config.PreHandlerConfig {
	if l == nil {
		l = new(Log)
	}
	l.Configs = make(map[string]*LoggerConfig)
	l.Modules = make(map[string]string)
	return l
}

func (l *Log) PostHandle() (config.PostHandlerConfig, error) {
	if l.Encoding != "json" && l.Encoding != "console" {
		return nil, InvalidLogEncodingError
	}

	for _, logConfig := range l.Configs {
		// post handle custom logger config
		def.SetDefault(&logConfig.Level, l.Level)
		def.SetDefault(&logConfig.DevelopmentP, l.DevelopmentP)
		def.SetDefault(&logConfig.DisableCallerP, l.DisableCallerP)
		def.SetDefault(&logConfig.DisableStacktraceP, l.DisableStacktraceP)
		def.SetDefault(&logConfig.Sampling, l.Sampling)
		def.SetDefault(&logConfig.Encoding, l.Encoding)
		if len(logConfig.OutputPaths) == 0 {
			logConfig.OutputPaths = l.OutputPaths
		}
		if len(logConfig.ErrorOutputPaths) == 0 {
			logConfig.ErrorOutputPaths = l.ErrorOutputPaths
		}
		if len(logConfig.InitialFields) == 0 && len(l.InitialFields) != 0 {
			if logConfig.InitialFields == nil {
				logConfig.InitialFields = make(map[string]interface{}, len(l.InitialFields))
			}
			for key, value := range l.InitialFields {
				logConfig.InitialFields[key] = value
			}
		}

		// post handle custom logger json encoder config
		jsonEncoder := &logConfig.JsonEncoder
		def.SetDefault(&jsonEncoder.MessageKey, l.JsonEncoder.MessageKey)
		def.SetDefault(&jsonEncoder.TimeKey, l.JsonEncoder.TimeKey)
		def.SetDefault(&jsonEncoder.NameKey, l.JsonEncoder.NameKey)
		def.SetDefault(&jsonEncoder.CallerKey, l.JsonEncoder.CallerKey)
		def.SetDefault(&jsonEncoder.FunctionKey, l.JsonEncoder.FunctionKey)
		def.SetDefault(&jsonEncoder.StacktraceKey, l.JsonEncoder.StacktraceKey)
		def.SetDefault(&jsonEncoder.LineEnding, l.JsonEncoder.LineEnding)
		def.SetDefault(&jsonEncoder.SkipLineEndingP, l.JsonEncoder.SkipLineEndingP)

		if jsonEncoder.EncodeLevel == nil {
			jsonEncoder.EncodeLevel = l.JsonEncoder.EncodeLevel
		}
		if jsonEncoder.EncodeTime == nil {
			jsonEncoder.EncodeTime = l.JsonEncoder.EncodeTime
		}
		if jsonEncoder.EncodeDuration == nil {
			jsonEncoder.EncodeDuration = l.JsonEncoder.EncodeDuration
		}
		if jsonEncoder.EncodeCaller == nil {
			jsonEncoder.EncodeCaller = l.JsonEncoder.EncodeCaller
		}
		if jsonEncoder.EncodeName == nil {
			jsonEncoder.EncodeName = l.JsonEncoder.EncodeName
		}

		// post handle custom logger console encoder config
		consoleEncoder := &logConfig.ConsoleEncoder
		def.SetDefault(&consoleEncoder.DisableLevelP, l.ConsoleEncoder.DisableLevelP)
		def.SetDefault(&consoleEncoder.DisableTimeP, l.ConsoleEncoder.DisableTimeP)
		def.SetDefault(&consoleEncoder.DisableNameP, l.ConsoleEncoder.DisableNameP)
		def.SetDefault(&consoleEncoder.DisableCallerP, l.ConsoleEncoder.DisableCallerP)
		def.SetDefault(&consoleEncoder.DisableFunctionP, l.ConsoleEncoder.DisableFunctionP)
		def.SetDefault(&consoleEncoder.DisableStacktraceP, l.ConsoleEncoder.DisableStacktraceP)
		def.SetDefault(&consoleEncoder.SkipLineEndingP, l.ConsoleEncoder.SkipLineEndingP)

		if consoleEncoder.EncodeLevel == nil {
			consoleEncoder.EncodeLevel = l.ConsoleEncoder.EncodeLevel
		}
		if consoleEncoder.EncodeTime == nil {
			consoleEncoder.EncodeTime = l.ConsoleEncoder.EncodeTime
		}
		if consoleEncoder.EncodeDuration == nil {
			consoleEncoder.EncodeDuration = l.ConsoleEncoder.EncodeDuration
		}
		if consoleEncoder.EncodeCaller == nil {
			consoleEncoder.EncodeCaller = l.ConsoleEncoder.EncodeCaller
		}
		if consoleEncoder.EncodeName == nil {
			consoleEncoder.EncodeName = l.ConsoleEncoder.EncodeName
		}
	}

	// check module logger config
	for _, conf := range l.Modules {
		if _, has := l.Configs[conf]; !has {
			if conf != "default" {
				return nil, InvalidLogEncodingError
			}
		}
	}
	return l, nil
}

func (l *Log) Build(modules ...string) (*log.Logger, error) {
	var conf *LoggerConfig
	for _, module := range modules {
		if c, has := l.Configs[module]; has {
			conf = c
		}
	}
	// logger config not found, use default
	if conf == nil {
		conf = &l.LoggerConfig
	}

	var encoder log.Encoder
	switch conf.Encoding {
	case "json":
		jsonConfig := conf.JsonEncoder
		encoder = log.NewJSONEncoder(log.JsonEncoderConfig{
			MessageKey:     jsonConfig.MessageKey,
			LevelKey:       jsonConfig.LevelKey,
			TimeKey:        jsonConfig.TimeKey,
			NameKey:        jsonConfig.NameKey,
			CallerKey:      jsonConfig.CallerKey,
			FunctionKey:    jsonConfig.FunctionKey,
			StacktraceKey:  jsonConfig.StacktraceKey,
			SkipLineEnding: jsonConfig.SkipLineEnding(),
			LineEnding:     jsonConfig.LineEnding,
			EscapeESC:      jsonConfig.EscapeESC(),
			EncodeLevel:    jsonConfig.EncodeLevel,
			EncodeTime:     jsonConfig.EncodeTime,
			EncodeDuration: jsonConfig.EncodeDuration,
			EncodeCaller:   jsonConfig.EncodeCaller,
			EncodeName:     jsonConfig.EncodeName,
		})
	case "console":
		consoleConfig := conf.ConsoleEncoder
		encoder = log.NewConsoleEncoder(log.ConsoleEncoderConfig{
			DisableLevel:      consoleConfig.DisableLevel(),
			DisableTime:       consoleConfig.DisableTime(),
			DisableName:       consoleConfig.DisableName(),
			DisableCaller:     consoleConfig.DisableCaller(),
			DisableFunction:   consoleConfig.DisableFunction(),
			DisableStacktrace: consoleConfig.DisableStacktrace(),
			SkipLineEnding:    consoleConfig.SkipLineEnding(),
			LineEnding:        consoleConfig.LineEnding,
			EncodeLevel:       consoleConfig.EncodeLevel,
			EncodeTime:        consoleConfig.EncodeTime,
			EncodeDuration:    consoleConfig.EncodeDuration,
			EncodeCaller:      consoleConfig.EncodeCaller,
			EncodeName:        consoleConfig.EncodeName,
			ConsoleSeparator:  consoleConfig.ConsoleSeparator,
		})
	default:
		panic("internal error: invalid log encoding")
	}

	logConfig := &log.Config{
		Level:             conf.Level,
		Development:       conf.Development(),
		DisableCaller:     conf.DisableCaller(),
		DisableStacktrace: conf.DisableStacktrace(),
		Sampling:          conf.Sampling,
		Encoder:           encoder,
		OutputPaths:       conf.OutputPaths,
		ErrorOutputPaths:  conf.ErrorOutputPaths,
	}

	if len(conf.InitialFields) > 0 {
		logConfig.InitialFields = make(map[string]interface{}, len(conf.InitialFields))
		for key, value := range conf.InitialFields {
			logConfig.InitialFields[key] = value
		}
	}

	return logConfig.Build()
}
