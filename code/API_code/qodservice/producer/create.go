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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sfnuser/camara/qodmodels/api"
	nefAsqSpec "github.com/sfnuser/nef/assessionwithqos"
	qodContext "github.com/sfnuser/qodservice/context"
	"github.com/sfnuser/qodservice/logger"
	"github.com/sfnuser/qodservice/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func HandleCreateSessionRequest(req *util.CreateSessionReq) *util.CreateSessionResp {
	sessionReq := req.SessionReq
	rsp := util.CreateSessionResp{}

	// These are validated already
	ueIpv4Addr := sessionReq.UeId.Ipv4addr
	asIpv4Addr := sessionReq.AsId.Ipv4addr

	qodCtx := qodContext.GetSelf()
	// Get provisioned data
	asData, err := qodCtx.Db.GetCamaraProvQoDAppServerData(*asIpv4Addr)
	if err != nil {
		logger.Prod.Sugar().Errorf("failed to get prov data for asIpv4Addr %v", *asIpv4Addr)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "INVALID_INPUT",
			Message: fmt.Sprintf("asIpv4Addr %v not provisioned", *asIpv4Addr),
		}
		return &rsp
	}
	// Get the QoSReference and ScsAsId for UeIpv4Addr
	scsAsId := asData.ScsAsId
	qosReference, ok := asData.QoSMap[string(sessionReq.Qos)]
	if !ok {
		logger.Prod.Sugar().Errorf("qosProfile %v not provisioned for asIpv4Addr %v", sessionReq.Qos, *asIpv4Addr)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "INVALID_INPUT",
			Message: fmt.Sprintf("qosProfile %v not provisioned", sessionReq.Qos),
		}
		return &rsp
	}

	var uePortFormat string
	var asPortFormat string
	if sessionReq.UePorts != nil {
		uePortFormat = util.ConvertPortSpecToFilterFormat(sessionReq.UePorts)
	}
	if sessionReq.AsPorts != nil {
		asPortFormat = util.ConvertPortSpecToFilterFormat(sessionReq.AsPorts)
	}
	flowDesc := []string{
		fmt.Sprintf("permit in any from %s %s to %s %s", *ueIpv4Addr, uePortFormat, *asIpv4Addr, asPortFormat),
		fmt.Sprintf("permit out any from %s %s to %s %s", *asIpv4Addr, asPortFormat, *ueIpv4Addr, uePortFormat),
	}
	logger.Prod.Sugar().Debugf("CreateSession: got prov data. ueIpv4Addr %v, asIpv4Addr %v scsAsId %v, qosReference %v, flowDesc %v",
		*ueIpv4Addr, *asIpv4Addr, scsAsId, qosReference, flowDesc)

	// Check for existing sessions
	ueSessions, err := qodCtx.Db.GetAllCamaraQoDServiceUeSession(*ueIpv4Addr, scsAsId, string(sessionReq.Qos))
	if err != nil {
		logger.Prod.Sugar().Errorf("failed to get existing UeSessions. ueIpv4Addr %v, scsAsId %v, qosProfile %v",
			*ueIpv4Addr, scsAsId, sessionReq.Qos)
		// Not a major error. Proceed
	}
	if ueSessions != nil && len(*ueSessions) > 0 {
		for i := 0; i < len(*ueSessions); i++ {
			ueSession := (*ueSessions)[i]
			flowInfo := ueSession.FlowInfo
			// We have other sessions for UeIpv4Addr, scsAsId and qosProfile
			// We just compare the FlowDesc.
			if flowInfo.FlowDescriptions != nil {
				if len(flowDesc) == len(*flowInfo.FlowDescriptions) {
					var mismatch bool = false
					for i := 0; i < len(flowDesc); i++ {
						if flowDesc[i] != (*flowInfo.FlowDescriptions)[i] {
							mismatch = true
							break
						}
					}
					if !mismatch {
						// Exact same flowDescription exists
						logger.Prod.Sugar().Errorf("existing session with same flowDesc %v exist", flowDesc)
						rsp.ErrorInfo = &api.ErrorInfo{
							Code:    "CONFLICT",
							Message: fmt.Sprintf("existing session with same flowDesc %v exist", flowDesc),
						}
						return &rsp
					}
				}
			}
		}
	}

	// Allocate a FlowId
	medCompN := 1 // Always the same medCompN

	// NOTE: We use fNum as an incremented counter for a given UeIpv4Addr, AS combo.
	// We could simply count the existing number of sessions and use that as fNum.
	// THe problem will be when the first flow is deleted and created again, it will
	// use the same fNum as the already established second flow, which is a problem.
	// We rather use fNum as a counter and we have 2 ** 16 flows possible. Here also,
	// there is a possibility of first flow clashing with 65537th flow but these are not
	// practical scenarios. When a UE establishes a sessions, it sets up all the flows
	// with required qos. Any change in QOS will be 'update' of existing flow, for which
	// we don't have an existing procedure in CAMARA. Writing down the corner cases
	// here, so that we can revisit at appropriate time.

	ueFlows, err := qodCtx.Db.GetCamaraQoDServiceIncrementUeFlow(*ueIpv4Addr, scsAsId)
	if err != nil {
		// Unable to get the Flow Number
		logger.Prod.Sugar().Errorf("failed to get the fNum. err %v", err)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "INTERNAL",
			Message: fmt.Sprintf("failed to get the fNum. err %v", err),
		}
		return &rsp
	}

	// Encode FlowId as per 24.008 Section 10.5.1.6.2
	flowId := (medCompN << 16) | int(ueFlows.FlowCounter)

	// Setup OAuth2 client credentials to be accepted by NEF
	oAuth2Cfg := clientcredentials.Config{
		ClientID:     qodCtx.OAuth2Cli.ClientId,
		ClientSecret: qodCtx.OAuth2Cli.ClientSecret,
		TokenURL:     qodCtx.OAuth2Cli.TokenURL,
	}
	// Get a new Context
	tokenSource := oAuth2Cfg.TokenSource(context.Background())
	oAuthCtx := context.WithValue(oauth2.NoContext, nefAsqSpec.ContextOAuth2, tokenSource)

	// Create a AsqRequest with the given details
	nefAsqReq := nefAsqSpec.NewAsSessionWithQoSSubscriptionWithDefaults()
	nefAsqReq.UeIpv4Addr = sessionReq.UeId.Ipv4addr
	nefAsqReq.FlowInfo = &[]nefAsqSpec.FlowInfo{
		{
			FlowId:           int32(flowId),
			FlowDescriptions: &flowDesc,
		},
	}
	nefAsqReq.QosReference = &qosReference
	nefAsqReq.SupportedFeatures = &qodCtx.NefSuppFeat
	// Populate notification destination only if this service can handle it
	if qodCtx.NotificationServiceUrl != "" {
		nefAsqReq.NotificationDestination = qodCtx.NotificationServiceUrl
	}

	// Log the JSON data
	{
		reqBodyStr, err := json.MarshalIndent(nefAsqReq, "", " ")
		if err == nil {
			logger.Prod.Sugar().Debugf("NefAsSessionWithQoSCreate: Req: JSON(asq): %s", reqBodyStr)
		}
	}
	// Make NEF Client request
	configuration := nefAsqSpec.NewConfiguration()
	// Update APIRoot default server path
	server := configuration.Servers[0].Variables["apiRoot"]
	server.DefaultValue = qodCtx.NefServiceUrl
	configuration.Servers[0].Variables["apiRoot"] = server
	configuration.HTTPClient = &http.Client{
		Timeout: time.Second * time.Duration(qodCtx.NefHttpTimeoutSecs),
	}

	cli := nefAsqSpec.NewAPIClient(configuration)
	subsPostReq := cli.AsSessionWithQoSAPISubscriptionLevelPOSTOperationApi.ScsAsIdSubscriptionsPost(oAuthCtx, scsAsId)
	subsPostReq = subsPostReq.AsSessionWithQoSSubscription(*nefAsqReq)
	rspAsq, hdr, err := subsPostReq.Execute()
	if err != nil || hdr == nil {
		var errString string
		if hdr != nil {
			errString = fmt.Sprintf("nef assessionwithqos subscription create failed. http response %v, err %v", hdr.StatusCode, err)
		} else {
			errString = fmt.Sprintf("nef assessionwithqos subscription create failed. err %v", err)
		}
		logger.Prod.Sugar().Errorln(errString)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "INTERNAL",
			Message: errString,
		}
		return &rsp
	}
	// Get the ResourceId in locationHdr & subscriptionId
	locationHdr := hdr.Header.Get("Location")
	subscriptionId, err := util.ExtractSubstr(locationHdr, "subscriptions/", "")
	if err != nil {
		logger.Prod.Sugar().Errorf("NefAsSessionWithQoSSubscriptionCreate failed to get subscriptionId. err %v", err)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "INTERNAL",
			Message: fmt.Sprintf("nef assessionwithqos subscription create failed to get subscriptionId. err %v", err),
		}
		return &rsp
	}
	if hdr.StatusCode != http.StatusCreated {
		logger.Prod.Sugar().Errorf("NefAsSessionWithQoSSubscriptionCreate failed. httpStatusCode %v", hdr.StatusCode)
		rsp.ErrorInfo = &api.ErrorInfo{
			Code:    "INTERNAL",
			Message: fmt.Sprintf("nef assessionwithqos subscription create failed. httpStatusCode %v", hdr.StatusCode),
		}
		return &rsp
	}
	logger.Prod.Sugar().Infof("NefAsSessionWithQoS Subscription Create: Success. SubscriptionId %v, Resource %v, Self %v",
		subscriptionId, locationHdr, rspAsq.Self)

	// We have a valid NEF session created.
	var duration int32 = 86400 // Seconds in 24hrs
	if sessionReq.Duration != nil {
		duration = *sessionReq.Duration
	}
	now := time.Now().Unix()
	rsp.SessionInfo = &api.SessionInfo{
		Duration:              duration,
		StartedAt:             now,
		ExpiresAt:             now + int64(duration),
		Id:                    uuid.New().String(), // UUID format
		UeId:                  sessionReq.UeId,
		AsId:                  sessionReq.AsId,
		Qos:                   sessionReq.Qos,
		UePorts:               sessionReq.UePorts,
		AsPorts:               sessionReq.AsPorts,
		NotificationUri:       sessionReq.NotificationUri,
		NotificationAuthToken: sessionReq.NotificationAuthToken,
	}

	// Update the DB with the new params
	apiData := util.QoDApiSessionInfo{
		UeIpv4Addr:              *ueIpv4Addr,
		ScsAsId:                 scsAsId,
		SessionId:               rsp.SessionInfo.Id,
		NefSubscriptionId:       subscriptionId,
		NefSubscriptionResource: locationHdr,
		QosReference:            qosReference,
		FlowId:                  uint32(flowId),
		FlowDescriptions:        &flowDesc,
		SessionReq:              sessionReq,
		SessionInfo:             rsp.SessionInfo,
	}
	logger.Prod.Sugar().Infof("CreateSession: Success. SubscriptionId %v, SessionId %v", subscriptionId, apiData.SessionId)
	dbData := util.ConvertSpecToDbSessionInfo(&apiData)
	matchCount, err := qodCtx.Db.PutCamaraQoDServiceUeSession(*ueIpv4Addr, apiData.SessionId, dbData)
	if err != nil || matchCount != 0 {
		logger.Prod.Sugar().Errorf("CreateSession: failed in db write. ueIpv4Addr %v, sessionId %v, err %v, matchCount %v",
			*ueIpv4Addr, apiData.SessionId, err, matchCount)
		// Not sure what we can do here. @todo. Should we delete the nef session and return error?
	}
	return &rsp
}
