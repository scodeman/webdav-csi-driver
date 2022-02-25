/*
Original code:
- https://github.com/cyverse/irods-csi-driver/blob/master/pkg/driver/irods_secret.go
*/
package driver

import (
	"io/ioutil"
	"path"
	"strings"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

// ReadSecrets reads secrets from secret volume mount
func ReadSecrets(secretPath string) (map[string]string, error) {
	exist, err := PathExists(secretPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Secret path %s does not exist - %s", secretPath, err)
	}

	if !exist {
		return nil, status.Errorf(codes.NotFound, "Secret path %s does not exist", secretPath)
	}

	secrets := make(map[string]string)

	files, err := ioutil.ReadDir(secretPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			k := path.Base(file.Name())
			fullPath := path.Join(secretPath, k)
			content, readErr := ioutil.ReadFile(fullPath)
			if readErr == nil {
				contentString := string(content)
				v := strings.TrimSpace(contentString)
				secrets[k] = v
			}
		}
	}
	return secrets, nil
}
