// Package implements basic functionality of any service,
// requests and responses to users and between other APIs.
package services

import (
	"database/sql"
	"fmt"

	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type ServiceState = uint8

const (
	StateOnline ServiceState = iota
	StateIdle
	StateDown
)

type ConnInfo struct {
	Type     string `mapstructure:"type"`
	Name     string `mapstructure:"name"`
	Ip       string `mapstructure:"ip"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Possible service statuses.
var stateTable = map[ServiceState]string{
	StateOnline: "online",
	StateIdle:   "idle",
	StateDown:   "down",
}

// IService defines what functions should have any Service.
type IService interface {
	// Start makes the service public, allowing for incoming
	// connections. Function should implement how Service shuts down.
	Start() error
	// HealthCheck validates a service's configuration.
	HealthCheck() error
	// ExposeConnInfo should return configuration info.
	ExposeConnInfo() *ConnInfo
}

// Service implements basic functionality for any Service.
type Service struct {
	Logger       *log.Logger
	Router       *gin.Engine
	DB           *sql.DB
	ConnInfo     *ConnInfo
	State        ServiceState
	ConfigReader *viper.Viper
}

// FetchUser queries the database and returns user if exists. User can only be
// fetched by username. Caller's DB context is used.
func (s *Service) FetchUser(username *string) (*User, error) {
	var u User
	userFetched := s.DB.QueryRow("SELECT * FROM user_credentials WHERE username=?", *username)
	if err := userFetched.Scan(&u.Id, &u.Username, &u.Password, &u.Email); err != nil {
		return nil, fmt.Errorf("user doesn't exist")
	}
	return &u, nil
}
