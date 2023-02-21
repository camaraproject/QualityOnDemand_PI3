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

package producer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sfnuser/camara/qodmodels/api"
	nefAsqSpec "github.com/sfnuser/nef/assessionwithqos"
	qodContext "github.com/sfnuser/qodservice/context"
	"github.com/sfnuser/qodservice/logger"
	"github.com/sfnuser/qodservice/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func HandleDeleteSessionRequest(req *util.DeleteSessionReq) *util.DeleteSessionResp {
	sessionId := req.SessionId
	rsp := util.DeleteSessionResp{}
	qodCtx := qodContext.GetSelf()
	// Check if session exists
	sessionInfo, err := qodCtx.Db.GetCamaraQoDServiceUeSession(sessionId)
	if err != nil {
		logger.Prod.Sugar().Errorf("deleteSession: sessionId %v not found", sessionId)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "NOT_FOUND",
			Message: fmt.Sprintf("sessionId %v does not exist", sessionId),
		}
		return &rsp
	}
	logger.Prod.Sugar().Infow("Delete Session:", "sessionId", sessionId,
		"NEF subscriptionId", sessionInfo.NefSubscriptionId,
		"scsAsId", sessionInfo.ScsAsId)

	// Setup OAuth2 client credentials to be accepted by NEF
	oAuth2Cfg := clientcredentials.Config{
		ClientID:     qodCtx.OAuth2Cli.ClientId,
		ClientSecret: qodCtx.OAuth2Cli.ClientSecret,
		TokenURL:     qodCtx.OAuth2Cli.TokenURL,
	}
	// Get a new Context
	tokenSource := oAuth2Cfg.TokenSource(context.Background())
	oAuthCtx := context.WithValue(oauth2.NoContext, nefAsqSpec.ContextOAuth2, tokenSource)

	configuration := nefAsqSpec.NewConfiguration()
	// Update APIRoot default server path
	server := configuration.Servers[0].Variables["apiRoot"]
	server.DefaultValue = qodCtx.NefServiceUrl
	configuration.Servers[0].Variables["apiRoot"] = server

	configuration.HTTPClient = &http.Client{
		Timeout: time.Second * time.Duration(qodCtx.NefHttpTimeoutSecs),
	}

	cli := nefAsqSpec.NewAPIClient(configuration)
	subsPostDel := cli.AsSessionWithQoSAPISubscriptionLevelDELETEOperationApi.ScsAsIdSubscriptionsSubscriptionIdDelete(oAuthCtx,
		sessionInfo.ScsAsId, sessionInfo.NefSubscriptionId)
	notifData, nefRsp, err := subsPostDel.Execute()
	if err != nil {
		logger.Prod.Sugar().Errorf("NefAsSessionWithQoSSubscriptionDelete failed. err %v", err)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "INTERNAL",
			Message: fmt.Sprintf("nef assessionwithqos subscription delete failed. err %v", err),
		}
		return &rsp
	} else if nefRsp != nil {
		if !(nefRsp.StatusCode == http.StatusNoContent || nefRsp.StatusCode == http.StatusOK) {
			logger.Prod.Sugar().Errorf("NefAsSessionWithQoSSubscriptionDelete failed. statusCode %v", nefRsp.StatusCode)
			rsp.ErrorInfo = &api.ErrorInfo{
				Code:    "INTERNAL",
				Message: fmt.Sprintf("nef assessionwithqos subscription delete failed. err %v", err),
			}
			return &rsp
		}
		// Success case
		// Check if there is an event occurred during delete
		if notifData.Transaction != "" {
			logger.Prod.Sugar().Warnw("Notification event handling on delete is not implemented", "transaction", notifData.Transaction)
		}
		// Delete the session from QoD DB
		matchCount, err := qodCtx.Db.DeleteCamaraQoDServiceUeSession(sessionId)
		if err != nil || matchCount != 1 {
			logger.Prod.Sugar().Errorf("deleteSession: failed to delete sessionId %v from db. err %v", sessionId, err)
			rsp.ErrorInfo = &api.ErrorInfo{
				Code:    "INTERNAL",
				Message: fmt.Sprintf("deleteSession failed to delete db entry. err %v", err),
			}
			return &rsp
		}
		// No error means the operation succeeded
		return &rsp
	}

	// No response but not error
	logger.Prod.Sugar().Errorf("NefAsSessionWithQoSSubscriptionDelete failed. no response")
	rsp.ErrorInfo = &api.ErrorInfo{
		Code:    "INTERNAL",
		Message: fmt.Sprintf("nef assessionwithqos subscription delete failed. no response"),
	}
	return &rsp
}
