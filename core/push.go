package core

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"

	"context"
	"fmt"
	"reflect"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var (
	PM PushManager = PushManager{
		Pusher:      &model.Pusher{},
		pusherCache: make(map[string]*model.Pusher),
	}
	Push PushCore = PushCore{
		PushDevices: make(map[string]string),
		PushTokens:  make(map[string]string),
		Enabled:     false,
	}
)

type PushManager struct {
	*model.Pusher
	pusherCache map[string]*model.Pusher
}

func (pm *PushManager) Register() error {
	server := os.Getenv("SERVER")

	if server == "" {
		return errors.New("server is empty")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return errors.New("failed to get hostname: " + err.Error())
	}

	pusher, err := mongo.GetPushSettings(server, hostname)
	if err != nil {

		// register with nettica
		pm.Pusher.Server = server
		pm.Pusher.Host = hostname
		t := true
		pm.Pusher.Enabled = &t
		pm.Pusher.Version = "1.0"

		jsonData, err := json.Marshal(pm.Pusher)
		if err != nil {
			return errors.New("failed to marshal pusher to JSON: " + err.Error())
		}
		rsp, err := http.Post("https://my.nettica.com/api/v1.0/push", "application/json", strings.NewReader(string(jsonData)))
		if err != nil {
			log.Errorf("failed to register pusher: %v", err)
			return errors.New("failed to register pusher: " + err.Error())
		}
		defer rsp.Body.Close()
		if rsp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(rsp.Body)
			log.Errorf("failed to register pusher: %s", string(body))
			return errors.New("failed to register pusher: " + string(body))
		}

		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Errorf("failed to read response body: %v", err)
			return errors.New("failed to read response body: " + err.Error())
		}
		err = json.Unmarshal(body, &pm.Pusher)
		if err != nil {
			log.Errorf("failed to unmarshal pusher: %v", err)
			return errors.New("failed to unmarshal pusher: " + err.Error())
		}
		err = mongo.Serialize(pm.Pusher.Id, "id", "push", pm.Pusher)
		if err != nil {
			log.Errorf("failed to serialize pusher: %v", err)
			return errors.New("failed to serialize pusher: " + err.Error())
		}
	} else {

		// update the settings from nettica
		req, err := http.NewRequest("GET", "https://my.nettica.com/api/v1.0/push/"+pusher.Id, nil)
		if err != nil {
			return errors.New("pusher: register: failed to get request: " + err.Error())
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", pusher.ApiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Nettica Push")
		client := &http.Client{}
		rsp, err := client.Do(req)
		if err != nil {
			return errors.New("pusher: register: failed to get request: " + err.Error())
		}
		defer rsp.Body.Close()
		if rsp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(rsp.Body)
			log.Errorf("failed to get pusher: %s", string(body))
			return errors.New("failed to get pusher: " + string(body))
		}
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Errorf("failed to read response body: %v", err)
			return errors.New("failed to read response body: " + err.Error())
		}
		err = json.Unmarshal(body, &pm.Pusher)
		if err != nil {
			log.Errorf("failed to unmarshal pusher: %v", err)
			return errors.New("failed to unmarshal pusher: " + err.Error())
		}
		err = mongo.Serialize(pm.Pusher.Id, "id", "push", pm.Pusher)
		if err != nil {
			log.Errorf("failed to serialize pusher: %v", err)
			return errors.New("failed to serialize pusher: " + err.Error())
		}
	}

	return nil
}

func (pm *PushManager) Send(msg *model.Push) error {
	if msg == nil {
		return errors.New("push is nil")
	}

	msg.Version = pm.Version
	msg.ApiKey = pm.ApiKey
	msg.Id = pm.Id

	if err := msg.IsValid(); err != nil {
		return fmt.Errorf("error validating push notification: %v", err)
	}

	if pm.Enabled != nil && *pm.Enabled {
		msg.Version = pm.Version
		msg.ApiKey = pm.ApiKey
		msg.Id = pm.Id

		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error marshaling push notification: %v", err)
		}

		req, err := http.NewRequest("POST", "https://my.nettica.com/api/v1.0/push/"+msg.Id, strings.NewReader(string(data)))
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", pm.ApiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Nettica Push")

		client := &http.Client{}
		rsp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %v", err)
		}
		defer rsp.Body.Close()
		if rsp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(rsp.Body)
			log.Errorf("failed to send push notification: %s", string(body))
			return errors.New("failed to send push notification: " + string(body))
		}
		_, err = io.ReadAll(rsp.Body)
		if err != nil {
			log.Errorf("failed to read response body: %v", err)
			return errors.New("failed to read response body: " + err.Error())
		}
		log.Infof("Push: title: %s, message: %s", msg.Title, msg.Message)

	}
	return nil
}

func (pm *PushManager) Add(pusher *model.Pusher) error {

	if pusher.Id == "" {
		log.Errorf("pusher id is empty %v", pusher)
		return fmt.Errorf("pusher id is empty")
	}

	err := mongo.Serialize(pusher.Id, "id", "push", pusher)
	if err != nil {
		return errors.New("failed to serialize pusher: " + err.Error())
	}

	pm.pusherCache[pusher.Id] = pusher
	return nil
}

