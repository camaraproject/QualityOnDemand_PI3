// Copyright 2023 Spry Fox Networks
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"runtime"
)

// These symbols will be populated at build time
var (
	Version    string
	BuildTime  string
	CommitHash string
)

func GetVersionInfo(appName string) (info string) {
	info = fmt.Sprintf("\n\tAppName:    %-10s"+
		"\n\tVersion:    %-10s"+
		"\n\tBuildTime:  %-10s"+
		"\n\tCommitHash: %-10s"+
		"\n\tGo Version: %-10s %s/%s\n",
		appName, Version, BuildTime, CommitHash,
		runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return
}
