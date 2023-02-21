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

package logger

import (
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Level zap.AtomicLevel
	Log   *zap.Logger
	Cfg   *zap.Logger
	Api   *zap.Logger
	Init  *zap.Logger
	Prod  *zap.Logger
	Ctx   *zap.Logger
	Util  *zap.Logger
	Gin   *zap.Logger
)

const (
	LOGDIR   = "/logs/"
	FILENAME = "logfile.txt"
)

// The Middleware will write the Gin logs to Zap.
func ginToZap(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Sugar().Infof("| %3d | %15s | %-7s | %s | %s",
			statusCode, clientIP, method, path, errorMessage)
	}
}

func getZapEncoder(env string) zapcore.Encoder {
	if env == "dev" {
		encCfg := zap.NewDevelopmentEncoderConfig()
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encCfg.ConsoleSeparator = " " // Default is tab
		return zapcore.NewConsoleEncoder(encCfg)
	} else if env == "prod" {
		encCfg := zap.NewProductionEncoderConfig()
		return zapcore.NewJSONEncoder(encCfg)
	}
	return nil
}

func getLogWriter(env string) zapcore.WriteSyncer {
	file := os.Stderr
	if env == "prod" {
		path, err := os.Getwd()
		if err != nil {
			return zapcore.AddSync(file)
		}
		dirName := path + LOGDIR
		if _, err := os.Stat(dirName); os.IsNotExist(err) {
			if err = os.Mkdir(dirName, os.ModePerm); err != nil {
				return zapcore.AddSync(file)
			}
		}
		// Dir exists or created by this point
		fileName := dirName + FILENAME
		if f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			return zapcore.NewMultiWriteSyncer(zapcore.AddSync(file), zapcore.AddSync(f)) // Writes to file and stderr - both
		}
	}
	return zapcore.AddSync(file)
}

func Initialize(env string) {
	Level = zap.NewAtomicLevel() // Info logging level enabled by default
	Log = zap.New(zapcore.NewCore(getZapEncoder(env), getLogWriter(env), Level))

	Cfg = Log.Named("[Cfg]")
	Init = Log.Named("[Init]")
	Prod = Log.Named("[Producer]")
	Util = Log.Named("[Util]")
	Api = Log.Named("[Api]")
	Ctx = Log.Named("[Ctx]")
	Gin = Log.Named("[Gin]")

}

func SetLogLevel(l string) {
	var level zapcore.Level
	level.Set(l)
	Level.SetLevel(level)
}

func LogString(key, val string) zapcore.Field {
	return zap.String(key, val)
}

func LogInt(key string, val int) zapcore.Field {
	return zap.Int(key, val)
}

// Returns a new router with Zap logger middleware
func NewRouterWithLogger(log *zap.Logger) *gin.Engine {
	engine := gin.New()
	engine.Use(ginToZap(log), gin.Recovery())

	return engine
}
