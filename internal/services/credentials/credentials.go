// Package `credentials` implements the login and register logic for the
// service.
package credentials

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	gsdmysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
)

var SecretTokenKey = []byte("bardzo_tajny_token_nie_udostepniac")

const (
	DefaultExpirationTime = 3600
)

type Credentials struct {
	services.Service
}

// NewCredentials creates Service from configuration file.
func NewCredentials() (services.IService, error) {
	c := Credentials{}
	l := log.New(os.Stdout, "Credentials: ", log.LstdFlags)
	c.Logger = l
	c.Router = gin.Default()

	v, err := c.SetConfig()
	if err != nil {
		c.Logger.Printf("Error: %v.\n", err)
		c.Logger.Println("Service goes into `Idle mode`.")
		c.Logger.Println("Please note the service configuration is incomplete.")
		c.ConfigReader = nil
		c.State = services.StateIdle
		return &c, err
	}
	c.ConfigReader = v

	var ci services.ConnInfo
	if err := c.ConfigReader.UnmarshalKey("ConnInfo", &ci); err != nil {
		fmt.Println(err)
	}
	c.ConnInfo = &ci
	cfg := gsdmysql.NewConfig()
	cfg.User = ci.Username
	cfg.Passwd = ci.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", ci.Ip, ci.Port)
	cfg.DBName = ci.Name

	db, err := sql.Open(ci.Type, cfg.FormatDSN())
	if err != nil {
		c.Logger.Printf("Error: %v.\n", err)
		c.Logger.Println("Service goes into `Idle mode`.")
		c.Logger.Println("Please note the service configuration is incomplete.")
		c.State = services.StateIdle
		return &c, err
	}
	c.DB = db

	return &c, nil
}

func (c *Credentials) Start() error {
	// v1 of api.
	{
		v1 := c.Router.Group("/v1")
		v1.POST("auth/login", c.OnUserLogin)
		v1.POST("auth/register", c.OnUserRegister)
		// v1.POST("auth/forgot", c.OnForgotPassword)
	}

	go func() {
		if err := c.Router.Run(":9999"); err != nil && err != http.ErrServerClosed {
			c.Logger.Fatalf("Router failed: %v\n", err)
		}
	}()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	sig := <-kill
	c.Logger.Printf("Gracefully shutting down the server: %v\n.", sig)
	c.State = services.StateDown
	c.DB.Close()
	return fmt.Errorf("server closed")
}

func (c *Credentials) SetConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("CredentialsConfig")
	v.SetConfigType("toml")
	v.AddConfigPath(os.Getenv("CREDENTIALS_CONFIG_DIR_PATH"))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

// validateUser extracts password from database and checks if the user exists.
// Then two password are compared to each other - if success, nil is returned.
func (c *Credentials) validateUser(fetchedPassword, requestPassword []byte) error {
	if err := bcrypt.CompareHashAndPassword(fetchedPassword, requestPassword); err != nil {
		return fmt.Errorf("user credentials don't match")
	}
	return nil
}

// OnUserLogin implements logic for logging.
func (c *Credentials) OnUserLogin(ctx *gin.Context) {
	var ulr services.UserLoginRequest
	if err := ctx.ShouldBindJSON(&ulr); err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	requestedUsername, ok := ulr.Username.(string)
	if !ok {
		c.Logger.Printf("username (%v) field has invalid type\n", ulr.Username)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	requestedPassword, ok := ulr.Password.(string)
	if !ok {
		c.Logger.Printf("password (%v) field has invalid type\n", ulr.Password)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	requestedPasswordInBytes := []byte(requestedPassword)
	user, err := c.FetchUser(&requestedUsername)
	if err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.LoginErrorMessage)
		return
	}

	fetchedPasswordInBytes, ok := user.Password.([]byte)
	if !ok {
		c.Logger.Printf("password (%v) can't be casted into bytes\n", user.Password)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	if err := c.validateUser(fetchedPasswordInBytes, requestedPasswordInBytes); err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.LoginErrorMessage)
		return
	}

	ctx.JSON(http.StatusOK, services.CredentialsCoreResponse{
		AccessToken: c.GenerateSessionToken(requestedUsername, requestedPassword),
		TokenType:   "Bearer",
		ExpiresIn:   DefaultExpirationTime,
		Message:     "credentials are correct",
	})
}

// OnUserRegister implements logic when user tries to register.
func (c *Credentials) OnUserRegister(ctx *gin.Context) {
	var u services.UserRegisterRequest
	if err := ctx.ShouldBindJSON(&u); err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	username, ok := u.Username.(string)
	if !ok {
		c.Logger.Printf("username (%v) field has invalid type\n", u.Username)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	if _, err := c.FetchUser(&username); err == nil {
		c.Logger.Printf("user `%v` exists.\n", username)
		// IMPORTANT: To prevent some bad actors, don't inform a user about it.
		services.NewBadCredentialsCoreResponse(ctx, services.LoginErrorMessage)
		return
	}

	// Start a transaction to push full user credentials.
	tx, err := c.DB.BeginTx(ctx, nil)
	defer tx.Rollback()

	if err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	requestPassword, ok := u.Password.(string)
	if !ok {
		c.Logger.Printf("password (%v) field couldn't be casted to string\n", u.Password)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPassword), 10)
	if err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	passwordStringified := string(hashedPassword)
	res, err := tx.Exec(`INSERT INTO user_credentials(username, password, email) VALUES (?, ?, ?)`, u.Username, passwordStringified, u.Email)
	if err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegisterErrorMessage)
		return
	}

	id, _ := res.LastInsertId()
	_, err = tx.Exec(`INSERT INTO user_identity(ID, birthday, gender) VALUES (?, ?, ?)`, id, u.Birthday, u.Gender)
	if err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegisterErrorMessage)
		return
	}

	if err = tx.Commit(); err != nil {
		c.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegisterErrorMessage)
		return
	}

	ctx.JSON(http.StatusOK, services.CredentialsCoreResponse{
		AccessToken: c.GenerateSessionToken(username, passwordStringified),
		TokenType:   "Bearer",
		ExpiresIn:   DefaultExpirationTime,
		Message:     "registration completed",
	})
}

func (c *Credentials) GenerateSessionToken(username, password string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"password": password,
			"exp":      time.Now().Add(time.Hour * 1).Unix(),
		})

	// Dodam potem errory
	tok, _ := token.SignedString(SecretTokenKey)
	return tok
}

// ExposeConnInfo exposes configuration.
func (c *Credentials) ExposeConnInfo() *services.ConnInfo {
	return c.ConnInfo
}
