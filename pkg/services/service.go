// Package implements basic functionality of any service,
// requests and responses to users and between other APIs.
package services

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	gsdmysql "github.com/go-sql-driver/mysql"
)

var GlobalServiceLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags)

type ServiceState = uint8

const (
	StateOnline ServiceState = iota
	StateIdle
	StateDown
)

type Connection struct {
	Type     string `mapstructure:"type"`
	Name     string `mapstructure:"name"`
	Ip       string `mapstructure:"ip"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Service implements basic functionality for any Service.
type Service struct {
	Logger       *log.Logger
	Router       *gin.Engine
	DB           *sql.DB
	ConnInfo     *Connection
	State        ServiceState
	ConfigReader *viper.Viper
}

// IService defines what functions should have any Service.
type IService interface {
	// Start makes the service public, allowing for incoming
	// connections. Function should implement how Service shuts down.
	Start() error
	// HealthCheck validates a service's configuration.
	HealthCheck() error
	// ExposeConnection should return configuration info.
	ExposeConnection() *Connection
}

func NewLogger(out io.Writer, prefix string, flags int) *log.Logger {
	if flags&log.Llongfile == log.Llongfile {
		GlobalServiceLogger.Fatalln("`log.Llongfile` flag is not allowed for the service purposes.")
	}
	if out == os.Stderr {
		GlobalServiceLogger.Fatalln("`os.Stderr` io.Writer is not allowed for the service purposes.")
	}
	return log.New(out, prefix, flags)
}

func NewRouter(opts ...gin.OptionFunc) *gin.Engine {
	return gin.Default(opts...)
}

func NewViper(filename, ext string, cfgPaths ...string) *viper.Viper {
	v := viper.New()
	v.SetConfigName(filename)
	v.SetConfigType(ext)
	for _, cfgPath := range cfgPaths {
		v.AddConfigPath(cfgPath)
	}
	if err := v.ReadInConfig(); err != nil {
		GlobalServiceLogger.Fatalf("Got error while parsing config file: %v\n", err)
	}
	return v
}

func NewConnection(tableName string, v *viper.Viper) *Connection {
	if v == nil {
		GlobalServiceLogger.Fatalln("Viper instance is not initialized.")
	}
	var c Connection
	if err := v.UnmarshalKey(tableName, &c); err != nil {
		GlobalServiceLogger.Fatalf("Got error while unmarshalling: %v\n", err)
	}
	return &c
}

func NewDatabase(c *Connection) *sql.DB {
	cfg := ParseDriverConfig(c)
	db, err := sql.Open(c.Type,
		fmt.Sprintf("%v?parseTime=true", cfg.FormatDSN()))

	if err != nil {
		GlobalServiceLogger.Fatalf("Got error while creating a database driver: %v\n", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		GlobalServiceLogger.Fatalf("No database connection. Check the database configuration.")
		return nil
	}
	return db
}

func ParseDriverConfig(ci *Connection) *gsdmysql.Config {
	cfg := gsdmysql.NewConfig()
	cfg.User = ci.Username
	cfg.Passwd = ci.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", ci.Ip, ci.Port)
	cfg.DBName = ci.Name
	return cfg
}
