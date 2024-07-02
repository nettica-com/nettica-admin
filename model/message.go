package model

type VPNConfig struct {
	NetName string `json:"netName"  bson:"netName"`
	NetId   string `json:"netid"    bson:"netid"`
	VPNs    []VPN  `json:"vpns"     bson:"vpns"`
}

type Message struct {
	Version string      `json:"version,omitempty"   bson:"version,omitempty"`
	Id      string      `json:"id"                  bson:"id"`
	Device  *Device     `json:"device"              bson:"device"`
	Config  []VPNConfig `json:"config"              bson:"config"`
}

type ServiceMessage struct {
	Version string    `json:"version,omitempty"     bson:"version,omitempty"`
	Id      string    `json:"id"                    bson:"id"`
	Config  []Service `json:"config"                bson:"config"`
}

type StatusResponse struct {
	Status  int    `json:"status" bson:"status"`
	Message string `json:"message" bson:"message"`
}

// AgentNotification structure
// type: dns, info, error
// text: message
type AgentNotification struct {
	Type string `json:"type" bson:"type"`
	Text string `json:"text" bson:"text"`
}
