package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type WebServerConfig struct {
	RunAddress  string `mapstructure:"run-address"` // run address
	EnableHTTPS bool   `mapstructure:"enable-https"`
	CertFile    string `mapstructure:"cert-file"`
	KeyFile     string `mapstructure:"key-file"`
}

type AppConfig struct {
	Name string `mapstructure:"name"` // app name
}

type LoggingConfiguration struct {
	Level      string `mapstructure:"level"`       // logging level
	Path       string `mapstructure:"path"`        // logging path
	MaxSize    int    `mapstructure:"max-size"`    // logging max size
	MaxBackups int    `mapstructure:"max-backups"` // logging max backups
	MaxAge     int    `mapstructure:"max-age"`     // logging max age
}

type DBConfig struct {
	ConnectionString string `mapstructure:"connection-string"` // database connection string
}

type JWTConfig struct {
	Secret string        `mapstructure:"secret"` // jwt secret
	Expire time.Duration `mapstructure:"expire"` // jwt expire
}

// Config struct
type Config struct {
	Debug   bool                  `mapstructure:"debug"`   // debug mode
	App     *AppConfig            `mapstructure:"app"`     // app name
	Logging *LoggingConfiguration `mapstructure:"logging"` // logging config
	Web     *WebServerConfig      `mapstructure:"web"`     // web server config
	PostDB  *DBConfig             `mapstructure:"postdb"`  // database config
	JWT     *JWTConfig            `mapstructure:"jwt"`     // jwt config
}

var C *Config = new(Config)

// InitConfiguration returns a new instance of Config
func InitConfiguration() *Config {
	initConfig()
	return C
}

func initConfig() {
	loadDefault()
	loadFile()
	loadEnv()
	loadFlags()

	if viper.GetBool("debug") {
		viper.SetDefault("logging.level", "debug")
	}

	viper.Unmarshal(C)
}

func loadEnv() {
	viper.SetEnvPrefix("GOPKEEPER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.BindEnv("web.run-address", "RUN_ADDRESS")
	viper.BindEnv("postdb.connection-string", "DATABASE_URI")
	viper.BindEnv("accrual.address", "ACCRUAL_SYSTEM_ADDRESS")
}

func loadFlags() {
	runAddr := flag.String("a", "", "server run address")
	dbURI := flag.String("d", "", "database connection string")

	flag.Parse()

	if *runAddr != "" {
		viper.Set("web.run-address", *runAddr)
	}

	if *dbURI != "" {
		viper.Set("postdb.connection-string", *dbURI)
	}
}

const (
	AppName = "gopkeeper"
)

func loadDefault() {
	viper.SetDefault("debug", false)
	viper.SetDefault("app.name", AppName)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.path", "logs")
	viper.SetDefault("logging.max-size", 500)
	viper.SetDefault("logging.max-backups", 3)
	viper.SetDefault("logging.max-age", 30)

	viper.SetDefault("web.run-address", ":8100")
	viper.SetDefault("web.enable-https", true)
	viper.SetDefault("web.cert-file", "server.crt")
	viper.SetDefault("web.key-file", "server.key")

	viper.SetDefault("jwt.expire", "1h")
}

func loadFile() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("../config/")
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", viper.GetString(AppName)))
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/", viper.GetString(AppName)))
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/config/", viper.GetString(AppName)))

	var fileLookupError viper.ConfigFileNotFoundError
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &fileLookupError) {
			if err := viper.WriteConfigAs("./config.json"); err != nil {
				fmt.Printf("Error writing config file: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Println(fmt.Printf("Error reading config file, %s. Use default only.\n", err))
		}
	}
}
