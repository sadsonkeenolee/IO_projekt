package services

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UserCore implements very basic info about user. Username and Password are
// either of type string or []byte.
type UserCore[T string | []byte] struct {
	Username T `json:"username"`
	Password T `json:"password"`
	Token    T `json:"token"`
}

// User implements current schema of a database table.
type User[T string | []byte] struct {
	Id    uint
	Email string
	UserCore[T]
}

// For readability
type UserLoginRequest = UserCore[string]

type UserRegisterRequest[T string | []byte] struct {
	UserCore[T]
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

type ContentRequestReponse struct {
	Timestamp int64 `json:"timestamp"`
	Content   any   `json:"content"`
}

type TvResponse struct {
	Content []any
}

const (
	InternalMessage           = "service is having some troubles"
	InvalidRequestMessage     = "service couldn't answer your request"
	LoginFailedMessage        = "username or/and password are incorrect"
	RegistrationFailedMessage = "service is having some troubles while trying to register you"
	// Here constants that are not send out in requests
	JsonParsingProblemMessage      = "Error while binding JSON: %v\n"
	PasswordParsingProblemMessage  = "Password (%v) field has invalid type\n"
	TransactionNotCompletedMessage = "Couldn't complete this transaction: %v\n"
	UserDoesntExistMessage         = "User (%v) doesn't exist"
	UserFetchingProblemMessage     = "Error while fetching a user: %v\n"
	UserValidationProblemMessage   = "Error while validating user: %v\n"
	UsernameParsingProblemMessage  = "Username (%v) field has invalid type\n"
	TableRebuildMessage            = "Can't rebuild table `(%v)`, reason: `(%v)`"
)

func NewGoodContentRequest(ctx *gin.Context, content any) {
	ctx.JSON(
		http.StatusFound,
		ContentRequestReponse{
			Timestamp: time.Now().Unix(),
			Content:   content,
		})
}

func NewBadContentRequest(ctx *gin.Context, message string) {
	ctx.JSON(
		http.StatusNotFound,
		ContentRequestReponse{
			Timestamp: time.Now().Unix(),
			Content:   message,
		})
}

// NewBadCredentialsCoreResponse defines a basic error message to reduce
// boilerplate.
func NewBadCredentialsCoreResponse(ctx *gin.Context, message string) {
	ctx.JSON(
		http.StatusBadRequest,
		CredentialsCoreResponse{
			AccessToken: "None",
			TokenType:   "None",
			ExpiresIn:   0,
			Message:     message,
		},
	)
}
