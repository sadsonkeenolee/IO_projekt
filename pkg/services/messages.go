package services

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UserCore implements very basic info about user. Username and Password are
// either of type string or []byte.
type UserCore struct {
	Username any `json:"username"`
	Password any `json:"password"`
}

// User implements current schema of a database table.
type User struct {
	Id    uint
	Email string
	UserCore
}

// For readability
type UserLoginRequest = UserCore

type UserRegisterRequest struct {
	UserCore
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
	InternalErrorMessage  = "service is having some troubles"
	InvalidRequestMessage = "service couldn't answer your request"
	LoginErrorMessage     = "username or/and password are incorrect"
	RegisterErrorMessage  = "service is having some troubles while trying to register you"
	// Here constants that are not send out in requests
	JsonParsing           = "Error while binding JSON: %v\n"
	UsernameParsing       = "Username (%v) field has invalid type\n"
	PasswordParsing       = "Password (%v) field has invalid type\n"
	InvalidFetching       = "Error while fetching a user: %v\n"
	InvalidUserValidation = "Error while validating user: %v\n"
	TransactionProblem    = "Couldn't complete this transaction: %v\n"
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
