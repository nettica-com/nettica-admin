package push

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/mongo"

	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var (
	app         *firebase.App
	client      *messaging.Client
	PushDevices = make(map[string]string)
	PushTokens  = make(map[string]string)
	enabled     = false
)

// Initialize initializes the push notification service
func Initialize() error {

	var err error
	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_SERVER_KEY_PATH"))
	app, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}

	client, err = app.Messaging(context.Background())
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %v", err)
	}

	devices, err := mongo.GetDevicesForPushNotifications()
	if err != nil {
		return fmt.Errorf("error getting devices for push notifications: %v", err)
	}

	for _, device := range devices {
		PushDevices[device.Id] = device.Push
		PushTokens[device.Push] = device.Id
	}

	enabled = true

	return nil

}

// SendPushNotification sends a push notification to a device
func SendPushNotification(pushToken, title, body string) error {

	if enabled {
		notification := messaging.Message{
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Token: pushToken,
		}
		_, err := client.Send(context.Background(), &notification)
		if err != nil {
			// if not found, remove the push token from the device
			if strings.Contains(err.Error(), "404") {
				deviceId, ok := PushTokens[pushToken]
				if ok {
					delete(PushTokens, pushToken)
					delete(PushDevices, deviceId)
					// remove the push token from the device
					d, err := mongo.Deserialize(deviceId, "id", "devices", reflect.TypeOf(model.Device{}))
					if err != nil {
						log.WithFields(log.Fields{
							"err": err,
						}).Error("failed to read device")
					} else {
						device := d.(*model.Device)
						device.Push = ""
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
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}
