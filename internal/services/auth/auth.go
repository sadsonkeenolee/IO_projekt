// Package `auth` implements logic for a user login and registration.
package auth

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
var AuthLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Llongfile)

// Dummy SecretTokenKey, in the production code you never store it there.
var SecretTokenKey = []byte("bardzo_tajny_token_nie_udostepniac")

const (
	DefaultExpirationTime = 3600
)

type AuthService struct {
	services.Service
}

func WithLogger(l *log.Logger) func(a *AuthService) {
	return func(a *AuthService) {
		a.Logger = l
	}
}

func WithRouter(ge *gin.Engine) func(a *AuthService) {
	return func(a *AuthService) {
		a.Router = ge
	}
}

func WithViper(v *viper.Viper) func(a *AuthService) {
	return func(a *AuthService) {
		a.ConfigReader = v
	}
}

func WithConnectionInfo(c *services.Connection) func(a *AuthService) {
	return func(a *AuthService) {
		a.ConnInfo = c
	}
}

func WithDatabase(db *sql.DB) func(a *AuthService) {
	return func(a *AuthService) {
		a.DB = db
	}
}

// Main constructor for the authorization service. Provide all necessary
// functions into `opts` - they will be executed in the given order.
func AuthBuilder(opts ...func(*AuthService)) services.IService {
	a := &AuthService{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *AuthService) Start() error {
	if err := a.HealthCheck(); err != nil {
		AuthLogger.Fatalf("HealthCheck failed, reason: %v\n", err)
	}
	// v1 of api.
	{
		v1 := a.Router.Group("/v1")
		v1.POST("auth/login", a.OnUserLogin)
		v1.POST("auth/register", a.OnUserRegister)
		v1.POST("auth/event/push", a.OnUserEventPush)
		v1.POST("auth/event/pull", a.OnUserEventPull)
	}

	go func() {
		if err := a.Router.Run(":9999"); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatalf("Router failed: %v\n", err)
		}
	}()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	sig := <-kill
	a.Logger.Printf("Gracefully shutting down the server: %v\n.", sig)
	a.State = services.StateDown
	a.DB.Close()
	return fmt.Errorf("server closed")
}

func (a *AuthService) OnUserEventPush(ctx *gin.Context) {
	var u database.UserEventPushRequest
	if err := ctx.ShouldBindBodyWithJSON(&u); err != nil {
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		a.Logger.Println(err)
		return
	}
	if !u.ValidateFields() {
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}
	if _, err := a.DB.Exec(`call push_events(?, ?, ?, ?)`,
		u.Token, u.EventName, u.ItemType, u.ItemId); err != nil {
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		a.Logger.Println(err)
		return
	}
	services.NewGoodContentRequest(ctx, "added")
}

