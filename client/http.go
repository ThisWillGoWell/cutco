package client

import (
	"github.com/Yamashou/gqlgenc/client"
	"net/http"
)

func AddToken(token string) client.HTTPRequestOption {
	return func(req *http.Request) {
		req.Header.Set("Authorization", token)
	}
}
