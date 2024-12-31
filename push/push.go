package push

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/nettica-com/nettica-admin/mongo"

	"os"

	"google.golang.org/api/option"
)

var (
	app         *firebase.App
	client      *messaging.Client
	PushDevices = make(map[string]string)
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
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}
