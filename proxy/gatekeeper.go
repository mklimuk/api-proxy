package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mklimuk/goerr"
)

type checkToken struct {
	Token  string `json:"token"`
	Update bool   `json:"update"`
	Claims claims `json:"claims,omitempty"`
}

type claims struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	Permissions int    `json:"permissions"`
}

//Gatekeeper is responsible for checking access privileges for an API
type Gatekeeper interface {
	CheckAccess(token string, accessPrivileges int, updateToken bool) (string, error)
}

//NewGatekeeper is the gatekeeper constructor
func NewGatekeeper(auth *url.URL) Gatekeeper {
	k := keeper{auth, &http.Client{Timeout: 10 * time.Second}}
	return Gatekeeper(&k)
}

type keeper struct {
	auth   *url.URL
	client *http.Client
}

func (k *keeper) CheckAccess(token string, accessPrivileges int, updateToken bool) (string, error) {
	if token == "" {
		if accessPrivileges > 0 {
			return token, goerr.NewError("Authorization token required but not present", goerr.Unauthorized)
		}
		return token, nil
	}
	//call authentication service to check the token and compare privileges afterwards
	req := &checkToken{
		Token:  token,
		Update: updateToken,
	}
	var b []byte
	var err error
	if b, err = json.Marshal(&req); err != nil {
		return token, err
	}

	var res *http.Response
	if res, err = k.client.Post(fmt.Sprintf("%s%s", k.auth.String(), "/token/check"), "application/x.token.check+json", bytes.NewReader(b)); err != nil {
		return token, err
	}

	if res.StatusCode != 200 {
		log.WithFields(log.Fields{"logger": "api-proxy.gatekeeper", "method": "CheckAccess", "status": res.StatusCode}).
			Error("Got invalid status code from auth service")
		return token, goerr.NewError("Got invalid status code from auth service", goerr.Unauthorized)
	}

	if err = json.NewDecoder(res.Body).Decode(req); err != nil {
		return token, err
	}

	if req.Claims.Permissions < accessPrivileges {
		return req.Token, goerr.NewError("Too low privileges", goerr.Unauthorized)
	}
	return req.Token, err
}
