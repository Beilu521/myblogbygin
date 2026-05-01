package logger

import (
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Filename   string        // 日志文件路径（例："./logs/app.log"）
	MaxSize    int           // 单个日志文件最大多大，单位MB（例：10 = 10MB，超过则切割）
	MaxBackups int           // 最多保留几个日志文件（例：5 = 保留最新5个旧的）
	MaxAge     int           // 日志保留多少天（例：30 = 30天前的旧日志删除）
	Compress   bool          // 是否压缩旧日志（true = 压缩成.gz，节省空间）
	Level      zapcore.Level // 日志输出级别（见下方级别说明）
}

var S *zap.SugaredLogger

func Init(cfg ...Config) {
	config := Config{
		Filename:   "./logs/app.log",  // 日志写入这个文件
		MaxSize:    10,                // 单文件超过10MB就新建一个
		MaxBackups: 5,                 // 保留最近5个日志文件
		MaxAge:     30,                // 30天之前的日志删除
		Compress:   true,              // 旧日志压缩
		Level:      zapcore.InfoLevel, // 日志级别（生产用Info，开发用Debug）
	}
	if len(cfg) > 0 {
		config = cfg[0]
	}

	ensureDir(filepath.Dir(config.Filename))

	fileWriter := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",                      // 日志里的时间字段叫什么
		LevelKey:      "level",                     // 日志级别字段叫什么
		NameKey:       "logger",                    // logger名
		CallerKey:     "caller",                    // 代码调用位置字段叫什么
		MessageKey:    "msg",                       // 日志内容字段叫什么
		StacktraceKey: "stacktrace",                // 堆栈跟踪字段叫什么
		LineEnding:    zapcore.DefaultLineEnding,   // 行结束符（不用改）
		EncodeTime:    zapcore.ISO8601TimeEncoder,  // 时间格式（ISO8601 = 2026-04-23T10:30:00）
		EncodeLevel:   zapcore.CapitalLevelEncoder, // 级别格式（Capital = INFO/ERROR/DEBUG）
		EncodeCaller:  zapcore.ShortCallerEncoder,  // 调用位置格式（简短路径）
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 日志输出成JSON格式（方便收集分析）
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),  // 同时输出到控制台
			zapcore.AddSync(fileWriter), // 同时写入文件
		),
		config.Level, // 用什么级别过滤
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	S = logger.Sugar()
	zap.ReplaceGlobals(logger)
}

func ensureDir(dir string) {
	if dir == "" || dir == "." {
		return
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

func Sync() {
	if S != nil {
		_ = S.Sync()
	}
}
