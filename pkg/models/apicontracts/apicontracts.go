package apicontracts

import "time"

type TrafficMeta struct {
	DestinationAddr       string `json:"destination_addr"`
	DestinationDNSLabel   string `json:"destination_dns_label"`
	DestinationGeoCity    string `json:"destination_geo_city"`
	DestinationGeoCountry string `json:"destination_geo_country"`
	DestinationID         string `json:"destination_id"`
	DestinationName       string `json:"destination_name"`
	DestinationType       string `json:"destination_type"`
	Direction             string `json:"direction"`
	FlowID                string `json:"flow_id"`
	ICMPCode              int    `json:"icmp_code"`
	ICMPType              int    `json:"icmp_type"`
	PolicyID              string `json:"policy_id"`
	PolicyName            string `json:"policy_name"`
	Protocol              int    `json:"protocol"`
	ReceivedTimestamp     string `json:"received_timestamp"`
	ReporterID            string `json:"reporter_id"`
	RxBytes               int    `json:"rx_bytes"`
	RxPackets             int    `json:"rx_packets"`
	SourceAddr            string `json:"source_addr"`
	SourceDNSLabel        string `json:"source_dns_label"`
	SourceGeoCity         string `json:"source_geo_city"`
	SourceGeoCountry      string `json:"source_geo_country"`
	SourceID              string `json:"source_id"`
	SourceName            string `json:"source_name"`
	SourceType            string `json:"source_type"`
	SourcePort            string `json:"source_port"`
	TxBytes               int    `json:"tx_bytes"`
	TxPackets             int    `json:"tx_packets"`
	UserID                string `json:"user_id"`
}

type AuditMeta struct {
	CreatedAt            time.Time `json:"created_at"`
	Fqdn                 string    `json:"fqdn"`
	LocationCityName     string    `json:"location_city_name"`
	LocationCountryCode  string    `json:"location_country_code"`
	LocationConnectionIp string    `json:"location_connection_ip"`
	LocationGeoNameId    int       `json:"location_geo_name_id"`
	Ip                   string    `json:"ip"`
}

type TrafficEvent struct {
	ID          string      `json:"ID"`
	InitiatorID string      `json:"InitiatorID"`
	Message     string      `json:"Message"`
	Meta        TrafficMeta `json:"Meta"`
	Reference   string      `json:"Reference"`
	TargetID    string      `json:"target_id"`
	Timestamp   time.Time   `json:"Timestamp"`
}

type AuditEvent struct {
	ID          int       `json:"ID"`
	InitiatorID string    `json:"InitiatorID"`
	Message     string    `json:"Message"`
	Meta        AuditMeta `json:"meta"`
	Reference   string    `json:"reference"`
	TargetID    string    `json:"target_id"`
	Timestamp   time.Time `json:"Timestamp"`
}

type SplunkTrafficEvent struct {
	Time       float64 `json:"time"`
	Protocol   string  `json:"protocol"`
	SrcIP      string  `json:"srcip"`
	SrcPort    int     `json:"srcport"`
	SourceName string  `json:"sourcename"`
	Email      string  `json:"email"`
	DstIP      string  `json:"dstip"`
	DstPort    int     `json:"dstport"`
	ExitNode   string  `json:"exitnode"`
	Message    string  `json:"message"`
}

type SplunkAuditEvent struct {
	Time                 float64 `json:"time"`
	User                 string  `json:"user"`
	Message              string  `json:"message"`
	LocationCityName     string  `json:"location_city_name"`
	LocationCountryCode  string  `json:"location_country_code"`
	LocationConnectionIp string  `json:"location_connection_ip"`
	LocationGeoNameId    int     `json:"location_geo_name_id"`
	Ip                   string  `json:"ip"`
	Fqdn                 string  `json:"fqdn"`
}
