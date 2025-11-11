// Package `credentials` implements the login and register logic for this
// service.
package credentials

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	// "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/sadsonkeenolee/IO_projekt/internal/services"
)

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Credentials struct {
	services.Service
}

func NewCredentials() (services.IService, error) {
	l := log.New(os.Stdout, "Credentials: ", log.LstdFlags)
	c := Credentials{}
	c.Logger = l
	if err := c.ReadConfig(); err != nil {
		c.Logger.Printf("Error: %v.\n", err)
		c.Logger.Println("Service goes into `Idle mode`.")
		c.Logger.Println("Please note the service configuration is incomplete.")
		c.State = services.StateIdle
		return &c, err
	}
	return &c, nil
}

// The method `Start` makes the service public, allowing for incoming
// connections.
func (c *Credentials) Start() error {
	// Always define groups here.
	{
		v1 := c.Router.Group("/v1")
		v1.POST(fmt.Sprintf("%v%v", c.BaseUrl, "login/"), c.UserLogin)
		v1.POST(fmt.Sprintf("%v%v", c.BaseUrl, "register/"), c.UserRegister)
	}
	c.Router.Run(":9999")
	return nil
}
func (c *Credentials) Stop() error { panic("Stop not implemented") }

// The method `ReadConfig` reads config and applies it to the current
// configuration of the service. The path to the config filename
// should be present in the environment variables.
func (c *Credentials) ReadConfig() error {
	viper.SetConfigName("CredentialService")
	viper.AddConfigPath(os.Getenv("CREDENTIALS_CONFIG_DIR_PATH"))

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	dbName := viper.Get("DatabaseConfig.name").(string)
	// WARNING: Nieuzywane na razie
	// port := config.Get("DatabaseConfig.port").(string)
	username := viper.Get("DatabaseConfig.username").(string)
	password := viper.Get("DatabaseConfig.password").(string)

	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@/%v", username, password, dbName))
	if err != nil {
		return err
	}

	c.DB = db
	c.BaseUrl = viper.Get("RouterConfig.base_url").(string)
	c.Router = gin.Default()
	return nil
}

func (c *Credentials) UserLogin(ctx *gin.Context) {
	var ur UserRequest
	if err := ctx.ShouldBindJSON(&ur); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ping pong
	ctx.JSON(http.StatusOK, ur)
}
func (c *Credentials) UserRegister(ctx *gin.Context) {
	var ur UserRequest
	if err := ctx.ShouldBindJSON(&ur); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ping pong
	ctx.JSON(http.StatusOK, ur)
}
