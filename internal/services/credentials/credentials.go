// Package `credentials` implements the login and register logic for the
// service.
package credentials

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
)

// Use it as global logger that will log everything that happens globally
var CredentialsLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Llongfile)
var SecretTokenKey = []byte("bardzo_tajny_token_nie_udostepniac")

const (
	DefaultExpirationTime = 3600
)

type Credentials struct {
	services.Service
}

func WithLogger(out io.Writer, prefix string, flags int) func(c *Credentials) {
	return func(c *Credentials) {
		if flags&log.Llongfile == log.Llongfile {
			CredentialsLogger.Fatalln("`log.Llongfile` flag is not allowed for the service purposes.")
		}
		if out == os.Stderr {
			CredentialsLogger.Fatalln("`os.Stderr` io.Writer is not allowed for the service purposes.")
		}
		c.Logger = log.New(out, prefix, flags)
	}
}

func WithRouter(opts ...gin.OptionFunc) func(c *Credentials) {
	return func(c *Credentials) {
		c.Router = gin.Default(opts...)
	}
}

func WithConfig(filename, ext string, cfgPaths ...string) func(c *Credentials) {
	return func(c *Credentials) {
		v := viper.New()
		v.SetConfigName(filename)
		v.SetConfigType(ext)
		for _, cfgPath := range cfgPaths {
			v.AddConfigPath(cfgPath)
		}

		if err := v.ReadInConfig(); err != nil {
			CredentialsLogger.Fatalf("Got error while parsing config file: %v\n", err)
		}
		c.ConfigReader = v
	}
}

func WithConnectionInfo(tableName string) func(c *Credentials) {
	return func(c *Credentials) {
		if c.ConfigReader == nil {
			CredentialsLogger.Fatalln("Before parsing a connection info, initialize your config reader.")
		}

		var ci services.ConnInfo
		if err := c.ConfigReader.UnmarshalKey(tableName, &ci); err != nil {
			CredentialsLogger.Fatalf("Got error while unmarshalling: %v\n", err)
		}
		c.ConnInfo = &ci
	}
}

func WithDatabase() func(c *Credentials) {
	return func(c *Credentials) {
		cfgParsed := database.ParseDriverConfig(c.ConnInfo)
		db, err := sql.Open(c.ConnInfo.Type, cfgParsed.FormatDSN())
		if err != nil {
			CredentialsLogger.Fatalf("Got error while creating a database driver: %v\n", err)
		}
		c.DB = db
	}
}

func CredentialsBuilder(opts ...func(*Credentials)) services.IService {
	c := &Credentials{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Credentials) Start() error {
	if err := c.HealthCheck(); err != nil {
		CredentialsLogger.Fatalf("HealthCheck failed, reason: %v\n", err)
	}
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

// validateUser extracts password from database and checks if the user exists.
// Then two password are compared to each other - if success, nil is returned.
func (c *Credentials) validateUser(fetchedPassword, requestPassword []byte) error {
	if err := bcrypt.CompareHashAndPassword(fetchedPassword, requestPassword); err != nil {
		return fmt.Errorf("user credentials don't match")
	}
	return nil
}

func (c *Credentials) HealthCheck() error {
	if c.Logger == nil {
		return fmt.Errorf("No logger setup")
	}

	if c.Router == nil {
		return fmt.Errorf("No router setup")
	}

	if c.DB == nil {
		return fmt.Errorf("No database setup")
	}

	if c.ConnInfo == nil {
		return fmt.Errorf("No connection info setup")
	}

	if c.ConfigReader == nil {
		return fmt.Errorf("No config setup")
	}
	return nil
}

// OnUserLogin implements logic for logging.
func (c *Credentials) OnUserLogin(ctx *gin.Context) {
	var ulr services.UserLoginRequest
	if err := ctx.ShouldBindJSON(&ulr); err != nil {
		c.Logger.Printf(services.JsonParsing, err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	requestedUsername, ok := ulr.Username.(string)
	if !ok {
		c.Logger.Printf(services.UsernameParsing, ulr.Username)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	requestedPassword, ok := ulr.Password.(string)
	if !ok {
		c.Logger.Printf(services.PasswordParsing, ulr.Password)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	requestedPasswordInBytes := []byte(requestedPassword)
	user, err := c.FetchUser(&requestedUsername)
	if err != nil {
		c.Logger.Printf(services.InvalidFetching, err)
		services.NewBadCredentialsCoreResponse(ctx, services.LoginErrorMessage)
		return
	}

	fetchedPasswordInBytes, ok := user.Password.([]byte)
	if !ok {
		c.Logger.Printf(services.PasswordParsing, user.Password)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	if err := c.validateUser(fetchedPasswordInBytes, requestedPasswordInBytes); err != nil {
		c.Logger.Printf(services.InvalidUserValidation, err)
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
		c.Logger.Printf(services.JsonParsing, err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	username, ok := u.Username.(string)
	if !ok {
		c.Logger.Printf(services.UsernameParsing, u.Username)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	if _, err := c.FetchUser(&username); err == nil {
		c.Logger.Printf("User `%v` exists.\n", username)
		// IMPORTANT: To prevent some bad actors, don't inform a user about it.
		services.NewBadCredentialsCoreResponse(ctx, services.LoginErrorMessage)
		return
	}

	// Start a transaction to push full user credentials.
	tx, err := c.DB.BeginTx(ctx, nil)
	defer tx.Rollback()

	if err != nil {
		c.Logger.Printf(services.TransactionProblem, err)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	requestPassword, ok := u.Password.(string)
	if !ok {
		c.Logger.Printf(services.PasswordParsing, u.Password)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPassword), 10)
	if err != nil {
		c.Logger.Printf("Couldn't hash the password: %v\n", requestPassword)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalErrorMessage)
		return
	}

	passwordStringified := string(hashedPassword)
	res, err := tx.Exec(`INSERT INTO user_credentials(username, password, email) VALUES (?, ?, ?)`, u.Username, passwordStringified, u.Email)
	if err != nil {
		c.Logger.Printf(services.TransactionProblem, err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegisterErrorMessage)
		return
	}

	id, _ := res.LastInsertId()
	_, err = tx.Exec(`INSERT INTO user_identity(ID, birthday, gender) VALUES (?, ?, ?)`, id, u.Birthday, u.Gender)
	if err != nil {
		c.Logger.Printf(services.TransactionProblem, err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegisterErrorMessage)
		return
	}

	if err = tx.Commit(); err != nil {
		c.Logger.Printf(services.TransactionProblem, err)
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

func (c *Credentials) String() string {
	return "Credentials"
}