func (a *AuthService) OnUserEventPull(ctx *gin.Context) {
	var u database.UserEventPushRequest
	if err := ctx.ShouldBindBodyWithJSON(&u); err != nil {
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		a.Logger.Println(err)
		return
	}

	var ue database.UserEventPullResponse
	eventRows, err := a.DB.Query(`call pull_events(?, ?)`, u.Token, u.EventName)
	if err != nil {
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	oppositeRows, err := a.DB.Query(`call pull_events(?, ?)`, u.Token, database.OppositeEvents[u.EventName])
	if err != nil {
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	defer eventRows.Close()
	defer oppositeRows.Close()

	// Event : Name (event's name)
	id2event := map[uint64]database.Event{}

	for eventRows.Next() {
		event := database.Event{}
		var timestamp time.Time
		if err := eventRows.Scan(&event.ItemId, &event.Name, &event.ItemType, &timestamp); err != nil {
			continue
		}
		event.Timestamp = timestamp.Unix()
		id2event[event.ItemId] = event
	}

	for oppositeRows.Next() {
		event := database.Event{}
		var timestamp time.Time
		if err := oppositeRows.Scan(&event.ItemId, &event.Name, &event.ItemType, &timestamp); err != nil {
			continue
		}

		event.Timestamp = timestamp.Unix()
		if item, ok := id2event[event.ItemId]; ok {
			if item.Timestamp < event.Timestamp {
				id2event[event.ItemId] = event
			}
		}
	}

	items := make([]database.Event, 0, len(id2event))
	for _, v := range id2event {
		if v.Name == u.EventName {
			items = append(items, v)
		}
	}
	ue.Items = items
	services.NewGoodContentRequest(ctx, ue)
}

func (a *AuthService) HealthCheck() error {
	if a.Logger == nil {
		return fmt.Errorf("No logger setup")
	}

	if a.Router == nil {
		return fmt.Errorf("No router setup")
	}

	if a.DB == nil {
		return fmt.Errorf("No database setup")
	}

	if a.ConnInfo == nil {
		return fmt.Errorf("No connection info setup")
	}

	if a.ConfigReader == nil {
		return fmt.Errorf("No config setup")
	}
	return nil
}

// UserLoginValidation compares the salted password in the database with the password
// provided by the user.
func (a *AuthService) UserLoginValidation(fetchedPassword, requestPassword []byte) error {
	if err := bcrypt.CompareHashAndPassword(fetchedPassword, requestPassword); err != nil {
		return fmt.Errorf(services.LoginFailedMessage)
	}
	return nil
}

// OnUserLogin implements logic for logging.
func (a *AuthService) OnUserLogin(ctx *gin.Context) {
	var ulr services.UserLoginRequest
	if err := ctx.ShouldBindJSON(&ulr); err != nil {
		a.Logger.Printf(services.JsonParsingProblemMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	user, err := a.FetchUser(&ulr.Username)
	if err != nil {
		a.Logger.Printf(services.UserFetchingProblemMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.LoginFailedMessage)
		return
	}

	if err := a.UserLoginValidation(user.Password, []byte(ulr.Password)); err != nil {
		a.Logger.Printf(services.UserValidationProblemMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.LoginFailedMessage)
		return
	}

	// Try to check if session exists
	var sessionTok string
	if err := a.DB.QueryRow(`select token from user_login_timestamps where user_id=?`, user.Id).Scan(&sessionTok); err != nil {
		a.Logger.Printf("Session token not found, reason: %v\n", err)
	}

	var isSessionValid bool = false
	if sessionTok != "" {
		if err := a.DB.QueryRow(`select check_if_session_is_valid(?)`, sessionTok).Scan(&isSessionValid); err != nil {
			a.Logger.Printf("Error while checking if the session is valid, reason: %v\n", err)
			return
		} else {
			isSessionValid = true
		}
	}

	if isSessionValid {
		ctx.JSON(http.StatusOK, services.CredentialsCoreResponse{
			AccessToken: sessionTok,
			TokenType:   "Bearer",
			ExpiresIn:   DefaultExpirationTime,
			Message:     "Logged in.",
		})
		return
	}

	// Construct new token
	if sessionTok = a.GenerateSessionToken(ulr.Username, ulr.Password); sessionTok != "" {
		a.DB.Exec(`call create_user_session(?, ?)`, ulr.Username, sessionTok)
		ctx.JSON(http.StatusOK, services.CredentialsCoreResponse{
			AccessToken: sessionTok,
			TokenType:   "Bearer",
			ExpiresIn:   DefaultExpirationTime,
			Message:     "Logged in.",
		})
		return
	}

}

// OnUserRegister implements logic when user tries to register.
func (a *AuthService) OnUserRegister(ctx *gin.Context) {
	var u services.UserRegisterRequest[string]
	if err := ctx.ShouldBindJSON(&u); err != nil {
		a.Logger.Printf(services.JsonParsingProblemMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	if _, err := a.FetchUser(&u.Username); err == nil {
		a.Logger.Printf("User `%v` exists.\n", u.Username)
		// IMPORTANT: To prevent some bad actors, don't inform a user about it.
		services.NewBadCredentialsCoreResponse(ctx, services.LoginFailedMessage)
		return
	}

	// Start a transaction to push full user credentials.
	tx, err := a.DB.BeginTx(ctx, nil)
	defer tx.Rollback()

	if err != nil {
		a.Logger.Printf(services.TransactionNotCompletedMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalMessage)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		a.Logger.Printf("Couldn't hash the password: %v\n", u.Password)
		services.NewBadCredentialsCoreResponse(ctx, services.InternalMessage)
		return
	}

	passwordStringified := string(hashedPassword)
	res, err := tx.Exec(`INSERT INTO user_credentials(username, password, email) VALUES (?, ?, ?)`, u.Username, passwordStringified, u.Email)
	if err != nil {
		a.Logger.Printf(services.TransactionNotCompletedMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegistrationFailedMessage)
		return
	}

	id, _ := res.LastInsertId()
	_, err = tx.Exec(`INSERT INTO user_identity(ID, birthday, gender) VALUES (?, ?, ?)`, id, u.Birthday, u.Gender)
	if err != nil {
		a.Logger.Printf(services.TransactionNotCompletedMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegistrationFailedMessage)
		return
	}

	if err = tx.Commit(); err != nil {
		a.Logger.Printf(services.TransactionNotCompletedMessage, err)
		services.NewBadCredentialsCoreResponse(ctx, services.RegistrationFailedMessage)
		return
	}

	ctx.JSON(http.StatusOK, services.CredentialsCoreResponse{
		AccessToken: "None",
		TokenType:   "None",
		ExpiresIn:   DefaultExpirationTime,
		Message:     "Registered.",
	})
}

// GenerateSessionToken generates cookies for logged in users.
func (a *AuthService) GenerateSessionToken(username, password string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"password": password,
			"exp":      time.Now().Add(time.Hour * 1).Unix(),
		})

	tok, _ := token.SignedString(SecretTokenKey)
	return tok
}

// FetchUser queries the database and returns user if exists. User can only be
// fetched by username. Caller's DB context is used.
func (a *AuthService) FetchUser(username *string) (*services.User[[]byte], error) {
	var u services.User[[]byte]
	if err := a.DB.QueryRow(`SELECT * FROM user_credentials WHERE username=?`,
		*username).Scan(&u.Id, &u.Username, &u.Password, &u.Email); err != nil {
		return nil, fmt.Errorf(services.UserDoesntExistMessage, *username)
	}
	return &u, nil
}

// ExposeConnection exposes configuration.
func (a *AuthService) ExposeConnection() *services.Connection {
	return a.ConnInfo
}
func (a *AuthService) String() string {
	return "Credentials"
}
