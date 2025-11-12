// Package implements basic functionality of any service,
// requests and responses to users and between other APIs.
package services

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type ServiceState = uint8

const (
	StateOnline ServiceState = iota
	StateIdle
	StateDown
)

// Possible service statuses.
var stateTable = map[ServiceState]string{
	StateOnline: "online",
	StateIdle:   "idle",
	StateDown:   "down",
}

// IService defines what functions should have any Service.
type IService interface {
	// Start makes the service public, allowing for incoming
	// connections.
	Start() error
	// Stop gracefully stops a Service.
	Stop() error
	// ReadConfig reads config and applies it to the current
	// configuration of the service. The path to the config filename
	// should be present in the environment variables.
	ReadConfig() error
}

// Service implements basic functionality for any Service.
type Service struct {
	Router *gin.Engine
	DB     *sql.DB
	Logger *log.Logger
	State  ServiceState
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Birthday string `json:"birthday"`
	Gender   string `json:"gender"`
}

type CredentialsCoreResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   uint32 `json:"expires_in"`
	Message     any    `json:"message"`
}
