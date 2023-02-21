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

package service

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/urfave/cli/v2"

	qodContext "github.com/sfnuser/qodservice/context"
	"github.com/sfnuser/qodservice/factory"
	"github.com/sfnuser/qodservice/logger"
	"github.com/sfnuser/qodservice/oauth2"
	"github.com/sfnuser/qodservice/qodapi"
	"github.com/sfnuser/qodservice/util"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type QoD struct{}

type Config struct {
	qodCfg string
}

var config Config

var qodCli = []cli.Flag{
	&cli.StringFlag{
		Name:  "qodservice_cfg",
		Usage: "qodapi config file",
	},
	&cli.StringFlag{
		Name:    "logmode",
		Usage:   "dev/prod log mode",
		Value:   "dev",
		EnvVars: []string{"QODAPI_LOGMODE"},
	},
}

func (*QoD) GetCliCmd() (flags []cli.Flag) {
	return qodCli
}

func StartHttpsServer(server *http.Server, env, srvDomainName string) (err error) {
	logger.Init.Sugar().Infof("Attempting https: env %s", env)

	c, err := util.GetTlsCredentialsWithoutRootCA(env, srvDomainName)
	if err != nil {
		return fmt.Errorf("failed to get TLS credentials. error %v", err)
	}

	cert, err := tls.X509KeyPair(c.GetMyCert(), c.GetMyKey())
	if err != nil {
		return fmt.Errorf("failed to get X509KeyPair. error %v", err)
	}

	// Add the server credential
	server.TLSConfig.Certificates = []tls.Certificate{
		cert,
	}

	return server.ListenAndServeTLS("", "") // Cert & Key are already added
}

func NewServer(addr string, handler http.Handler) (srv *http.Server) {
	srv = &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}
	return
}

func NewQoD() *QoD {
	return &QoD{}
}

func (q *QoD) Initialize(c *cli.Context) (err error) {
	// Read the Config
	config = Config{
		qodCfg: c.String("qodservice_cfg"), // QoDAPI_P config
	}

	if config.qodCfg != "" {
		if err := factory.InitConfigFactory(config.qodCfg); err != nil {
			return err
		}
	} else {
		DefaultQodCfgConfigPath := "config/qodservice_cfg.yaml" // QoDAPI service config
		if err := factory.InitConfigFactory(DefaultQodCfgConfigPath); err != nil {
			return err
		}
	}
	return nil
}

func (q *QoD) Start(c *cli.Context) {
	// Initialize the logger first
	logger.Initialize(c.String("logmode"))
	logger.Init.Sugar().Infof("VersionInfo: %s", util.GetVersionInfo(c.App.Name))
	router := logger.NewRouterWithLogger(logger.Gin)

	// Init QoD context
	err := qodContext.InitQodContext()
	if err != nil {
		logger.Init.Sugar().Fatalf("failed to init qodContext. err %v", err)
	}

	// Authorization middleware
	context := qodContext.GetSelf()
	oAuthConfig := oauth2.Config{
		AuthServerURL:       context.OAuth2Srv.AuthServerURL,
		IssuerURL:           context.OAuth2Srv.IssuerURL,
		PubKeyCacheDuration: context.OAuth2Srv.CacheDuration,
		Audience:            make([]string, len(context.OAuth2Srv.Audience)),
		AuthorizedScope:     make([]string, len(context.OAuth2Srv.AuthorizedScope)),
	}
	copy(oAuthConfig.Audience, context.OAuth2Srv.Audience)
	copy(oAuthConfig.AuthorizedScope, context.OAuth2Srv.AuthorizedScope)
	logger.Init.Sugar().Infof("OAuth2: Config %v", oAuthConfig)
	oAuth, err := oauth2.New(&oAuthConfig)
	if err != nil {
		logger.Init.Sugar().Fatalf("oauth2 config incorrect. conf %v", oAuthConfig)
	}
	authMiddlewareHandler, err := oAuth.AuthorizationMiddleware()
	if err != nil {
		logger.Init.Sugar().Fatalf("failed to setup OAuth2Middleware. err %v", err)
	}
	router.Use(authMiddlewareHandler)

	router.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "DELETE"},
		AllowHeaders: []string{
			"Authorization", "Origin", "Content-Length", "Content-Type", "User-Agent",
			"Referrer", "Host", "Token", "X-Requested-With",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           86400,
	}))

	// Add service handlers
	qodapi.AddService(router)

	// Handle Ctrl+C to gracefully terminate
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func(c *cli.Context) {
		<-signalChannel
		q.Terminate(c)
		os.Exit(0)
	}(c)

	// Start the server
	addr := fmt.Sprintf("%s:%d", qodContext.GetSelf().BindingDomainName, qodContext.GetSelf().Port)
	server := NewServer(addr, router)
	if factory.QodConfig.Configuration.Service.Scheme == "http" {
		err = server.ListenAndServe()
	} else if factory.QodConfig.Configuration.Service.Scheme == "https" {
		err = StartHttpsServer(server, factory.QodConfig.Configuration.Service.Env, factory.QodConfig.Configuration.Service.BindingDomainName)
	}

	if err != nil {
		logger.Init.Sugar().Fatalf("failed to start server. err %v", err)
	}
	logger.Init.Sugar().Infof("%s: Started", c.App.Name)
	return
}

func (q *QoD) Terminate(c *cli.Context) {
	logger.Init.Sugar().Infof("%s: Terminated", c.App.Name)

	qodContext.Terminate()
}
