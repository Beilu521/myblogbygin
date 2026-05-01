package core

import (
	"fmt"
	"time"

	"github.com/GoWeb/My_Blog/core/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ========== 全局数据库实例 ==========
// 整个应用共享一个数据库连接池
var DB *gorm.DB

// DBOptions 结构体：数据库连接池配置
// 用于微调数据库连接性能
type DBOptions struct {
	MaxIdleConns  int           // 最大空闲连接数：连接池中保持的最大空闲连接数
	MaxOpenConns  int           // 最大打开连接数：同时最大连接数
	ConnMaxLife   time.Duration // 连接最大生命周期：超过此时间连接会被关闭重建
	SlowThreshold time.Duration // 慢查询阈值：超过此时间的查询会被记录为慢查询（单位：毫秒）
}

// defaultDBOptions 默认连接池配置
// 如果调用 InitDB 时不传参数，就使用这些默认值
var defaultDBOptions = &DBOptions{
	MaxIdleConns:  10,                     // 默认保持10个空闲连接
	MaxOpenConns:  100,                    // 默认最多打开100个连接
	ConnMaxLife:   time.Hour,              // 连接1小时后自动重建
	SlowThreshold: 200 * time.Millisecond, // 超过200ms算慢查询
}

// InitDB 函数：初始化数据库连接
// 参数：
//   - opts: 可选的连接池配置，不传则使用默认配置
// 使用方式：
//   core.InitDB()  // 使用默认配置
//   core.InitDB(&DBOptions{MaxOpenConns: 50})  // 自定义配置
func InitDB(opts ...*DBOptions) error {
	// ========== 第1步：检查配置是否已加载 ==========
	// InitDB 依赖 GlobalConfig 中的数据库配置
	// 如果没有先调用 ReadConfig，这里会返回错误
	if GlobalConfig == nil {
		return fmt.Errorf("配置文件未加载，请先调用 core.ReadConfig()")
	}

	// ========== 第2步：获取数据库配置 ==========
	// 从全局配置中读取 Database 节点
	dbConfig := GlobalConfig.Database

	// ========== 第3步：构建 DSN 连接字符串 ==========
	// DSN (Data Source Name) 格式：
	// 用户名:密码@tcp(地址:端口)/数据库名?参数
	// 参数说明：
	//   charset=utf8mb4    支持 emoji 和所有字符
	//   parseTime=True    自动转换时间类型（time.Time <-> datetime）
	//   loc=Local         使用本地时区
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.Username, // 数据库用户名
		dbConfig.Password, // 数据库密码
		dbConfig.Host,     // 数据库地址（IP或域名）
		dbConfig.Port,     // 数据库端口（MySQL默认3306）
		dbConfig.Name,     // 数据库名
	)

	// ========== 第4步：处理连接参数 ==========
	// 如果传入了自定义参数就用自定义的，否则用默认参数
	options := defaultDBOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	// ========== 第5步：配置 GORM ==========
	// Logger: 使用 GORM 默认日志记录器
	// LogMode: Info 级别，会打印所有 SQL 语句（开发用）
	// 生产环境可以改成 Warn 或 Error，减少日志输出
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	}

	// ========== 第6步：建立数据库连接 ==========
	// gorm.Open() 打开数据库连接
	// 参数1: mysql.Open(dsn) 使用 MySQL 驱动和 DSN 连接字符串
	// 参数2: gormConfig GORM 配置
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// ========== 第7步：获取原生数据库连接 ==========
	// GORM 封装的 DB 底层是标准库的 *sql.DB
	// 需要获取它来设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// ========== 第8步：设置连接池 ==========
	// 连接池用于管理数据库连接的复用
	// SetMaxIdleConns: 空闲时保持的连接数（建议设置较小值）
	// SetMaxOpenConns: 最多同时打开的连接数（根据服务器性能调整）
	// SetConnMaxLifetime: 连接最大存活时间（避免连接长期占用）
	sqlDB.SetMaxIdleConns(options.MaxIdleConns)
	sqlDB.SetMaxOpenConns(options.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(options.ConnMaxLife)

	// ========== 第9步：记录成功日志 ==========
	// 使用结构化日志，记录关键信息方便排查
	logger.S.Infow("数据库连接成功",
		"host", dbConfig.Host,
		"port", dbConfig.Port,
		"database", dbConfig.Name,
	)

	// ========== 第10步：保存实例到全局变量 ==========
	// 将 *gorm.DB 保存到全局变量 DB
	// 后续通过 core.GetDB() 获取数据库实例
	DB = db
	return nil
}

// InitDBWithOptions 函数：使用自定义配置初始化数据库
// 是 InitDB 的别名，更语义化的写法
func InitDBWithOptions(opt *DBOptions) error {
	return InitDB(opt)
}

// CloseDB 函数：关闭数据库连接
// 使用方式：
//   defer core.CloseDB()  // 程序退出前自动关闭
func CloseDB() error {
	// 已经关闭或未初始化，直接返回 nil
	if DB == nil {
		return nil
	}

	// 获取底层 *sql.DB
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 关闭连接
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	logger.S.Info("数据库连接已关闭")
	return nil
}

// GetDB 函数：获取数据库实例
// 返回全局的 *gorm.DB
// 使用方式：
//   db := core.GetDB()
//   db.Create(&user)           // 创建记录
//   db.First(&user, id)        // 根据ID查询
//   db.Where("name = ?", "Tom").First(&user)  // 条件查询
func GetDB() *gorm.DB {
	return DB
}

// AutoMigrate 函数：自动迁移数据库表
// 根据 struct 定义自动创建或更新表结构
// 使用方式：
//   core.AutoMigrate(&model.User{}, &model.Article{})
//   // 会自动创建 users 和 articles 表
//
// 注意：
//   - 会自动添加缺失的列
//   - 不会删除或修改现有列
//   - 不会自动创建数据库，需要先手动创建数据库
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	return DB.AutoMigrate(models...)
}
