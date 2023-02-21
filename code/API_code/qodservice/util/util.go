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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sfnuser/camara/qodmodels/api"
	"github.com/sfnuser/camara/qodmodels/db"
	"github.com/sfnuser/qodservice/logger"
)

const (
	INVALID_INPUT       string = "INVALID_INPUT"
	UNAUTHORIZED        string = "UNAUTHORIZED"
	FORBIDDEN           string = "FORBIDDEN"
	NOT_FOUND           string = "NOT_FOUND"
	SERVICE_UNAVAILABLE string = "SERVICE_UNAVAILABLE"
	CONFLICT            string = "CONFLICT"
	INTERNAL            string = "INTERNAL"
)

var allowedErrorCodes = []string{
	"INVALID_INPUT",
	"UNAUTHORIZED",
	"FORBIDDEN",
	"NOT_FOUND",
	"SERVICE_UNAVAILABLE",
	"CONFLICT",
	"INTERNAL",
}

type CreateSessionReq struct {
	SessionReq *api.CreateSession
}
type CreateSessionResp struct {
	SessionInfo *api.SessionInfo
	ErrorInfo   *api.ErrorInfo
}
type DeleteSessionReq struct {
	SessionId string
}
type DeleteSessionResp struct {
	ErrorInfo *api.ErrorInfo // If no error then session is deleted successfully
}

type QoDApiSessionInfo struct {
	SessionReq              *api.CreateSession
	SessionInfo             *api.SessionInfo
	UeIpv4Addr              string
	ScsAsId                 string
	SessionId               string
	NefSubscriptionId       string
	NefSubscriptionResource string
	QosReference            string
	FlowId                  uint32
	FlowDescriptions        *[]string
}

