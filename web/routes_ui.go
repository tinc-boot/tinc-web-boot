package web

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gen2brain/beeep"
	"net"
	"strconv"
	"time"
	"tinc-web-boot/tincd"
	shared "tinc-web-boot/web/shared"
)

type uiRoutes struct {
	key           string
	port          uint16
	publicAddress []string
	config        shared.Config
	pool          *tincd.Tincd
}

func (srv *uiRoutes) issueToken(duration time.Duration, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":  time.Now().Add(duration),
		"role": role,
	})
	return token.SignedString([]byte(srv.key))
}

func (srv *uiRoutes) IssueAccessToken(validDays uint) (string, error) {
	return srv.issueToken(time.Duration(24*validDays)*time.Hour, "admin")
}

func (srv *uiRoutes) Notify(title, message string) (bool, error) {
	err := beeep.Notify(title, message, "")
	return err == nil, err
}

func (srv *uiRoutes) Endpoints() ([]shared.Endpoint, error) {
	var ans = make([]shared.Endpoint, 0)
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

			if ip, ok := addr.(*net.IPNet); ok {
				fmt.Println(ip)
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

func (srv *uiRoutes) Configuration() (*shared.Config, error) {
	return &srv.config, nil
}
