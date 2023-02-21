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

type Config struct {
	Configuration *Configuration `yaml:"configuration"`
	Logger        *LogComponents `yaml:"logger"`
}

type LogSetting struct {
	LogLevel string `yaml:"logLevel"`
}

type LogComponents struct {
	QodService *LogSetting `yaml:"qodService"`

	// Add more as needed
}

type Configuration struct {
	CompName  string         `yaml:"compName"`
	Service   *Service       `yaml:"service"`
	Nef       *Nef           `yaml:"nef"`
	OAuth2Srv *OAuth2Service `yaml:"oauth2Service"` // QoD's OAuth2 service configuration (incoming requests towards QoD)
	OAuth2Cli *OAuth2Client  `yaml:"oauth2Client"`  // QoD's outgoing request towards NEF
	Db        *Db            `yaml:"db"`
}

type Service struct {
	Scheme             string `yaml:"scheme"`
	RegisterDomainName string `yaml:"registerDomainName"` // IP/DomainName that is registered at NRF.
	BindingDomainName  string `yaml:"bindingDomainName"`  // IP/DomainName used to run the server in the node.
	Port               int    `yaml:"port"`
	NotifyPort         int    `yaml:"notifyPort,omitempty"` // If notifyPort is not provided then QoD will not subscribe to events from NEF
	Env                string `yaml:"env"`                  // The cert & key are in local dir or azure cloud
}

type Nef struct {
	Scheme            string `yaml:"scheme"`
	ServiceDomainName string `yaml:"serviceDomainName"` // Service domain name
	ServiceName       string `yaml:"serviceName"`       // Service base URL
	Port              int    `yaml:"port,omitempty"`
	SuppFeat          string `yaml:"suppFeatures"`
	TimeoutSecs       int    `yaml:"timeoutSecs,omitempty"`
}

type OAuth2Service struct {
	AuthServerUrl   string   `yaml:"authServerUrl"`
	IssuerUrl       string   `yaml:"issuerUrl,omitempty"`
	CacheDuration   int      `yaml:"cacheDuration,omitempty"`
	Audience        []string `yaml:"audience"`
	AuthorizedScope []string `yaml:"authorizedScope"`
}

type OAuth2Client struct {
	TokenURL     string `yaml:"tokenUrl"`
	ClientId     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}
type Db struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}
