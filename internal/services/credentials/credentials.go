// Package `credentials` implements the login and register logic for the
// service.
package credentials

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	// "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/sadsonkeenolee/IO_projekt/internal/services"
)

type Credentials struct {
	services.Service
}

// NewCredentials creates Service from configuration file.
func NewCredentials() (services.IService, error) {
	c := Credentials{}
	l := log.New(os.Stdout, "Credentials: ", log.LstdFlags)
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

func (c *Credentials) Start() error {
	// v1 of api.
	{
		v1 := c.Router.Group("/v1")
		v1.POST("auth/login", c.OnUserLogin)
		v1.POST("auth/register", c.OnUserRegister)
	}
	c.Router.Run(":9999")
	return nil
}

func (c *Credentials) Stop() error { panic("Stop not implemented") }

func (c *Credentials) ReadConfig() error {
	viper.SetConfigName("CredentialService")
	viper.AddConfigPath(os.Getenv("CREDENTIALS_CONFIG_DIR_PATH"))

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	dbName := viper.Get("DatabaseConfig.name").(string)
	username := viper.Get("DatabaseConfig.username").(string)
	password := viper.Get("DatabaseConfig.password").(string)
	port := viper.Get("DatabaseConfig.port").(string)
	sqlInfo := fmt.Sprintf("%v:%v@tcp(localhost:%v)/%v",
		username, password, port, dbName)

	db, err := sql.Open("mysql", sqlInfo)
	if err != nil {
		return err
	}

	c.DB = db
	c.Router = gin.Default()
	return nil
}

// validateUser extracts password from database and checks if the user exists.
// Then two password are compared to each other - if success, nil is returned.
func (c *Credentials) validateUser(userFetched *sql.Row, password string) error {
	var hashedPassword []byte
	if err := userFetched.Scan(&hashedPassword); err != nil {
		return fmt.Errorf("user doesn't exist")
	}
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		return fmt.Errorf("user credentials don't match")
	}
	return nil
}

// OnUserLogin implements logic for logging.
func (c *Credentials) OnUserLogin(ctx *gin.Context) {
	var u services.UserLoginRequest
	if err := ctx.ShouldBindJSON(&u); err != nil {
		c.Logger.Println(err)
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     err.Error(),
		})
		return
	}

	userFetched := c.DB.QueryRow("SELECT password FROM user_credentials WHERE username=?", u.Username)
	if err := c.validateUser(userFetched, u.Password); err != nil {
		c.Logger.Println(err)
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, services.CredentialsCoreResponse{
		AccessToken: "Not implemented",
		TokenType:   "Not implemented",
		ExpiresIn:   999,
		Message:     "",
	})
}

// OnUserRegister implements logic when user tries to register.
func (c *Credentials) OnUserRegister(ctx *gin.Context) {
	var u services.UserRegisterRequest
	if err := ctx.ShouldBindJSON(&u); err != nil {
		c.Logger.Println(err)
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     "incorrect request.",
		})
		return
	}

	// Start a transaction to push full user credentials.
	tx, err := c.DB.BeginTx(ctx, nil)
	defer tx.Rollback()

	if err != nil {
		c.Logger.Println(err)
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     "service has some troubles.",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		c.Logger.Println(err)
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     "service has some troubles.",
		})
		return
	}

	passwordStringified := string(hashedPassword)
	res, err := tx.Exec(`INSERT INTO user_credentials(username, password, email) VALUES (?, ?, ?)`, u.Username, passwordStringified, u.Email)
	if err != nil {
		c.Logger.Println(err)
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     "couldn't register you in.",
		})
		return
	}

	id, _ := res.LastInsertId()
	_, err = tx.Exec(`INSERT INTO user_identity(ID, birthday, gender) VALUES (?, ?, ?)`, id, u.Birthday, u.Gender)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     "couldn't register you in.",
		})
		return
	}

	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusBadRequest, services.CredentialsCoreResponse{
			AccessToken: "",
			TokenType:   "none",
			ExpiresIn:   0,
			Message:     "couldn't register you in.",
		})
		return
	}
	ctx.JSON(http.StatusOK, services.CredentialsCoreResponse{
		AccessToken: "Not implemented",
		TokenType:   "Not implemented",
		ExpiresIn:   999,
		Message:     "",
	})
}
