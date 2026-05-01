package core

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// ========== 配置结构体定义 ==========
// 使用 mapstructure tag 与 YAML 配置文件中的键对应

// SystemConfig 结构体：服务器配置
type SystemConfig struct {
	Ip   string `mapstructure:"ip"`   // IP 地址
	Port int    `mapstructure:"port"` // 端口号
}

// DatabaseConfig 结构体：数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`     // 数据库地址（IP 或域名）
	Port     int    `mapstructure:"port"`     // 数据库端口（MySQL 默认 3306）
	Username string `mapstructure:"username"` // 数据库用户名
	Password string `mapstructure:"password"` // 数据库密码
	Name     string `mapstructure:"name"`     // 数据库名
}

// JWTConfig 结构体：JWT 配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"` // JWT 签名密钥（盐值）
	Expire int    `mapstructure:"expire"`  // Token 过期时间（单位：小时）
}

// EmailConfig 结构体：邮件服务配置
type EmailConfig struct {
	Host     string `mapstructure:"host"`     // 邮件服务器地址
	Port     int    `mapstructure:"port"`     // 邮件服务器端口
	Username string `mapstructure:"username"` // 邮箱账号
	Password string `mapstructure:"password"` // 邮箱密码或授权码
	From     string `mapstructure:"from"`     // 发件人邮箱
}

// LogConfig 结构体：日志配置
type LogConfig struct {
	Filename   string `mapstructure:"filename"`   // 日志文件名
	MaxSize    int    `mapstructure:"maxSize"`   // 单个日志文件最大大小（MB）
	MaxBackups int    `mapstructure:"maxBackups"` // 保留的旧日志文件数量
	MaxAge     int    `mapstructure:"maxAge"`     // 旧日志文件保留天数
	Compress   bool   `mapstructure:"compress"`   // 是否压缩旧日志
	Level      string `mapstructure:"level"`      // 日志级别（debug、info、warn、error）
}

// Config 结构体：全局配置汇总
// 包含所有子配置
type Config struct {
	Server   SystemConfig   `mapstructure:"server"`   // 服务器配置
	Database DatabaseConfig `mapstructure:"database"` // 数据库配置
	JWT      JWTConfig      `mapstructure:"jwt"`      // JWT 配置
	Email    EmailConfig    `mapstructure:"email"`    // 邮件配置
	Log      LogConfig      `mapstructure:"log"`      // 日志配置
}

// GlobalConfig 全局配置变量
// 整个应用都可以访问这个变量来获取配置信息
var GlobalConfig *Config

// ReadConfig 函数：读取并合并配置文件
// 参数：
//   - env: 环境标识，如 "dev"（开发环境）、"prod"（生产环境）
//         会加载 config/app.yaml 和 config/app-{env}.yaml
//         两个文件会合并，后者覆盖前者
func ReadConfig(env string) error {
	// ========== 第1步：定义配置文件路径 ==========
	// 基础配置文件：config/app.yaml
	// 环境配置文件：config/app-{env}.yaml（如 config/app-dev.yaml）
	baseFile := "config/app.yaml"
	envFile := fmt.Sprintf("config/app-%s.yaml", strings.ToLower(env))

	// ========== 第2步：读取基础配置 ==========
	// viper.New() 创建一个新的 viper 实例
	v1 := viper.New()
	v1.SetConfigFile(baseFile)    // 设置配置文件路径
	v1.SetConfigType("yaml")      // 明确配置文件类型为 YAML
	if err := v1.ReadInConfig(); err != nil {
		// 读取失败，返回错误
		return fmt.Errorf("failed to read base config file: %w", err)
	}

	// ========== 第3步：读取环境配置 ==========
	// 环境配置会覆盖基础配置
	v2 := viper.New()
	v2.SetConfigFile(envFile)
	v2.SetConfigType("yaml")
	if err := v2.ReadInConfig(); err != nil {
		// 环境配置文件不存在也无所谓，可能只是没有环境特定的配置
		return fmt.Errorf("failed to read env config file: %w", err)
	}

	// ========== 第4步：合并配置 ==========
	// v1.MergeConfigMap 将 v2 的配置合并到 v1 中
	// v2 中的值会覆盖 v1 中同名的值
	if err := v1.MergeConfigMap(v2.AllSettings()); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	// ========== 第5步：解析配置到结构体 ==========
	// v1.Unmarshal(cfg) 将配置解析到 Config 结构体中
	// mapstructure tag 用于指定配置键与结构体字段的对应关系
	cfg := &Config{}
	if err := v1.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// ========== 第6步：设置全局配置 ==========
	GlobalConfig = cfg
	log.Println("Configuration loaded successfully:", baseFile, "+", envFile)
	return nil
}
