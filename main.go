package main

import (
	"fmt"

	"github.com/GoWeb/My_Blog/api/router"
	"github.com/GoWeb/My_Blog/core"
	"github.com/GoWeb/My_Blog/core/logger"
	"github.com/GoWeb/My_Blog/flags"
	"github.com/GoWeb/My_Blog/models"
	"go.uber.org/zap/zapcore"
)

// main 函数：程序入口
// 整个程序的启动流程：
// 1. 解析命令行参数
// 2. 加载配置文件
// 3. 初始化日志
// 4. 初始化数据库
// 5. 自动建表（仅开发/测试环境）
// 6. 启动 HTTP 服务
func main() {
	// ========== 第1步：解析命令行参数 ==========
	// 从命令行读取参数，如 -f、-port、-mode 等
	// 解析后存储到 flags.Opt 全局变量
	flags.Parse()

	// ========== 第2步：处理特殊参数 ==========
	// 如果用户传入 -v 参数，直接打印版本号并退出
	if flags.Opt.Version {
		fmt.Println("v1.0.0")
		return
	}

	// ========== 第3步：加载配置 ==========
	// 读取 YAML 配置文件
	// 会加载 config/app.yaml 和 config/app-{mode}.yaml 并合并
	// 例如：-mode dev 会加载 app.yaml + app-dev.yaml
	if err := core.ReadConfig(flags.Opt.Mode); err != nil {
		fmt.Println("配置加载失败:", err)
		return
	}

	// ========== 第4步：初始化日志 ==========
	// 根据配置中的日志级别初始化 zap 日志库
	// zap 是高性能的结构化日志库
	level := parseLogLevel(core.GlobalConfig.Log.Level)
	logger.Init(logger.Config{Level: level})
	defer logger.Sync() // 程序退出前刷新日志缓冲区

	// ========== 第5步：初始化数据库 ==========
	// 连接 MySQL 数据库
	// 连接参数来自配置文件
	if err := core.InitDB(); err != nil {
		logger.S.Fatalf("数据库连接失败: %v", err)
	}
	defer core.CloseDB() // 程序退出前关闭数据库连接

	// ========== 第5.5步：自动建表（仅开发/测试环境） ==========
	// 根据 Model 结构体自动创建或更新数据库表
	// 只在 dev 和 test 模式下执行，生产环境不会自动建表
	if !flags.IsRelease() {
		// models.AllModels() 返回所有需要建表的 Model
		if err := core.AutoMigrate(models.AllModels()...); err != nil {
			logger.S.Warnf("自动建表失败: %v", err)
		}
	}

	// ========== 第6步：获取服务地址 ==========
	// 格式：IP:Port，如 "0.0.0.0:8080"
	addr := flags.GetAddr()

	// ========== 第7步：打印启动日志 ==========
	// 使用结构化日志记录启动信息，方便排查
	logger.S.Infow("程序启动",
		"mode", flags.Opt.Mode,
		"config", fmt.Sprintf("app.yaml + app-%s.yaml", flags.Opt.Mode),
		"addr", addr,
	)

	// ========== 第8步：启动 Gin HTTP 服务 ==========
	// 初始化路由
	r := router.InitRouter()

	// 启动 HTTP 服务监听
	// r.Run(addr) 会阻塞，直到服务停止
	if err := r.Run(addr); err != nil {
		logger.S.Fatalf("服务启动失败: %v", err)
	}
}

// parseLogLevel 函数：将字符串日志级别转换为 zapcore.Level
// 参数：
//   - levelStr: 日志级别字符串，如 "debug"、"info"、"warn"、"error"
// 返回：
//   - 对应的 zap 日志级别
func parseLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel // 开发调试用，最详细
	case "info":
		return zapcore.InfoLevel  // 一般信息
	case "warn":
		return zapcore.WarnLevel  // 警告
	case "error":
		return zapcore.ErrorLevel // 错误
	default:
		return zapcore.InfoLevel // 默认 info 级别
	}
}
