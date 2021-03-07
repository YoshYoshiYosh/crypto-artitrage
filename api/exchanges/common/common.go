package common

import (
	"net/http"

	"gopkg.in/ini.v1"
)

var Config, _ = ini.Load("./api/config/config.ini")

type ApiClient struct {
	Key         string
	Secret      string
	HttpClient  *http.Client
	HttpRequest *http.Request
}

type ApiPathAndMethod struct {
	Path   string
	Method string
}
