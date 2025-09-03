package services

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/NorskHelsenett/netbird-log-forwarder/internal/cache/netbird"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/cache/protocols"
	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/pkg/models/apicontracts"
	"github.com/go-playground/validator/v10"
)

func ProcessTrafficEvent(request apicontracts.TrafficEvent) (any, error) {

	if !SplunktWorthy(request) {
		return nil, fmt.Errorf("skipping event")
	}

	sourcePeer, _ := netbird.GlobalPeerCache.GetPeerByID(request.Meta.SourceID)
	userId := sourcePeer.UserID

	user, _ := netbird.GlobalUserCache.GetUserByID(userId)
	unixTime := float64(request.Timestamp.UnixNano()) / 1e9
	srcIp := strings.Split(request.Meta.SourceAddr, ":")[0]
	srcPort := strings.Split(request.Meta.SourceAddr, ":")[1]
	srcPortInt, _ := strconv.Atoi(srcPort)
	dstIp := strings.Split(request.Meta.DestinationAddr, ":")[0]
	dstPort := strings.Split(request.Meta.DestinationAddr, ":")[1]
	dstPortInt, _ := strconv.Atoi(dstPort)
	sourceName := request.Meta.SourceName

	exitNode, _ := netbird.GlobalPeerCache.GetPeerByID(request.Meta.ReporterID)

	splunkEvent := apicontracts.SplunkTrafficEvent{
		Time:       unixTime,
		Protocol:   protocols.ProtocolsMap[request.Meta.Protocol],
		SrcIP:      srcIp,
		SrcPort:    srcPortInt,
		SourceName: sourceName,
		Email:      user.Email,
		DstIP:      dstIp,
		DstPort:    dstPortInt,
		ExitNode:   exitNode.Hostname,
		Message:    request.Message,
	}

	// jsonBytes, err := json.MarshalIndent(splunkEvent, "", "  ")

	// if err != nil {
	// 	logger.Log.Warnf("Failed to marshal event:", err)
	// } else {
	// 	logger.Log.Infoln(string(jsonBytes))
	// }

	// logger.Log.Infow("netbird traffic event", "event", splunkEvent)
	logger.SplunkTraffic.Infow(
		request.Message,
		"time", splunkEvent.Time,
		"protocol", splunkEvent.Protocol,
		"src_ip", splunkEvent.SrcIP,
		"src_port", splunkEvent.SrcPort,
		"source_name", splunkEvent.SourceName,
		"email", splunkEvent.Email,
		"dst_ip", splunkEvent.DstIP,
		"dst_port", splunkEvent.DstPort,
		"exit_node", splunkEvent.ExitNode,
	)

	// logger.SplunkTraffic.Infow("", splunkEvent)

	return request, nil

}

func ProcessAuditEvent(request apicontracts.AuditEvent) (any, error) {

	user, _ := netbird.GlobalUserCache.GetUserByID(request.InitiatorID)
	unixTime := float64(request.Timestamp.UnixNano()) / 1e9

	splunkEvent := apicontracts.SplunkAuditEvent{
		Time:                 unixTime,
		User:                 user.Name,
		LocationCityName:     request.Meta.LocationCityName,
		LocationCountryCode:  request.Meta.LocationCountryCode,
		LocationConnectionIp: request.Meta.LocationConnectionIp,
		LocationGeoNameId:    request.Meta.LocationGeoNameId,
		Ip:                   request.Meta.Ip,
		Fqdn:                 request.Meta.Fqdn,
	}

	// jsonBytes, err := json.MarshalIndent(request, "", "  ")

	// if err != nil {
	// 	logger.Log.Warnf("Failed to marshal event:", err)
	// } else {
	// 	logger.Log.Infoln(string(jsonBytes))
	// }

	// logger.Log.Infow("netbird traffic event", "event", splunkEvent)
	logger.SplunkAudit.Infow(
		request.Message,
		"time", splunkEvent.Time,
		"user", splunkEvent.User,
		"fqdn", splunkEvent.Fqdn,
		"location_city_name", splunkEvent.LocationCityName,
		"location_country_code", splunkEvent.LocationCountryCode,
		"location_geo_name_id", splunkEvent.LocationGeoNameId,
		"location_connection_ip", splunkEvent.LocationConnectionIp,
		"ip", splunkEvent.Ip,
		// "protocol", splunkEvent.Protocol,
		// "src_ip", splunkEvent.SrcIP,
		// "src_port", splunkEvent.SrcPort,
		// "source_name", splunkEvent.SourceName,
		// "email", splunkEvent.Email,
		// "dst_ip", splunkEvent.DstIP,
		// "dst_port", splunkEvent.DstPort,
		// "exit_node", splunkEvent.ExitNode,
	)

	// logger.SplunkTraffic.Infow("", splunkEvent)

	return request, nil

}

func ValidateRequest(request *apicontracts.TrafficEvent) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(*request)

	if err != nil {
		return err
	}

	return nil
}

func SplunktWorthy(request apicontracts.TrafficEvent) bool {
	if request.Meta == (apicontracts.TrafficMeta{}) {
		return false
	}
	meta := request.Meta

	if meta.Direction != "INGRESS" {
		return false
	}

	if meta.DestinationType != "PEER" {
		return false
	}

	if meta.DestinationName == "" {
		return false
	}

	addrStr := meta.DestinationAddr

	// fmt.Println("Destination Address:", addrStr)

	ipAddress, _, err := net.SplitHostPort(addrStr)
	if err != nil {
		// håndter feil
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		// håndter ugyldig IP
	}

	_, network, err := net.ParseCIDR("100.64.0.0/10")
	if err != nil {
		// håndter feil
	}

	if network.Contains(ip) {
		return false
	}

	return true
}
