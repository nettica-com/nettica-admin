package model

// Peer structure
type VPNConfig struct {
	NetName string `json:"netName"  bson:"netName"`
	NetId   string `json:"netid"    bson:"netid"`
	VPNs    []VPN  `json:"vpns"     bson:"vpns"`
}

// Host structure
type Message struct {
	Id     string      `json:"id"       bson:"id"`
	Device *Device     `json:"device"   bson:"device"`
	Config []VPNConfig `json:"config"   bson:"config"`
}

type ServiceMessage struct {
	Id     string    `json:"id"       bson:"id"`
	Config []Service `json:"config"   bson:"config"`
}
