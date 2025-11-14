package services

import (
	"net/http"

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

const (
	InternalErrorMessage  = "service is having some troubles"
	InvalidRequestMessage = "service couldn't parse your request"
	LoginErrorMessage     = "username or/and password are incorrect"
	RegisterErrorMessage  = "service is having some troubles while trying to register you"
)

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
