package web

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gen2brain/beeep"
	"time"
)

type uiRoutes struct {
	key string
}

func (srv *uiRoutes) IssueAccessToken(validDays uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": time.Now().Add(time.Duration(24*validDays) * time.Hour),
	})
	return token.SignedString([]byte(srv.key))
}

func (srv *uiRoutes) Notify(title, message string) (bool, error) {
	err := beeep.Notify(title, message, "")
	return err == nil, err
}
