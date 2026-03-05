package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config 是整个应用的配置结构体，包含服务器、数据库等配置

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      logConfig
}

type ServerConfig struct {
	Port int
	Host string
	Mode string
}

type logConfig struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	Filename  string `mapstructure:"filename"`
	MaxSize   int    `mapstructure:"max_size"`
	MaxBackup int    `mapstructure:"max_backup"`
	MaxAge    int    `mapstructure:"max_age"`
	Compress  bool   `mapstructure:"compress"`
	Stdout    bool   `mapstructure:"stdout"`
}

type DatabaseConfig struct {
	Psql  PsqlConfig
	Redis RedisConfig
}
type PsqlConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}
type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

// 全局配置实例（私有，避免外部直接修改）
var (
	once     sync.Once
	instance *Config
)

func InitConfig(configFileName string) {
	once.Do(func() { // 单例模式确保配置只被加载一次
		// 配置viper
		v := viper.New()
		v.SetConfigName(configFileName) // 配置文件名（不带扩展名）
		v.SetConfigType("yaml")         // 配置文件类型
		v.AddConfigPath("./configs")    // 配置文件目录

		// 设置默认值
		v.SetDefault("server.host", "localhost")
		v.SetDefault("server.port", 8080)
		v.SetDefault("server.mode", "debug")

		v.SetDefault("database.psql.host", "localhost")
		v.SetDefault("database.psql.port", 5432)
		v.SetDefault("database.psql.user", "postgres")
		v.SetDefault("database.psql.password", "password")
		v.SetDefault("database.psql.database", "bill_management")

		v.SetDefault("database.redis.host", "localhost")
		v.SetDefault("database.redis.port", 6379)
		v.SetDefault("database.redis.password", "")

		v.SetDefault("log.level", "debug")
		v.SetDefault("log.format", "console")
		v.SetDefault("log.filename", "./logs/app.log")
		v.SetDefault("log.max_size", 100)
		v.SetDefault("log.max_backup", 10)
		v.SetDefault("log.max_age", 30)
		v.SetDefault("log.compress", true)
		v.SetDefault("log.stdout", true)

		// 读取配置文件
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("读取配置文件失败: %v", err))
		}
		// 解析配置文件到 Config 结构体
		instance = &Config{}
		if err := v.Unmarshal(instance); err != nil {
			panic(fmt.Errorf("解析配置文件失败: %v", err))
		}

		// 监视配置文件变化，自动重新加载
		v.WatchConfig() // 监视配置文件变化
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("配置文件更新: %s，重新加载配置", e.Name)
			//重新绑定配置到 Config 结构体
			if err := v.Unmarshal(instance); err != nil {
				log.Printf("重新加载配置失败: %v", err)
			}
		})
		log.Printf("配置文件 %s 加载成功", v.ConfigFileUsed())

	})
}

// GetConfig 返回全局配置实例
func GetConfig() *Config {
	if instance == nil {
		log.Panic("配置未初始化，请先调用 InitConfig")
	}
	return instance
}

func GetPostgresDSN() string {
	cfg := GetConfig().Database.Psql
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Database)
}
