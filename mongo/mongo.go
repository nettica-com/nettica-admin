package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create a cache of mongo db connections
var mongoClient *mongo.Client
var m sync.Mutex

// getMongoClient returns a mongo client for the given connection string
func getMongoClient() (*mongo.Client, error) {

	m.Lock()
	defer m.Unlock()

	if mongoClient != nil {

		return mongoClient, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	o := options.Client().SetMinPoolSize(2)
	o = o.ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING"))
	client, err := mongo.Connect(ctx, o)

	if err != nil {
		return nil, err
	}

	mongoClient = client

	return mongoClient, nil
}

///
/// Mongo DB primitives
///

// Serialize write interface to disk
func Serialize(id string, parm string, col string, c interface{}) error {
	//b, err := json.MarshalIndent(c, "", "  ")
	//if err != nil {
	//	return err
	//}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	//	json := fmt.Sprintf("%v", user)
	var b interface{}
	err = bson.UnmarshalExtJSON([]byte(data), true, &b)
	if err != nil {
		return err
	}

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"%s\":\"%s\"}", parm, id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": b,
	}

	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)

	//	if res != nil && res.Err != nil {
	//		collection.InsertOne(ctx, b)
	//	}

	return err
}

// Deserialize read interface from disk
func Deserialize(id string, parm string, col string, t reflect.Type) (interface{}, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"%s\":\"%s\"}", parm, id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
	if err != nil {
		return nil, err
	}

	switch t.String() {

	case "model.Device":
		var c *model.Device
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Account":
		var c *model.Account
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Limits":
		var c *model.Limits
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.VPN":
		var c *model.VPN
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.User":
		var c *model.User
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Network":
		var c *model.Network
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Subscription":
		var c *model.Subscription
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Service":
		var c *model.Service
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Server":
		var c *model.Server
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err
	}
	log.Infof("reflect.TypeOf(t) = %v", t.String())

	return nil, nil
}

// DeleteVPN removes the given id from the given collection
func DeleteVPN(id string, col string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return err
	}

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"id\":\"%s\"}", id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
	if err != nil {
		return err
	}

	collection.FindOneAndDelete(ctx, filter)

	return nil
}

// Delete removes the given id from the given collection
func Delete(id string, ident string, col string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return err
	}

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"%s\":\"%s\"}", ident, id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
	if err != nil {
		return err
	}

	collection.FindOneAndDelete(ctx, filter)

	return nil
}

// ReadAllDevices from MongoDB
func ReadAllDevices(param string, id string) ([]*model.Device, error) {
	devices := make([]*model.Device, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("devices")

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", param, id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var device *model.Device
			err = cursor.Decode(&device)
			if err == nil {
				devices = append(devices, device)
			}
		}

	}

	return devices, err

}

// ReadDevicesAndVPNsForAccount
func ReadDevicesAndVPNsForAccount(accountid string) ([]*model.Device, error) {

	devices := make([]*model.Device, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	// Open an aggregation cursor
	coll := client.Database("nettica").Collection("devices")
	cursor, err := coll.Aggregate(ctx, bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "accountid", Value: accountid}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "vpns"},
					{Key: "localField", Value: "id"},
					{Key: "foreignField", Value: "deviceid"},
					{Key: "as", Value: "vpns"},
				},
			},
		},
	})

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var device *model.Device
			err = cursor.Decode(&device)
			if err == nil {
				sort.Slice(device.VPNs, func(i, j int) bool {
					return device.VPNs[i].Name < device.VPNs[j].Name
				})
				devices = append(devices, device)
			}
		}

	}

	return devices, err

}

// ReadVPNsforNetwork from MongoDB
func ReadVPNsforNetwork(netid string) ([]*model.VPN, error) {

	vpns := make([]*model.VPN, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return vpns, err
	}

	// Open an aggregation cursor
	coll := client.Database("nettica").Collection("vpns")
	cursor, err := coll.Aggregate(ctx, bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "netid", Value: netid}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "devices"},
					{Key: "localField", Value: "deviceid"},
					{Key: "foreignField", Value: "id"},
					{Key: "as", Value: "devices"},
				},
			},
		},
	})

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var vpn *model.VPN
			err = cursor.Decode(&vpn)
			if err == nil {
				vpns = append(vpns, vpn)
			}
		}

	}

	// Alphabetize the results
	sort.Slice(vpns, func(i, j int) bool {
		return vpns[i].Name < vpns[j].Name
	})

	return vpns, err

}

// ReadAllHosts from MongoDB
func ReadAllVPNs(param string, id string) ([]*model.VPN, error) {
	vpns := make([]*model.VPN, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("vpns")

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", param, id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var vpn *model.VPN
			err = cursor.Decode(&vpn)
			if err == nil {
				vpns = append(vpns, vpn)
			}
		}

	}

	// Alphabetize the results
	sort.Slice(vpns, func(i, j int) bool {
		return vpns[i].Name < vpns[j].Name
	})

	return vpns, err

}

// ReadAllNetworks from MongoDB
func ReadAllNetworks(param string, id string) ([]*model.Network, error) {
	nets := make([]*model.Network, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", param, id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	collection := client.Database("nettica").Collection("networks")
	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var net *model.Network
			err = cursor.Decode(&net)
			if err == nil {
				nets = append(nets, net)
			}
		}

	}

	return nets, nil

}

