package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Milvus MilvusConfig `mapstructure:"milvus"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Minio  MinioConfig  `mapstructure:"minio"`
	Doubao DoubaoConfig `mapstructure:"doubao"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// MySQLConfig MySQL 数据库配置
type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DB              string        `mapstructure:"db"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DSN 返回 MySQL 数据库连接字符串
func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DB)
}

// MilvusConfig Milvus 向量库配置
type MilvusConfig struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	CollectionName string `mapstructure:"collection_name"`
}

// Address 返回 Milvus 连接地址
func (c *MilvusConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// RedisConfig Redis 缓存配置
type RedisConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	DB          int           `mapstructure:"db"`
	Password    string        `mapstructure:"password"`
	PoolSize    int           `mapstructure:"pool_size"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
}

// Address 返回 Redis 连接地址
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// MinioConfig MinIO 对象存储配置
type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// DoubaoConfig 豆包大模型配置
type DoubaoConfig struct {
	APIKey         string  `mapstructure:"api_key"`
	EmbeddingModel string  `mapstructure:"embedding_model"`
	ChatModel      string  `mapstructure:"chat_model"`
	Temperature    float64 `mapstructure:"temperature"`
	MaxToken       int     `mapstructure:"max_token"`
}

// Load 加载配置文件
func Load(configPaths ...string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 默认配置文件搜索路径
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// 支持自定义配置路径
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// 支持环境变量覆盖
	v.AutomaticEnv()

	// 绑定环境变量到嵌套配置项
	bindEnvVars(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

// bindEnvVars 绑定环境变量到配置项
func bindEnvVars(v *viper.Viper) {
	// 豆包大模型配置
	_ = v.BindEnv("doubao.api_key", "DOUBAO_API_KEY")
	_ = v.BindEnv("doubao.embedding_model", "DOUBAO_EMBEDDING_MODEL")
	_ = v.BindEnv("doubao.chat_model", "DOUBAO_CHAT_MODEL")
	_ = v.BindEnv("doubao.temperature", "DOUBAO_TEMPERATURE")
	_ = v.BindEnv("doubao.max_token", "DOUBAO_MAX_TOKEN")

	// MySQL 配置
	_ = v.BindEnv("mysql.host", "MYSQL_HOST")
	_ = v.BindEnv("mysql.port", "MYSQL_PORT")
	_ = v.BindEnv("mysql.user", "MYSQL_USER")
	_ = v.BindEnv("mysql.password", "MYSQL_PASSWORD")
	_ = v.BindEnv("mysql.db", "MYSQL_DB")

	// Redis 配置
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")

	// Milvus 配置
	_ = v.BindEnv("milvus.host", "MILVUS_HOST")
	_ = v.BindEnv("milvus.port", "MILVUS_PORT")

	// MinIO 配置
	_ = v.BindEnv("minio.endpoint", "MINIO_ENDPOINT")
	_ = v.BindEnv("minio.access_key", "MINIO_ACCESS_KEY")
	_ = v.BindEnv("minio.secret_key", "MINIO_SECRET_KEY")
	_ = v.BindEnv("minio.bucket", "MINIO_BUCKET")

	// 服务配置
	_ = v.BindEnv("server.port", "SERVER_PORT")
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8000"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 10 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 10 * time.Second
	}
	if cfg.MySQL.MaxOpenConns == 0 {
		cfg.MySQL.MaxOpenConns = 20
	}
	if cfg.MySQL.MaxIdleConns == 0 {
		cfg.MySQL.MaxIdleConns = 10
	}
	if cfg.MySQL.ConnMaxLifetime == 0 {
		cfg.MySQL.ConnMaxLifetime = 30 * time.Minute
	}
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 10
	}
	if cfg.Redis.IdleTimeout == 0 {
		cfg.Redis.IdleTimeout = 30 * time.Minute
	}
	if cfg.Doubao.Temperature == 0 {
		cfg.Doubao.Temperature = 0.1
	}
	if cfg.Doubao.MaxToken == 0 {
		cfg.Doubao.MaxToken = 2048
	}
}
