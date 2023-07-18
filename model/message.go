package model

// Peer structure
type HostConfig struct {
	MeshName string `json:"meshName"  bson:"meshName"`
	MeshId   string `json:"meshid"    bson:"meshid"`
	Hosts    []Host `json:"hosts"      bson:"hosts"`
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
