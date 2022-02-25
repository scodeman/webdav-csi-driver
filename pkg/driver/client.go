/*
Original code:
- https://github.com/cyverse/irods-csi-driver/blob/master/pkg/driver/irods_client.go
*/
package driver

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClientType is a mount client type
type ClientType string

// mount driver (Client) types
const (
	// WebdavType is for WebDav client (Davfs2)
	WebdavType ClientType = "webdav"
)

// WebDAVConnectionInfo class
type WebDAVConnectionInfo struct {
	URL      string
	User     string
	Password string
}

// NewWebDAVConnectionInfo returns a new instance of WebDAVConnectionInfo
func NewWebDAVConnectionInfo(url string, user string, password string) *WebDAVConnectionInfo {
	return &WebDAVConnectionInfo{
		URL:      url,
		User:     user,
		Password: password,
	}
}

// ExtractClientType extracts  Client value from param map
func ExtractClientType(params map[string]string, secrets map[string]string, defaultClient ClientType) ClientType {
	client := ""
	for k, v := range secrets {
		if strings.ToLower(k) == "driver" || strings.ToLower(k) == "client" {
			client = v
			break
		}
	}

	for k, v := range params {
		if strings.ToLower(k) == "driver" || strings.ToLower(k) == "client" {
			client = v
			break
		}
	}

	return GetValidClientType(client, defaultClient)
}

// IsValidClientType checks if given client string is valid
func IsValidClientType(client string) bool {
	switch client {
	case string(WebdavType):
		return true
	default:
		return false
	}
}

// GetValidClientType checks if given client string is valid
func GetValidClientType(client string, defaultClient ClientType) ClientType {
	switch client {
	case string(WebdavType):
		return WebdavType
	default:
		return defaultClient
	}
}

// ExtractWebDAVConnectionInfo extracts WebDAVConnectionInfo value from param map
func ExtractWebDAVConnectionInfo(params map[string]string, secrets map[string]string) (*WebDAVConnectionInfo, error) {
	var user, password, url string

	for k, v := range secrets {
		switch strings.ToLower(k) {
		case "user":
			user = v
		case "password":
			password = v
		case "url":
			url = v
		default:
			// ignore
		}
	}

	for k, v := range params {
		switch strings.ToLower(k) {
		case "user":
			user = v
		case "password":
			password = v
		case "url":
			url = v
		default:
			// ignore
		}
	}

	// user and password fields are optional
	// if user is not given, it is regarded as anonymous user
	if len(user) == 0 {
		user = "anonymous"
	}

	// password can be empty for anonymous access
	if len(password) == 0 && user != "anonymous" {
		return nil, status.Error(codes.InvalidArgument, "Argument password is empty")
	}

	if len(url) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Argument url is empty")
	}

	conn := NewWebDAVConnectionInfo(url, user, password)
	return conn, nil
}
