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

package factory

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var QodConfig Config

const (
	QOD_DEFAULT_REGISTER_IPV4         = "127.0.0.1"
	QOD_DEFAULT_BINDING_IPV4          = "0.0.0.0"
	QOD_DEFAULT_PORT_INT              = 9000
	QOD_DEFAULT_NOTIFY_PORT_INT       = 9001
	QOD_DEFAULT_SERVICE               = "/qod/v0"
	QOD_DEFAULT_NOTIFICATION_SERVICE  = "/qod/callback/v0"
	QOD_DEFAULT_NEF_IPV4              = "127.0.0.1"
	QOD_DEFAULT_NEF_SERVICE           = "/3gpp-as-session-with-qos/v1" // QoS Service
	QOD_DEFAULT_NEF_SUPP_FEAT         = "0"
	QOD_DEFAULT_NEF_HTTP_TIMEOUT_SECS = 5 // secs

	QOD_DEFAULT_OAUTH_KEY_CACHE_DURATION_MINS = 5
)

func InitConfigFactory(f string) error {
	if content, err := ioutil.ReadFile(f); err != nil {
		return err
	} else {
		QodConfig = Config{}

		if yamlErr := yaml.Unmarshal(content, &QodConfig); yamlErr != nil {
			return yamlErr
		}
	}
	return nil
}
