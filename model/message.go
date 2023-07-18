package model

// Peer structure
type HostConfig struct {
	NetName string `json:"netName"  bson:"netName"`
	NetId   string `json:"netid"    bson:"netid"`
	Hosts   []Host `json:"hosts"    bson:"hosts"`
}

// Host structure
type Message struct {
	Id     string       `json:"id"       bson:"id"`
	Config []HostConfig `json:"config"   bson:"config"`
}

type ServiceMessage struct {
	Id     string    `json:"id"       bson:"id"`
	Config []Service `json:"config"   bson:"config"`
}