func NewQoDErrorInfo(code, message string) []byte {
	errorInfo := api.NewErrorInfo(code, message)
	data, err := json.Marshal(errorInfo)
	if err != nil {
		logger.Util.Sugar().Errorf("failed to encode error detail: %v", err)
	}
	return data
}
func validateUeId(ueId *api.UeId) error {
	if ueId.Ipv4addr == nil {
		//We need mandatory IPv4Addr within ueId at this point
		errString := "ueId did not have mandatory ipv4Addr property"
		logger.Util.Error("error:", logger.LogString("ueId", errString))
		return errors.New(errString)
	}
	if ueId.Ipv6addr != nil {
		logger.Util.Sugar().Warnf("ueId: ipv6addr processing unsupported. ipv6addr %v", ueId.Ipv4addr)
	}
	if ueId.ExternalId != nil {
		logger.Util.Sugar().Warnf("ueId: externalId processing unsupported. externalId %v", ueId.ExternalId)
	}
	if ueId.Msisdn != nil {
		logger.Util.Sugar().Warnf("ueId: msisdn processing unsupported. msisdn %v", ueId.Msisdn)
	}
	return nil
}
func validateAsId(asId *api.AsId) error {
	if asId.Ipv4addr == nil {
		//We need mandatory IPv4Addr within ueId at this point
		errString := "asId did not have mandatory ipv4Addr property"
		logger.Util.Error("error:", logger.LogString("asId", errString))
		return errors.New(errString)
	}
	if asId.Ipv6addr != nil {
		logger.Util.Sugar().Warnf("asId: ipv6addr processing unsupported. ipv6addr %v", asId.Ipv4addr)
	}
	return nil
}
func validateQoS(qos *api.QosProfile) error {
	switch *qos {
	case api.E:
	case api.L:
	case api.M:
	case api.S:
	default:
		errString := "qosProfile invalid"
		logger.Util.Error("error:", logger.LogString("qos", errString))
		return errors.New(errString)
	}
	return nil
}
func isValidPort(port int32) bool {
	if port >= 0 && port <= 65535 {
		return true
	}
	return false
}
func validatePorts(ports *api.PortsSpec) error {
	if ports == nil {
		return nil
	}
	if ports.Ports != nil {
		for _, port := range ports.Ports {
			if ok := isValidPort(port); !ok {
				return fmt.Errorf("Port %v not valid", port)
			}
		}
	}
	if ports.Ranges != nil {
		for _, portRange := range ports.Ranges {
			if ok := isValidPort(portRange.From); !ok {
				return fmt.Errorf("Port range from %v not valid", portRange.From)
			}
			if ok := isValidPort(portRange.To); !ok {
				return fmt.Errorf("Port range to %v not valid", portRange.To)
			}
			if portRange.From > portRange.To {
				return fmt.Errorf("Port range from %v to %v not valid", portRange.From, portRange.To)
			}
		}
	}
	return nil
}
func validateUePorts(ports *api.PortsSpec) error {
	err := validatePorts(ports)
	if err != nil {
		logger.Util.Sugar().Errorf("ueportspec: %v", err)
	}
	return err
}
func validateAsPorts(ports *api.PortsSpec) error {
	err := validatePorts(ports)
	if err != nil {
		logger.Util.Sugar().Errorf("asportspec: %v", err)
	}
	return err
}
func ValidateSessionReq(sessionReq *api.CreateSession) error {
	// Check IEs validity
	err := validateUeId(&sessionReq.UeId)
	if err == nil {
		err = validateAsId(&sessionReq.AsId)
		if err == nil {
			err = validateQoS(&sessionReq.Qos)
			if err == nil {
				err = validateUePorts(sessionReq.UePorts)
				if err == nil {
					err = validateAsPorts(sessionReq.AsPorts)
				}
			}
		}
	}
	return err
}
func ConvertErrorToHttpStatusCode(errCode string) int {
	switch errCode {
	case INVALID_INPUT:
		return http.StatusBadRequest
	case UNAUTHORIZED:
		return http.StatusUnauthorized
	case FORBIDDEN:
		return http.StatusForbidden
	case NOT_FOUND:
		return http.StatusNotFound
	case SERVICE_UNAVAILABLE:
		return http.StatusServiceUnavailable
	case CONFLICT:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
func ConvertPortSpecToFilterFormat(ports *api.PortsSpec) string {
	if ports == nil {
		return ""
	}
	var portFmt string
	if ports.Ranges != nil {
		for _, portRange := range ports.Ranges {
			portFmt += strconv.Itoa(int(portRange.From)) + "-" + strconv.Itoa(int(portRange.To)) + ", "
		}
	}
	if ports.Ports != nil {
		for _, port := range ports.Ports {
			portFmt += strconv.Itoa(int(port)) + ", "
		}
	}
	portFmt = strings.TrimSuffix(portFmt, ", ")
	logger.Util.Sugar().Debugf("ConvPortSpecToFilter: portFmt %v", portFmt)
	return portFmt
}
func ConvertSpecToDbSessionInfo(inSession *QoDApiSessionInfo) *db.ServiceQoDUeSession {
	session := db.ServiceQoDUeSession{
		UeIpv4Addr:              inSession.UeIpv4Addr,
		ScsAsId:                 inSession.ScsAsId,
		SessionId:               inSession.SessionId,
		NefSubscriptionId:       inSession.NefSubscriptionId,
		NefSubscriptionResource: inSession.NefSubscriptionResource,
		QosReference:            inSession.QosReference,
		FlowInfo: db.FlowInfo{
			FlowId:           inSession.FlowId,
			FlowDescriptions: inSession.FlowDescriptions,
		},
	}
	dbSessReq := &session.SessionReq
	dbSessInfo := &session.SessionInfo
	inSessReq := inSession.SessionReq

	// Copy SessionReq data
	dbSessReq.AsId.Ipv4addr = inSessReq.AsId.Ipv4addr
	dbSessReq.UeId.Ipv4addr = inSessReq.UeId.Ipv4addr
	dbSessReq.Duration = inSessReq.Duration
	if inSessReq.UePorts != nil {
		dbSessReq.UePorts = inSessReq.UePorts
		dbSessInfo.UePorts = dbSessReq.UePorts
	}
	if inSessReq.AsPorts != nil {
		dbSessReq.AsPorts = inSessReq.AsPorts
		dbSessInfo.AsPorts = dbSessReq.AsPorts
	}
	dbSessReq.Qos = inSessReq.Qos
	dbSessReq.NotificationUri = inSessReq.NotificationUri
	dbSessReq.NotificationAuthToken = inSessReq.NotificationAuthToken

	// Copy SessionInfo data
	dbSessInfo.Id = inSession.SessionInfo.Id
	dbSessInfo.Duration = inSession.SessionInfo.Duration
	dbSessInfo.StartedAt = inSession.SessionInfo.StartedAt
	dbSessInfo.ExpiresAt = inSession.SessionInfo.ExpiresAt

	// Copy information from SessionReq to SessionInfo

	dbSessInfo.AsId.Ipv4addr = dbSessReq.AsId.Ipv4addr
	dbSessInfo.UeId.Ipv4addr = dbSessReq.UeId.Ipv4addr
	dbSessInfo.Qos = dbSessReq.Qos
	dbSessInfo.NotificationUri = dbSessReq.NotificationUri
	dbSessInfo.NotificationAuthToken = dbSessReq.NotificationAuthToken

	return &session
}
func ExtractSubstr(sourceStr, startMarkerStr, endMarkerStr string) (string, error) {
	if idx := strings.Index(sourceStr, startMarkerStr); idx >= 0 {
		result := sourceStr[idx+len(startMarkerStr):]
		if endMarkerStr != "" {
			if idx := strings.Index(result, endMarkerStr); idx >= 0 {
				result = result[:idx]
				return result, nil
			}
		} else {
			return result, nil
		}
	}
	return "", fmt.Errorf("extractSubstr failed. sourceStr %s, startMarkerStr %s, endMarkerStr %s", sourceStr, startMarkerStr, endMarkerStr)
}
