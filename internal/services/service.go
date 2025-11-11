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

var stateTable = map[ServiceState]string{
	StateOnline: "online",
	StateIdle:   "idle",
	StateDown:   "down",
}

type IService interface {
	Start() error
	Stop() error
	ReadConfig() error
}

// Base for any `Service`.
type Service struct {
	Router  *gin.Engine
	DB      *sql.DB
	Logger  *log.Logger
	State   ServiceState
	BaseUrl string
}
