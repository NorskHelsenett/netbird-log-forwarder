package apicontracts

import (
	"encoding/json"
	"strings"
	"time"
)

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

// type AuditMeta struct {
// 	CreatedAt            time.Time `json:"created_at"`
// 	Fqdn                 string    `json:"fqdn"`
// 	LocationCityName     string    `json:"location_city_name"`
// 	LocationCountryCode  string    `json:"location_country_code"`
// 	LocationConnectionIp string    `json:"location_connection_ip"`
// 	LocationGeoNameId    int       `json:"location_geo_name_id"`
// 	Ip                   string    `json:"ip"`
// }

type TrafficEvent struct {
	ID          string      `json:"ID"`
	InitiatorID string      `json:"InitiatorID"`
	Message     string      `json:"Message"`
	Meta        TrafficMeta `json:"Meta"`
	Reference   string      `json:"Reference"`
	TargetID    string      `json:"target_id"`
	Timestamp   time.Time   `json:"Timestamp"`
}

// type AuditEvent struct {
// 	ID          int       `json:"ID"`
// 	InitiatorID string    `json:"InitiatorID"`
// 	Message     string    `json:"Message"`
// 	Meta        AuditMeta `json:"meta"`
// 	Reference   string    `json:"reference"`
// 	TargetID    string    `json:"target_id"`
// 	Timestamp   time.Time `json:"Timestamp"`
// }

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

// type SplunkAuditEvent struct {
// 	Time                 float64 `json:"time"`
// 	User                 string  `json:"user"`
// 	Message              string  `json:"message"`
// 	LocationCityName     string  `json:"location_city_name"`
// 	LocationCountryCode  string  `json:"location_country_code"`
// 	LocationConnectionIp string  `json:"location_connection_ip"`
// 	LocationGeoNameId    int     `json:"location_geo_name_id"`
// 	Ip                   string  `json:"ip"`
// 	Fqdn                 string  `json:"fqdn"`
// }

type AuditEventEnvelope struct {
	ID          int             `json:"ID"`
	Timestamp   time.Time       `json:"-"` // normalisert til time.Time
	Message     string          `json:"message"`
	InitiatorID string          `json:"initiator_id"`
	TargetID    string          `json:"target_id"`
	Extra       map[string]any  `json:"-"` // alle ukjente felter
	Raw         json.RawMessage `json:"-"` // original JSON (for re-logging)
}

func (e *AuditEventEnvelope) UnmarshalJSON(b []byte) error {
	e.Raw = append(e.Raw[:0], b...) // behold original
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	// Timestamp (RFC3339Nano først, så RFC3339)
	if v, ok := m["Timestamp"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			if ts, err := time.Parse(time.RFC3339Nano, s); err == nil {
				e.Timestamp = ts.UTC()
			} else if ts, err := time.Parse(time.RFC3339, s); err == nil {
				e.Timestamp = ts.UTC()
			} else {
				e.Timestamp = time.Now().UTC() // fallback
			}
		}
		delete(m, "Timestamp")
	}

	if v, ok := m["Message"]; ok {
		_ = json.Unmarshal(v, &e.Message)
		delete(m, "Message")
	}
	if v, ok := m["InitiatorID"]; ok {
		_ = json.Unmarshal(v, &e.InitiatorID)
		delete(m, "InitiatorID")
	}
	if v, ok := m["target_id"]; ok {
		_ = json.Unmarshal(v, &e.TargetID)
		delete(m, "target_id")
	}

	e.Extra = make(map[string]any)

	// Reserverte feltnavn som ikke må overskrives
	reserved := map[string]struct{}{
		"timestamp":   {},
		"message":     {},
		"initiatorid": {},
		// "targetid":    {},
	}

	// 1) Flate ut Meta -> Extra (hvis Meta er objekt)
	if mv, ok := m["meta"]; ok {
		var metaObj map[string]json.RawMessage
		if err := json.Unmarshal(mv, &metaObj); err == nil {
			for k, raw := range metaObj {
				var anyv any
				if err := json.Unmarshal(raw, &anyv); err != nil {
					anyv = string(raw)
				}
				kl := strings.ToLower(k)
				if _, isReserved := reserved[kl]; isReserved {
					e.Extra["meta_"+k] = anyv // ikke overskriv reserverte
				} else if _, exists := e.Extra[k]; exists {
					e.Extra["meta_"+k] = anyv // kollisjon: prefiksér
				} else {
					e.Extra[k] = anyv
				}
			}
		}
		// else {
		// 	// Meta var ikke et objekt → lagre som "Meta" i Extra
		// 	var anyv any
		// 	if err := json.Unmarshal(mv, &anyv); err != nil {
		// 		anyv = string(mv)
		// 	}
		// 	e.Extra["Meta"] = anyv
		// }
		delete(m, "Meta")
	}

	// 2) Resten av ukjente toppnivå-felter -> Extra
	for k, raw := range m {
		var anyv any
		if err := json.Unmarshal(raw, &anyv); err != nil {
			anyv = string(raw)
		}
		kl := strings.ToLower(k)
		if _, isReserved := reserved[kl]; isReserved {
			e.Extra["extra_"+k] = anyv // sikkerhet mot overskriving
			continue
		}
		if _, exists := e.Extra[k]; exists {
			e.Extra["extra_"+k] = anyv // kollisjon: prefiksér
		} else {
			e.Extra[k] = anyv
		}
	}

	return nil
}
