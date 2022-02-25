/*
Original code:
- https://github.com/cyverse/irods-csi-driver/blob/master/pkg/driver/version.go

Following functions or objects are from the code under APL2 License.
- VersionInfo
- GetVersion
- GetVersionJSON
Original code: https://github.com/kubernetes-sigs/aws-efs-csi-driver/blob/master/pkg/driver/version.go


Copyright 2019 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	driverVersion string
	gitCommit     string
	buildDate     string
)

// VersionInfo object contains version related info
type VersionInfo struct {
	DriverVersion string `json:"driverVersion"`
	GitCommit     string `json:"gitCommit"`
	BuildDate     string `json:"buildDate"`
	GoVersion     string `json:"goVersion"`
	Compiler      string `json:"compiler"`
	Platform      string `json:"platform"`
}

// GetVersion returns VersionInfo object
func GetVersion() VersionInfo {
	return VersionInfo{
		DriverVersion: driverVersion,
		GitCommit:     gitCommit,
		BuildDate:     buildDate,
		GoVersion:     runtime.Version(),
		Compiler:      runtime.Compiler,
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetDriverVersion returns driver version in string
func GetDriverVersion() string {
	return driverVersion
}

// GetVersionJSON returns VersionInfo object in JSON string
func GetVersionJSON() (string, error) {
	info := GetVersion()
	marshalled, err := json.MarshalIndent(&info, "", "  ")
	if err != nil {
		return "", err
	}
	return string(marshalled), nil
}
