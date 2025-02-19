package model

type DiscountRequest struct {
	UserName  string `json:"userName" bson:"userName"`
	ProductId string `json:"productId" bson:"productId"`
	OfferId   string `json:"identifier" bson:"identifier"`
}

type DiscountResponse struct {
	UserName  string `json:"userName" bson:"userName"`
	ProductId string `json:"productId" bson:"productId"`
	OfferId   string `json:"identifier" bson:"identifier"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
	KeyId     string `json:"keyIdentifier" bson:"keyIdentifier"`
	Nonce     string `json:"nonce" bson:"nonce"`
	Signature string `json:"signature" bson:"signature"`
}

type DiscountOffers struct {
	Offers []string `json:"offers" bson:"offers"`
}
