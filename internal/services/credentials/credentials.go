// Package `credentials` implements the login and register logic for this
// service.
package credentials

import "github.com/sadsonkeenolee/IO_projekt/internal/services"

type Credentials struct {
	services.IService
}

func NewCredentials() (services.IService, error) {
	return &Credentials{}, nil
}

func (c *Credentials) Start() { panic("Start not implemented") }
func (c *Credentials) Stop()  { panic("Stop not implemented") }
