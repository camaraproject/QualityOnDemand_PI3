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

package oauth2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

// Some defaults when the values are not configured
const (
	defaultCacheDuration time.Duration = 5 * time.Minute
)

// Config related to JWT based OAuth2 Authorization
type Config struct {
	AuthServerURL       string        // URL of the Auth server (e.g. http://oauthserver:8080/realms/sfn.nef for KeyCloak)
	IssuerURL           string        // In case when Auth Server exposes an external or different IssuerURL. If empty AuthServerURL will be used
	PubKeyCacheDuration time.Duration // Duration to store the RSA Pubic Key
	Audience            []string      // The intended Audience the AccessToken should have (as configured in AuthServer)
	AuthorizedScope     []string      // Allowed scopes to be validated against. At this point they MUST have http.Methods to be validated with route
}

type AudienceCustomClaims struct {
	Scope           string   `json:"scope"` // This is a mandatory claim that MUST be present in the token
	authorizedScope []string // Not exported
}

type OAuth2Provider struct {
	Conf Config
}

func (a *AudienceCustomClaims) Validate(ctx context.Context) error {
	scope := strings.Split(a.Scope, " ")
	for i := range scope {
		var scopeValidated bool = false
		for j := range a.authorizedScope {
			if scope[i] == a.authorizedScope[j] {
				scopeValidated = true
				break
			}
		}
		if !scopeValidated {
			return fmt.Errorf("scope %v is not allowed", scope[i])
		}
	}
	return nil
}

func New(conf *Config) (*OAuth2Provider, error) {
	auth := &OAuth2Provider{
		Conf: *conf,
	}
	// Sanity checks
	if conf.AuthServerURL == "" {
		return nil, errors.New("invalid authserverURL")
	}
	if conf.PubKeyCacheDuration == 0 {
		auth.Conf.PubKeyCacheDuration = defaultCacheDuration
	}
	if len(conf.Audience) == 0 {
		return nil, errors.New("invalid audience")
	}
	if conf.AuthorizedScope == nil {
		return nil, errors.New("invalid scope")
	}
	return auth, nil
}

// This can be used as a Middleware or a GinHandlerFunc for a specific route
func (o *OAuth2Provider) AuthorizationMiddleware() (gin.HandlerFunc, error) {
	authServerURL, err := url.Parse(o.Conf.AuthServerURL)
	if err != nil {
		return nil, fmt.Errorf("bad AuthServerURL string %v. err %v", o.Conf.AuthServerURL, err)
	}
	cacheProvider := jwks.NewCachingProvider(authServerURL, o.Conf.PubKeyCacheDuration)

	audienceCustomClaims := func() validator.CustomClaims {
		return &AudienceCustomClaims{
			authorizedScope: o.Conf.AuthorizedScope,
		}
	}
	issuer := authServerURL.String()
	if o.Conf.IssuerURL != "" {
		issuerURL, err := url.Parse(o.Conf.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("bad IssuerURL string %v. err %v", o.Conf.IssuerURL, err)
		}
		issuer = issuerURL.String()
	}

	jwtValidator, err := validator.New(cacheProvider.KeyFunc,
		validator.RS256,
		issuer,
		o.Conf.Audience,
		validator.WithCustomClaims(audienceCustomClaims),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator. err %v", err)
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("error while validating jwt. err %v", err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	ginHandler := func(ctx *gin.Context) {
		procError := true
		// This handler gets called at the end of CheckJWT function
		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			ctx.Request = r

			// At this point CheckJWT validations would be completed.
			// Below is an extra check to compare the http.Method (route) with
			// the claimed scope. For e.g. the method GET should be present in the scope claim.
			// This is to avoid a valid access token with scope say DELETE only, to use the GET route.
			// Without this check, the token & scope are valid and allowed to proceed. But here
			// we tie up the scope with the route
			claims, ok := ctx.Request.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
			if !ok {
				return
			}
			customClaims, ok := claims.CustomClaims.(*AudienceCustomClaims)
			if !ok {
				return
			}
			scope := strings.Split(customClaims.Scope, " ")
			var routeValidated bool = false
			for i := range scope {
				if scope[i] == ctx.Request.Method {
					routeValidated = true
					break
				}
			}
			if !routeValidated {
				return
			}
			// If we are here then the route is validated.
			// procError can be false now and the next gin Handler is called
			procError = false
			ctx.Next()
		}
		middleware.CheckJWT(handler).ServeHTTP(ctx.Writer, ctx.Request)

		// All the gin handlers are executed post validation of JWT and procError must be false at this point
		if procError {
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				map[string]string{"message": "invalid token"},
			)
		}
	}
	return ginHandler, nil
}
