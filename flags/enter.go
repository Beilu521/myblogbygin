package flags

import (
	"flag"      // Go 标准库：命令行参数解析
	"fmt"       // 格式化输出
	"os"        // 操作系统功能，获取环境变量
	"strings"   // 字符串处理
)

// Options 结构体：命令行参数选项
// 通过 flag 库解析命令行参数
type Options struct {
	ConfigFile string // -f: 配置文件路径（默认 config/dev.yaml）
	DBInit     bool   // -db: 是否初始化数据库（true=执行数据库迁移）
	Version    bool   // -v: 显示版本号
	Port       int    // -port: 服务端口（默认 8080）
	Host       string // -host: 服务绑定地址（默认 0.0.0.0）
	Mode       string // -mode: 运行模式（默认 dev，可选 dev/test/release）
	LogLevel   string // -log-level: 日志级别（默认 debug，可选 debug/info/warn/error）
}

// Opt 全局命令行参数变量
// 整个应用可以通过 flags.Opt 访问命令行参数
var Opt *Options

// Parse 函数：解析命令行参数
// 需要在程序启动时调用
// 使用方式：
//   ./myapp -f config/prod.yaml -port 8080 -mode release
//   ./myapp -db  // 初始化数据库
func Parse() {
	Opt = &Options{}

	// ========== 定义命令行参数 ==========
	// flag.StringVar / flag.IntVar / flag.BoolVar
	// 参数1: 绑定到的变量地址
	// 参数2: 参数名（如 -f）
	// 参数3: 默认值
	// 参数4: 参数说明（-h 时显示）

	flag.StringVar(&Opt.ConfigFile, "f", "config/dev.yaml", "配置文件路径")
	flag.BoolVar(&Opt.DBInit, "db", false, "是否初始化数据库")
	flag.BoolVar(&Opt.Version, "v", false, "显示版本号")
	flag.IntVar(&Opt.Port, "port", 8080, "服务端口")
	flag.StringVar(&Opt.Host, "host", "0.0.0.0", "服务绑定地址")
	flag.StringVar(&Opt.Mode, "mode", "dev", "运行模式：dev/test/release")
	flag.StringVar(&Opt.LogLevel, "log-level", "debug", "日志级别：debug/info/warn/error")

	// ========== 执行解析 ==========
	// 解析命令行参数并填充到上述变量
	// 必须在定义完所有参数后调用
	flag.Parse()
}

// GetEnv 函数：获取环境变量（带默认值）
// 参数：
//   - key: 环境变量名
//   - defaultValue: 如果环境变量未设置，返回此默认值
// 使用方式：
//   dbHost := flags.GetEnv("DB_HOST", "localhost")
func GetEnv(key, defaultValue string) string {
	// os.Getenv 获取环境变量，如果未设置返回空字符串
	if value := os.Getenv(key); value != "" {
		return value
	}
	// 未设置时返回默认值
	return defaultValue
}

// PrintUsage 函数：打印命令行参数使用说明
// 相当于 -h / --help 的输出
func PrintUsage() {
	// flag.Usage 打印内置的使用说明
	flag.Usage()
	fmt.Println("\n示例：")
	fmt.Println("  开发环境: myapp -mode dev -port 8080")
	fmt.Println("  测试环境: myapp -mode test -port 8081")
	fmt.Println("  生产环境: myapp -mode release -port 80 -host 0.0.0.0")
	fmt.Println("  初始化DB: myapp -db")
	fmt.Println("  查看版本: myapp -v")
}

// IsRelease 函数：判断是否为生产模式
func IsRelease() bool {
	return Opt.Mode == "release"
}

// IsDev 函数：判断是否为开发模式
func IsDev() bool {
	return Opt.Mode == "dev"
}

// IsTest 函数：判断是否为测试模式
func IsTest() bool {
	return Opt.Mode == "test"
}

// GetAddr 函数：获取服务地址
// 返回格式：IP:Port，如 "0.0.0.0:8080"
func GetAddr() string {
	return fmt.Sprintf("%s:%d", Opt.Host, Opt.Port)
}

// CheckRequiredFlags 函数：检查必需的参数
// 参数：
//   - required: 必需的参数名列表，如 ["config", "port"]
// 返回：
//   - 如果缺少必需参数，返回 error
//   - 如果都满足，返回 nil
func CheckRequiredFlags(required []string) error {
	for _, flagName := range required {
		switch flagName {
		case "config":
			// 检查配置文件参数
			if Opt.ConfigFile == "" {
				return fmt.Errorf("缺少必需参数: -f <配置文件路径>")
			}
		case "port":
			// 检查端口参数
			if Opt.Port == 0 {
				return fmt.Errorf("缺少必需参数: -port <端口号>")
			}
		}
	}
	return nil
}

// ValidateFlags 函数：验证命令行参数的合法性
// 检查 Mode 和 LogLevel 是否在有效选项内
func ValidateFlags() error {
	// ========== 验证 Mode ==========
	// 检查运行模式是否有效
	validModes := []string{"dev", "test", "release"}
	modeValid := false
	for _, m := range validModes {
		if Opt.Mode == m {
			modeValid = true
			break
		}
	}
	if !modeValid {
		return fmt.Errorf("无效的 mode: %s，有效选项: %s", Opt.Mode, strings.Join(validModes, "/"))
	}

	// ========== 验证 LogLevel ==========
	// 检查日志级别是否有效
	validLogLevels := []string{"debug", "info", "warn", "error"}
	logLevelValid := false
	for _, l := range validLogLevels {
		if Opt.LogLevel == l {
			logLevelValid = true
			break
		}
	}
	if !logLevelValid {
		return fmt.Errorf("无效的 log-level: %s，有效选项: %s", Opt.LogLevel, strings.Join(validLogLevels, "/"))
	}

	return nil
}
