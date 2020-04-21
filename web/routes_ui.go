package web

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gen2brain/beeep"
	"net"
	"strconv"
	"time"
	shared "tinc-web-boot/web/shared"
)

type uiRoutes struct {
	key           string
	port          uint16
	publicAddress []string
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

func (srv *uiRoutes) Endpoints() ([]shared.Endpoint, error) {
	var ans []shared.Endpoint
	list, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		addrs, err := item.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			if ip, ok := addr.(*net.IPAddr); ok {
				if v4 := ip.IP.To4(); v4 != nil {
					ans = append(ans, shared.Endpoint{
						Host: v4.String(),
						Port: srv.port,
						Kind: shared.Local,
					})
				}
			}
		}
	}
	for _, pub := range srv.publicAddress {
		host, port, err := net.SplitHostPort(pub)
		if err != nil {
			return nil, err
		}
		portNum, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}

		ans = append(ans, shared.Endpoint{
			Host: host,
			Port: uint16(portNum),
			Kind: shared.Public,
		})
	}
	return ans, nil
}
