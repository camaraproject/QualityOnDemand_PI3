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

package qodContext

import (
	"strconv"
	"time"

	"github.com/sfnuser/dbapi"
	"github.com/sfnuser/qodservice/factory"
	"github.com/sfnuser/qodservice/logger"
)

type OAuth2ServiceCfg struct {
	AuthServerURL   string
	IssuerURL       string
	CacheDuration   time.Duration
	Audience        []string
	AuthorizedScope []string
}

type OAuth2ClientCfg struct {
	TokenURL     string
	ClientId     string
	ClientSecret string
}

// Running Bsf intance qodContext. Any param that is global to
// replicas of this instance needs to be stored in DB
type QodContext struct {
	CompName               string
	Port                   int
	NotifyPort             int
	RegisterDomainName     string
	BindingDomainName      string
	UriScheme              string
	ServiceUrl             string
	NotificationServiceUrl string
	NefScheme              string
	NefServiceDomainName   string
	NefPort                int
	NefServiceName         string
	NefServiceUrl          string
	NefSuppFeat            string
	NefHttpTimeoutSecs     int
	OAuth2Srv              *OAuth2ServiceCfg
	OAuth2Cli              *OAuth2ClientCfg
	Db                     *dbapi.DbApi
}

var qodContext QodContext

func InitQodContext() (err error) {
	config := factory.QodConfig
	configuration := config.Configuration

	logger.SetLogLevel(factory.QodConfig.Logger.QodService.LogLevel)

	db := config.Configuration.Db
	// Connect to DB
	qodContext.Db = dbapi.NewDbApi(db.Name, db.Url)
	qodContext.Db.Connect()

	qodContext.CompName = configuration.CompName
	qodContext.UriScheme = configuration.Service.Scheme              // default uri scheme
	qodContext.RegisterDomainName = factory.QOD_DEFAULT_BINDING_IPV4 // default localhost
	qodContext.Port = factory.QOD_DEFAULT_PORT_INT                   // default port
	qodContext.BindingDomainName = factory.QOD_DEFAULT_BINDING_IPV4
	qodContext.NefScheme = "http"
	qodContext.NefServiceDomainName = factory.QOD_DEFAULT_NEF_IPV4
	qodContext.NefServiceName = factory.QOD_DEFAULT_NEF_SERVICE
	qodContext.NefSuppFeat = factory.QOD_DEFAULT_NEF_SUPP_FEAT
	qodContext.NefHttpTimeoutSecs = factory.QOD_DEFAULT_NEF_HTTP_TIMEOUT_SECS

	service := configuration.Service
	if service != nil {
		if service.RegisterDomainName != "" {
			qodContext.RegisterDomainName = service.RegisterDomainName
		}
		if service.Port != 0 {
			qodContext.Port = service.Port
		}
		if service.NotifyPort != 0 {
			qodContext.NotifyPort = service.NotifyPort
		}
		if service.BindingDomainName != "" {
			qodContext.BindingDomainName = service.BindingDomainName
		}
	}
	nef := configuration.Nef
	if nef != nil {
		if nef.Scheme != "" {
			qodContext.NefScheme = nef.Scheme
		}
		if nef.Port != 0 {
			qodContext.NefPort = nef.Port
		}
		if nef.ServiceDomainName != "" {
			qodContext.NefServiceDomainName = nef.ServiceDomainName
		}
		if nef.ServiceName != "" {
			qodContext.NefServiceName = nef.ServiceName
		}
		if nef.TimeoutSecs != 0 {
			qodContext.NefHttpTimeoutSecs = nef.TimeoutSecs
		}
	}
	if configuration.OAuth2Srv != nil {
		qodContext.OAuth2Srv = &OAuth2ServiceCfg{
			AuthServerURL:   configuration.OAuth2Srv.AuthServerUrl,
			IssuerURL:       configuration.OAuth2Srv.IssuerUrl,
			Audience:        make([]string, len(configuration.OAuth2Srv.Audience)),
			AuthorizedScope: make([]string, len(configuration.OAuth2Srv.AuthorizedScope)),
		}
		copy(qodContext.OAuth2Srv.Audience, configuration.OAuth2Srv.Audience)
		copy(qodContext.OAuth2Srv.AuthorizedScope, configuration.OAuth2Srv.AuthorizedScope)
		if configuration.OAuth2Srv.CacheDuration == 0 {
			qodContext.OAuth2Srv.CacheDuration = factory.QOD_DEFAULT_OAUTH_KEY_CACHE_DURATION_MINS
		}
	} else {
		// Making OAuth2 mandatory
		logger.Ctx.Sugar().Fatal("OAuth2Service config missing")
	}
	if configuration.OAuth2Cli != nil {
		qodContext.OAuth2Cli = &OAuth2ClientCfg{
			TokenURL:     configuration.OAuth2Cli.TokenURL,
			ClientId:     configuration.OAuth2Cli.ClientId,
			ClientSecret: configuration.OAuth2Cli.ClientSecret,
		}
	} else {
		// Making OAuth2 mandatory
		logger.Ctx.Sugar().Fatal("OAuth2Client config missing")
	}

	qodContext.ServiceUrl = string(qodContext.UriScheme) + "://" + qodContext.RegisterDomainName + ":" + strconv.Itoa(qodContext.Port) +
		factory.QOD_DEFAULT_SERVICE
	// Only populate NotificationServiceUrl when NotifyPort is provided
	if qodContext.NotifyPort != 0 {
		qodContext.NotificationServiceUrl = string(qodContext.UriScheme) + "://" + qodContext.RegisterDomainName + ":" +
			strconv.Itoa(qodContext.NotifyPort) + factory.QOD_DEFAULT_NOTIFICATION_SERVICE
	}
	// Service name will be used in checking the client API
	if qodContext.NefPort != 0 {
		qodContext.NefServiceUrl = qodContext.NefScheme + "://" + qodContext.NefServiceDomainName + ":" + strconv.Itoa(qodContext.NefPort)
	} else {
		// Use default scheme ports
		qodContext.NefServiceUrl = qodContext.NefScheme + "://" + qodContext.NefServiceDomainName
	}

	logger.Ctx.Info("Init:", logger.LogString("CompName:", qodContext.CompName), logger.LogString("QodServiceUrl:", qodContext.ServiceUrl),
		logger.LogString("NefServiceUrl:", qodContext.NefServiceUrl))
	return
}

func GetSelf() *QodContext {
	return &qodContext
}

func Terminate() {
	qodContext.Db.Disconnect()
}