func (pm *PushManager) Get(pusherId string) (*model.Pusher, error) {
	if pusherId == "" {
		log.Errorf("pusher id is empty %v", pusherId)
		return nil, fmt.Errorf("pusher id is empty")
	}

	pusher, ok := pm.pusherCache[pusherId]
	if !ok {
		return nil, errors.New("pusher not found")
	}
	return pusher, nil
}

func (pm *PushManager) Remove(pusherId string) error {
	if pusherId == "" {
		log.Errorf("pusher id is empty %v", pusherId)
		return fmt.Errorf("pusher id is empty")
	}

	err := mongo.Delete(pusherId, "id", "push")
	if err != nil {
		return errors.New("failed to delete pusher: " + err.Error())
	}

	delete(pm.pusherCache, pusherId)
	return nil
}

func (pm *PushManager) Load() error {

	if pm.Pusher == nil {
		pm.Pusher = &model.Pusher{}
	}
	pusherCache, err := mongo.GetPushers()
	if err != nil {
		return errors.New("failed to get pushers: " + err.Error())
	}

	pm.pusherCache = make(map[string]*model.Pusher, len(pusherCache))
	for _, pusher := range pusherCache {
		pm.pusherCache[pusher.Id] = pusher
	}

	// If we are the server, disable the pusher
	// because we are going to send the push notifications
	f := false
	pm.Pusher.Enabled = &f

	return nil
}

type PushCore struct {
	app         *firebase.App
	client      *messaging.Client
	PushDevices map[string]string
	PushTokens  map[string]string
	Enabled     bool
}

// Initialize initializes the push notification service
func (p *PushCore) Initialize() error {

	var err error
	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_SERVER_KEY_PATH"))
	p.app, err = firebase.NewApp(context.Background(), nil, opt)
	if err == nil {
		p.client, err = p.app.Messaging(context.Background())
		if err != nil {
			log.Error(err)
			// we are on the client so register with the server
			err = PM.Register()
			if err != nil {
				return fmt.Errorf("error initializing push client: %v", err)
			}
			log.Info("Push Client Registered")
		} else {

			// we are on the server so load the clients
			err = PM.Load()
			if err != nil {
				return fmt.Errorf("error loading push settings: %v", err)
			}
		}
	}
	devices, err := mongo.GetDevicesForPushNotifications()
	if err != nil {
		return fmt.Errorf("error getting devices for push notifications: %v", err)
	}

	for _, device := range devices {
		if device.Push != nil {
			p.PushDevices[device.Id] = *device.Push
			p.PushTokens[*device.Push] = device.Id
		}
	}

	p.Enabled = true

	return nil

}

func (p *PushCore) AddDevice(deviceId, pushToken string) {
	if p.Enabled {
		p.PushDevices[deviceId] = pushToken
		p.PushTokens[pushToken] = deviceId
	}
}

func (p *PushCore) RemoveDevice(deviceId string) {
	if p.Enabled {
		pushToken, ok := p.PushDevices[deviceId]
		if ok {
			delete(p.PushDevices, deviceId)
			delete(p.PushTokens, pushToken)
		}
	}
}

func (p *PushCore) RemovePushToken(pushToken string) {
	if p.Enabled {
		deviceId, ok := p.PushTokens[pushToken]
		if ok {
			delete(p.PushTokens, pushToken)
			delete(p.PushDevices, deviceId)
		}
	}
}

// SendPushNotification sends a push notification to a device
func (p *PushCore) SendPushNotification(pushToken, title, body string) error {

	log.Infof("Push: %s - %s", title, body)

	if PM.Enabled != nil && *PM.Enabled {
		msg := &model.Push{
			Title:   title,
			Message: body,
			Token:   pushToken,
			Version: "1.0",
			Id:      PM.Id,
			ApiKey:  PM.ApiKey,
		}
		err := msg.IsValid()
		if err != nil {
			return fmt.Errorf("error validating push notification: %v", err)
		}
		err = PM.Send(msg)
		if err != nil {
			p.RemovePushTokenFromDevice(pushToken)
			return fmt.Errorf("error sending push notification: %v", err)
		}
		return nil
	}

	if p.Enabled {
		notification := messaging.Message{
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Token: pushToken,
		}
		_, err := p.client.Send(context.Background(), &notification)
		if err != nil {
			// if not found, remove the push token from the device
			if strings.Contains(err.Error(), "404") {
				p.RemovePushTokenFromDevice(pushToken)
			}
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}

func (p *PushCore) RemovePushTokenFromDevice(pushToken string) {
	deviceId, ok := p.PushTokens[pushToken]
	if ok {
		delete(p.PushTokens, pushToken)
		delete(p.PushDevices, deviceId)
		// remove the push token from the device
		d, err := mongo.Deserialize(deviceId, "id", "devices", reflect.TypeOf(model.Device{}))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read device")
		} else {
			device := d.(*model.Device)
			*device.Push = ""
			err = mongo.Serialize(device.Id, "id", "devices", device)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("failed to serialize device")
			}
		}
		log.Infof("Push token %s removed for device %s", pushToken, deviceId)
	}
}