// ReadAllServices from MongoDB
func ReadServices(param string, id string) ([]*model.Service, error) {
	services := make([]*model.Service, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)

		return nil, err
	}

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", param, id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	collection := client.Database("nettica").Collection("services")
	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var service *model.Service
			err = cursor.Decode(&service)
			if err == nil {
				services = append(services, service)
			}
		}

	}

	return services, nil

}

// ReadAllUsers from MongoDB
func ReadAllUsers() []*model.User {
	users := make([]*model.User, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil
	}
	collection := client.Database("nettica").Collection("users")

	cursor, err := collection.Find(ctx, bson.D{})

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var user *model.User
			err = cursor.Decode(&user)
			if err == nil {
				users = append(users, user)
			}
		}

	}

	return users

}

// ReadAllAccounts from MongoDB
func ReadAllAccounts(email string) ([]*model.Account, error) {
	accounts := make([]*model.Account, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("accounts")

	filter := bson.D{}
	if email != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", "email", email)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var account *model.Account
			err = cursor.Decode(&account)
			if err == nil {
				accounts = append(accounts, account)
			}
		}

	}

	return accounts, err

}

// ReadAllAccountsForID from MongoDB
func ReadAllAccountsForID(id string) ([]*model.Account, error) {
	accounts := make([]*model.Account, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("accounts")

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", "parent", id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var account *model.Account
			err = cursor.Decode(&account)
			if err == nil {
				accounts = append(accounts, account)
			}
		}

	}

	return accounts, err

}

// ReadAccountForUser from MongoDB
func ReadAccountForUser(email string, accountid string) (*model.Account, error) {
	var account *model.Account

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("accounts")

	filter := bson.D{}
	if email != "" {
		findstr := fmt.Sprintf("{\"email\":\"%s\", \"parent\":\"%s\"}", email, accountid)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			account = &model.Account{}
			err = cursor.Decode(&account)
			if err == nil {
				return account, nil
			}
		}

	}

	return nil, err

}

// ReadAllSubscriptions from MongoDB
func ReadAllSubscriptions(email string) ([]*model.Subscription, error) {
	subscriptions := make([]*model.Subscription, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("subscriptions")

	filter := bson.D{}
	if email != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", "email", email)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var subscription *model.Subscription
			err = cursor.Decode(&subscription)
			if err == nil {
				subscriptions = append(subscriptions, subscription)
			}
		}

	}

	return subscriptions, err

}

// ReadAllServices from MongoDB
func ReadAllServices(email string) ([]*model.Service, error) {
	services := make([]*model.Service, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("services")

	filter := bson.D{}
	if email != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", "email", email)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var service *model.Service
			err = cursor.Decode(&service)
			if err == nil {
				services = append(services, service)
			}
		}

	}

	return services, err

}

// ReadAllServers from MongoDB
func ReadAllServers() ([]*model.Server, error) {
	servers := make([]*model.Server, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("servers")
	cursor, err := collection.Find(ctx, bson.D{})
	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var server *model.Server
			err = cursor.Decode(&server)
			if err == nil {
				servers = append(servers, server)
			}
		}
	}

	// Alphabetize the results
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Description < servers[j].Description
	})
	return servers, err
}

// ReadServiceHost from MongoDB
func ReadServiceHost(id string) ([]*model.Service, error) {
	services := make([]*model.Service, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return nil, err
	}

	collection := client.Database("nettica").Collection("services")

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", "serviceGroup", id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
		if err != nil {
			return nil, err
		}

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var service *model.Service
			err = cursor.Decode(&service)
			if err == nil {
				services = append(services, service)
			}
		}

	}

	return services, err

}

// UpsertUser to MongoDB
func UpsertUser(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return err
	}

	collection := client.Database("nettica").Collection("users")

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	var b interface{}
	err = bson.UnmarshalExtJSON([]byte(data), true, &b)
	if err != nil {
		return err
	}

	findstr := fmt.Sprintf("{\"email\":\"%s\"}", user.Email)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": b,
	}

	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Error(err)
	}

	return nil
}

// Initialize the mongo db and create the indexes
func Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := getMongoClient()
	if err != nil {
		log.Errorf("getMongoClient: %v", err)
		return err
	}

	// users

	collection := client.Database("nettica").Collection("users")

	// create an index for the email field
	index := mongo.IndexModel{Keys: bson.M{"email": 1}, Options: nil}
	_, err = collection.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Error(err)
	}

	// accounts

	_, err = client.Database("nettica").Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"email": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"parent": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"apiKey": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	// devices

	_, err = client.Database("nettica").Collection("devices").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("devices").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"accountid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("devices").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"createdBy": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("devices").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"apiKey": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("devices").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"instanceid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	// networks

	_, err = client.Database("nettica").Collection("networks").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("networks").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"accountid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	// vpns

	_, err = client.Database("nettica").Collection("vpns").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("vpns").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"accountid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("vpns").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"deviceid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("vpns").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"netid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("vpns").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"createdBy": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	// subscriptions

	_, err = client.Database("nettica").Collection("subscriptions").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("subscriptions").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"accountid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("subscriptions").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"email": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	// services

	_, err = client.Database("nettica").Collection("services").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("services").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"accountid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("services").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"email": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("services").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"serviceGroup": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	// servers

	_, err = client.Database("nettica").Collection("servers").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("servers").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"serviceGroup": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	// limits

	_, err = client.Database("nettica").Collection("limits").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"id": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}
	_, err = client.Database("nettica").Collection("limits").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"accountid": 1}, Options: nil})
	if err != nil {
		log.Error(err)
	}

	return nil
}
