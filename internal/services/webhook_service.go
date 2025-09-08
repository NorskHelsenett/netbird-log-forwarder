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
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func ProcessTrafficEvent(request apicontracts.TrafficEvent) (any, error) {

	if !SplunktWorthy(request) {
		return nil, fmt.Errorf("not_splunk_worthy")
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

	xlateMap := viper.GetStringMapString("x-late")
	xlateIp := xlateMap[exitNode.Hostname]
	if xlateIp != "" {
		srcIp = xlateIp
	}

	splunkEvent := apicontracts.SplunkTrafficEvent{
		Time:       unixTime,
		Protocol:   protocols.ProtocolsMap[request.Meta.Protocol],
		SrcIP:      srcIp, // xlate
		SrcPort:    srcPortInt,
		SourceName: sourceName,
		Email:      user.Email,
		DstIP:      dstIp,
		DstPort:    dstPortInt,
		ExitNode:   exitNode.Hostname,
		Message:    request.Message,
	}

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

	return request, nil

}

func ProcessAuditEvent(ev apicontracts.AuditEventEnvelope) (any, error) {

	initator := ev.InitiatorID
	target := ev.TargetID
	netbirdInitiatorUser, _ := netbird.GlobalUserCache.GetUserByID(ev.InitiatorID)

	if netbirdInitiatorUser.Name != "" {
		initator = netbirdInitiatorUser.Name
	}

	netbirdTargetUser, _ := netbird.GlobalUserCache.GetUserByID(ev.TargetID)
	if netbirdTargetUser.Name != "" {
		target = netbirdTargetUser.Name
	}

	unixTime := float64(ev.Timestamp.UnixNano()) / 1e9

	fields := []any{
		zap.Float64("time", unixTime),
		zap.String("message", ev.Message),
		zap.String("initiator_id", initator),
		zap.String("target_id", target),
		zap.ByteString("raw_event", ev.Raw),
	}
	for k, v := range ev.Extra {
		fields = append(fields, zap.Any(k, v))
	}

	logger.SplunkAudit.Infow(
		ev.Message,
		fields...,
	)

	return ev, nil

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
	meta := request.Meta
	baselogger := logger.Log.Desugar()

	if meta.DestinationName != "" && meta.Direction == "INGRESS" && meta.DestinationType == "PEER" {
		// if meta.DestinationName != "" && (meta.Direction == "INGRESS" || meta.Direction == "EGRESS") && meta.DestinationType == "PEER" {
		ipString, _, err := net.SplitHostPort(meta.DestinationAddr)
		if err != nil {
			logger.Log.Warnf("Failed to split host and port: %v", err)
		}

		ip := net.ParseIP(ipString)
		if ip == nil {
			logger.Log.Warnf("Invalid IP address: %s", ipString)
		}

		_, network, err := net.ParseCIDR("100.110.0.0/16")
		if err != nil {
			logger.Log.Errorf("Failed to parse CIDR: %v", err)
		}

		if network.Contains(ip) {
			// logger.Log.Infoln("--- Traffic event rejected ---")
			// baselogger.Info("incoming_event", zap.Any("event", request))
			return false
		}
		logger.Log.Infoln("--- Traffic event accepted ---")
		baselogger.Info("incoming_event", zap.Any("event", request))
		return true
	}

	logger.Log.Infoln("--- Traffic event rejected ---")
	baselogger.Info("incoming_event", zap.Any("event", request))
	return false
}
