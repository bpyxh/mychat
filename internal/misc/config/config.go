package config

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const ConfigName = "config"

type MysqlConfig struct {
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DbName      string `mapstructure:"dbname"`
	Charset     string `mapstructure:"charset"`
	MaxIdle     int    `mapstructure:"maxIdle"`
	MaxConn     int    `mapstructure:"maxConn"`
	MaxLifeTime int    `mapstructure:"maxLifeTime"`
}

type AppConfig struct {
	DevMode bool        `mapstructure:"devMode"`
	Mysql   MysqlConfig `mapstructure:"mysql"`
}

var (
	Config *AppConfig
	ROOT   string
)

func init() {
	curFilename := os.Args[0]
	binaryPath, err := exec.LookPath(curFilename)
	if err != nil {
		panic(err)
	}

	binaryPath, err = filepath.Abs(binaryPath)
	if err != nil {
		panic(err)
	}

	ROOT = filepath.Dir(binaryPath)
}

func Init() *AppConfig {
	zap.S().Debug("开始加载配置文件...")

	Config = &AppConfig{}

	viper.SetConfigName(ConfigName)
	viper.SetConfigType("toml")

	viper.AddConfigPath("./configs")
	viper.AddConfigPath(ROOT + "/configs")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	} else {
		zap.S().Debug("配置文件路径: ", viper.ConfigFileUsed())
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		zap.S().Fatal("解析配置文件失败: ", err)
		panic(err)
	}

	zap.S().Debug("配置加载完成")
	return Config
}
