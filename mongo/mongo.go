package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	model "github.com/nettica-com/nettica/model"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	data, err := json.Marshal(c)
	//	json := fmt.Sprintf("%v", user)
	var b interface{}
	err = bson.UnmarshalExtJSON([]byte(data), true, &b)

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"%s\":\"%s\"}", parm, id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"%s\":\"%s\"}", parm, id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

	switch t.String() {
	case "model.Account":
		var c *model.Account
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Host":
		var c *model.Host
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.User":
		var c *model.User
		err = collection.FindOne(ctx, filter).Decode(&c)
		return c, err

	case "model.Mesh":
		var c *model.Mesh
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

// DeleteHost removes the given id from the given collection
func DeleteHost(id string, col string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"id\":\"%s\"}", id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

	collection.FindOneAndDelete(ctx, filter)

	return nil
}

// Delete removes the given id from the given collection
func Delete(id string, ident string, col string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	collection := client.Database("nettica").Collection(col)

	findstr := fmt.Sprintf("{\"%s\":\"%s\"}", ident, id)
	var filter interface{}
	err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

	collection.FindOneAndDelete(ctx, filter)

	return nil
}

// ReadAllHosts from MongoDB
func ReadAllHosts(param string, id string) ([]*model.Host, error) {
	hosts := make([]*model.Host, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	collection := client.Database("nettica").Collection("hosts")

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", param, id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

	}

	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var host *model.Host
			err = cursor.Decode(&host)
			if err == nil {
				hosts = append(hosts, host)
			}
		}

	}

	return hosts, err

}

// ReadAllMeshes from MongoDB
func ReadAllMeshes(param string, id string) []*model.Mesh {
	meshes := make([]*model.Mesh, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", param, id)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

	}

	collection := client.Database("nettica").Collection("mesh")
	cursor, err := collection.Find(ctx, filter)

	if err == nil {

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var mesh *model.Mesh
			err = cursor.Decode(&mesh)
			if err == nil {
				meshes = append(meshes, mesh)
			}
		}

	}

	return meshes

}

// ReadAllUsers from MongoDB
func ReadAllUsers() []*model.User {
	users := make([]*model.User, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	collection := client.Database("nettica").Collection("accounts")

	filter := bson.D{}
	if email != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", "email", email)
		err = bson.UnmarshalExtJSON([]byte(findstr), true, &filter)

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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_CONNECTION_STRING")))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()

	collection := client.Database("nettica").Collection("accounts")

	filter := bson.D{}
	if id != "" {
		findstr := fmt.Sprintf("{\"%s\":\"%s\"}", "parent", id)
		err = bson.UnmarshalExtJSON([]b