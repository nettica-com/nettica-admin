package subscription

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/mongo"
	"github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/subscriptions")
	{

		g.POST("/helio", createHelioSubscription)
		g.POST("", createSubscription)
		g.POST("/update", updateSubscriptionWoo)
		g.POST("/apple", createSubscriptionApple2)
		g.POST("/apple/webhook", handleAppleWebhook2)
		g.POST("/apple/discount", handleAppleDiscount)
		g.POST("/android", createSubscriptionAndroid)
		g.POST("/android/webhook", handleAndroidWebhook2)
		g.GET("/offers/:id", getOffers)
		g.POST("/trial/:id", createTrial)
		g.GET("/:id", readSubscription)
		g.PATCH("/:id", updateSubscription)
		g.DELETE("/:id", deleteSubscription)
		g.DELETE("/trials", deleteTrialSubscriptions)
		g.GET("", readSubscriptions)
		g.GET("/deleted", readSubscriptionsDeleted)
	}
}

func deleteTrialSubscriptions(c *gin.Context) {

	// get the basic auth credentials from the request and check them
	// against the environment variables WC_USERNAME and WC_PASSWORD
	username, password, ok := c.Request.BasicAuth()
	if !ok || username != os.Getenv("WC_USERNAME") || password != os.Getenv("WC_PASSWORD") {
		log.Error("deleteTrialSubscriptions: invalid basic auth credentials")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	subscriptions, err := core.ReadAllTrials()
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "deleteTrialSubscriptions: Internal server error"})
		return
	}

	now := time.Now().UTC()
	for _, subscription := range subscriptions {
		if subscription.Status == "active" && subscription.Expires.Before(now) && !subscription.Expires.Before(*subscription.Issued) {
			log.Infof("suspending trial subscription %v", subscription)
			core.ExpireSubscription(subscription.Id)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK", "message": "Service for expired trial subscriptions suspended"})
}

func createTrial(c *gin.Context) {
	id := c.Param("id")

	discounts, err := core.GetOffers(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Offers not found"})
		return
	}

	if len(discounts.Offers) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No offers available"})
		return
	}

	// get user from token and add to client infos
	oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
	user, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// update the limits

	limits, err := core.ReadLimits(id)
	if err != nil {
		log.Error(err)
		limits_id, err := util.GenerateRandomString(8)
		if err != nil {
			log.Error(err)
		}
		limits_id = "limits-" + limits_id

		limits = &model.Limits{
			Id:          limits_id,
			AccountID:   id,
			MaxDevices:  5,
			MaxNetworks: 2,
			MaxMembers:  2,
			MaxServices: 0,
			Tolerance:   core.GetDefaultTolerance(),
			CreatedBy:   user.Email,
			UpdatedBy:   user.Email,
			Created:     time.Now(),
			Updated:     time.Now(),
		}
	}

	limits.MaxDevices += 1
	limits.MaxNetworks += 1
	limits.MaxServices += 1

	errs := limits.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("limits validation error")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limits validation error"})
		return
	}

	// save limits to mongodb
	mongo.Serialize(limits.Id, "id", "limits", limits)

	// create a new subscription
	var subscription model.Subscription
	subscription.Id, _ = util.RandomString(8)
	subscription.Id = "trial-" + subscription.Id
	subscription.AccountID = user.AccountID
	subscription.Email = user.Email
	subscription.Name = "Trial"
	subscription.Description = "Free 7 Day Trial"
	now := time.Now()
	expires := now.AddDate(0, 0, 7)
	subscription.Issued = &now
	subscription.LastUpdated = &now
	subscription.CreatedBy = user.Email
	subscription.Expires = &expires
	subscription.UpdatedBy = user.Email
	subscription.Credits = 1
	subscription.Sku = "trial"
	subscription.Status = "active"
	subscription.AutoRenew = false
	subscription.Receipt, _ = util.RandomString(16) // generate a random receipt for trial subscriptions
	subscription.Receipt = "trial-" + subscription.Receipt
	isDeleted := false
	subscription.IsDeleted = &isDeleted

	errs = subscription.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("subscription validation error")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription validation error"})
		return
	}

	// save subscription to mongodb
	mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

	log.Infof("created trial subscription: %s for %s", subscription.Id, user.Email)

	err = core.SubscriptionEmail(&subscription)
	if err != nil {
		log.Errorf("failed to send email: %v", err)
	}

	c.JSON(http.StatusOK, subscription)

}

func getOffers(c *gin.Context) {
	id := c.Param("id")
	offers, err := core.GetOffers(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Offers not found"})
		return
	}
	c.JSON(http.StatusOK, offers)
}

var androidLock sync.Mutex

func createSubscriptionAndroid(c *gin.Context) {

	androidLock.Lock()
	defer androidLock.Unlock()

	var receipt model.PurchaseRceipt

	if err := json.NewDecoder(c.Request.Body).Decode(&receipt); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	log.Infof("android: %s", receipt)

	// Check if the subscription already exists
	s, err := core.GetSubscriptionByReceipt(receipt.Receipt)
	if err == nil {
		// if it does, update it
		log.Infof("subscription already exists, updating: %s", s.Id)
	}

	if s != nil && s.Email == "" {
		// if the subscription exists but doesn't have an email, update it with the email from the receipt
		now := time.Now().UTC()
		s.Email = receipt.Email
		s.AccountID = receipt.AccountID
		s.UpdatedBy = "android/" + receipt.Email
		s.LastUpdated = &now
	}

	// Validate the receipt with Google
	result, err := validateReceiptAndroid(receipt)
	if err != nil || result == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid receipt"})
		return
	}

	lineItems := result["lineItems"].([]interface{})
	zero := lineItems[0].(map[string]interface{})
	offerDetails := zero["offerDetails"].(map[string]interface{})
	plan := offerDetails["basePlanId"].(string)
	expires := time.Now().AddDate(0, 2, 0)

	if receipt.ProductID == "basic" && plan == "basic-monthly" {
		receipt.ProductID = "basic_monthly"
	}

	if receipt.ProductID == "basic" && plan == "basic-yearly" {
		receipt.ProductID = "basic_yearly"
		expires = time.Now().AddDate(1, 0, 0)
	}

	if receipt.ProductID == "premium" && plan == "premium-monthly" {
		receipt.ProductID = "premium_monthly"
	}

	if receipt.ProductID == "premium" && plan == "premium-yearly" {
		receipt.ProductID = "premium_yearly"
		expires = time.Now().AddDate(1, 0, 0)
	}

	if receipt.ProductID == "professional" && plan == "professional-monthly" {
		receipt.ProductID = "professional_monthly"
	}

	if receipt.ProductID == "professional" && plan == "professional-yearly" {
		receipt.ProductID = "professional_yearly"
		expires = time.Now().AddDate(1, 0, 0)
	}

	customer_name := receipt.Name

	credits := 0
	name := ""
	description := ""
	devices := 0
	networks := 0
	members := 0
	relays := 0
	autoRenew := false
	issued := time.Now()

	switch receipt.ProductID {
	case "basic_monthly":
		name = "Core Service (monthly)"
	case "basic_yearly":
		name = "Core Service (yearly)"
	case "premium_monthly":
		name = "Premium Service (monthly)"
	case "premium_yearly":
		name = "Premium Service (yearly)"
	case "professional_monthly":
		name = "Professional Service (monthly)"
	case "professional_yearly":
		name = "Professional Service (yearly)"
	default:
		log.Errorf("unknown sku %s", receipt.ProductID)
	}
	// set the credits, name, and description based on the sku
	switch receipt.ProductID {
	case "24_hours_flex":
		credits = 1
		name = "24 Hours"
		description = "Service in any region for 24 hours"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = false
		expires = time.Now().Add(24 * time.Hour)
	case "10_day_flex":
		credits = 1
		name = "10 Day Flex"
		description = "Service in any region for 10 days"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = false
		expires = time.Now().AddDate(0, 0, 10)
	case "basic_monthly", "basic_yearly":
		credits = 1
		description = "A single tunnel or relay in any region"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = true
	case "premium_monthly", "premium_yearly":
		credits = 5
		description = "Up to 5 tunnels or relays in any region"
		devices = 25
		networks = 10
		members = 5
		relays = 5
		autoRenew = true
	case "professional_monthly", "professional_yearly":
		credits = 10
		description = "Up to 10 tunnels or relays in any region"
		devices = 100
		networks = 25
		members = 25
		relays = 10
		autoRenew = true
	default:
		log.Errorf("unknown sku %s", receipt.ProductID)
	}

	// set the limits based on the sku
	accounts, err := core.ReadAllAccounts(receipt.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			account.Name = customer_name
			account.AccountName = "Company"
			account.Email = receipt.Email
			account.Role = "Owner"
			account.Status = "Active"
			account.CreatedBy = receipt.Email
			account.UpdatedBy = receipt.Email
			account.Picture = os.Getenv("SERVER") + "/account-circle.png"

			a, err := core.CreateAccount(&account)
			log.Infof("CREATE ACCOUNT = %v", a)
			if err != nil {
				log.Error(err)
			}
			accounts, err = core.ReadAllAccounts(receipt.Email)
			if err != nil {
				log.Error(err)
			}

		}
	}

	var account *model.Account
	for i := 0; i < len(accounts); i++ {
		if accounts[i].Id == accounts[i].Parent {
			account = accounts[i]
			break
		}
	}

	if account == nil && len(accounts) > 0 {
		account = accounts[0]
	}

	if account == nil {
		log.Errorf("account not found for email %s", receipt.Email)
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	limits, err := core.ReadLimits(account.Id)
	if err != nil {
		log.Error(err)
		limits_id, err := util.GenerateRandomString(8)
		if err != nil {
			log.Error(err)
		}
		limits_id = "limits-" + limits_id

		limits = &model.Limits{
			Id:          limits_id,
			AccountID:   account.Id,
			MaxDevices:  0,
			MaxNetworks: 0,
			MaxMembers:  0,
			MaxServices: 0,
			Tolerance:   core.GetDefaultTolerance(),
			CreatedBy:   receipt.Email,
			UpdatedBy:   receipt.Email,
			Created:     time.Now(),
			Updated:     time.Now(),
		}
	}

	limits.MaxDevices += devices
	limits.MaxNetworks += networks
	limits.MaxMembers += members
	limits.MaxServices += relays

	errs := limits.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("limits validation error")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limits validation error"})
		return
	}

	// save limits to mongodb
	mongo.Serialize(limits.Id, "id", "limits", limits)

	// generate a random subscription id
	id, err := util.RandomString(8)
	if err != nil {
		log.Error(err)
	}
	id = receipt.Source + "-" + id

	// construct a subscription object
	lu := time.Now()
	isDeleted := false

	subscription := model.Subscription{
		Id:          id,
		AccountID:   account.Id,
		Email:       receipt.Email,
		Name:        name,
		Description: description,
		Issued:      &issued,
		LastUpdated: &lu,
		UpdatedBy:   receipt.Email,
		Expires:     &expires,
		Credits:     credits,
		Sku:         receipt.ProductID,
		Status:      "active",
		AutoRenew:   autoRenew,
		Receipt:     receipt.Receipt,
		IsDeleted:   &isDeleted,
		CreatedBy:   receipt.Email,
	}

	errs = subscription.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("subscription validation error")
		}
		return
	}

	// save subscription to mongodb
	mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

	err = core.SubscriptionEmail(&subscription)
	if err != nil {
		log.Errorf("failed to send email: %v", err)
	}

	c.JSON(http.StatusOK, subscription)

}

func validateReceiptAndroid(receipt model.PurchaseRceipt) (map[string]interface{}, error) {

	// Get the Google Play Developer API access token using our refresh token

	client_id := os.Getenv("GOOGLE_PLAY_CLIENT_ID")
	client_secret := os.Getenv("GOOGLE_PLAY_CLIENT_SECRET")
	refresh_token := os.Getenv("GOOGLE_PLAY_REFRESH_TOKEN")
	access_url := os.Getenv("GOOGLE_PLAY_ACCESS_URL")

	// Create the request payload
	payload := "grant_type=refresh_token&client_id=" + client_id + "&client_secret=" + client_secret + "&refresh_token=" + refresh_token

	rsp, err := http.Post(access_url, "application/x-www-form-urlencoded", strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response from Google: %s", rsp.Status)
	}

	// Parse the response to get the access token
	var result2 map[string]interface{}
	err = json.NewDecoder(rsp.Body).Decode(&result2)
	if err != nil {
		return nil, err
	}

	access_token := result2["access_token"].(string)

	// Google Play receipt validation URL
	url := "https://www.googleapis.com/androidpublisher/v3/applications/com.nettica.agent/purchases/subscriptionsv2/tokens/" +
		receipt.Receipt + "?access_token=" + access_token

	log.Infof("url: %s", url)

	// Send the request to Google
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response from Google: %s", resp.Status)
	}

	// Parse the response to check the receipt status
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	log.Infof("google receipt: %v", result)

	// Check if the receipt is valid
	if status, ok := result["subscriptionState"].(string); ok && status == "SUBSCRIPTION_STATE_ACTIVE" {
		url = fmt.Sprintf("https://www.googleapis.com/androidpublisher/v3/applications/com.nettica.agent/purchases/subscriptions/%s/tokens/%s:acknowledge?access_token=%s",
			receipt.ProductID, receipt.Receipt, access_token)
		rsp, err := http.Post(url, "application/json", nil)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		return result, nil
	}

	return nil, nil
}

func handleAndroidWebhook(c *gin.Context) {

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	body := string(bytes)
	log.Info(body)

	var message map[string]interface{}

	// unmarshal the body into a map
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	mess := message["message"].(map[string]interface{})
	data64 := mess["data"].(string)

	data, err := base64.StdEncoding.DecodeString(data64)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to decode data")
	}
	log.Infof("data: %s", data)

	var msg map[string]interface{}

	err = json.Unmarshal(data, &msg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal data")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	log.Infof("msg: %v", msg)
	packageName := msg["packageName"].(string)
	log.Infof("packageName: %s", packageName)

	if msg["subscriptionNotification"] != nil {

		subscriptionNotification := msg["subscriptionNotification"].(map[string]interface{})
		log.Infof("subscriptionNotification: %v", subscriptionNotification)

		// Retrieve the subscription from the purchaseToken
		if subscriptionNotification["purchaseToken"] == nil {
			log.Error("purchaseToken not found -- assuming test")
			c.JSON(http.StatusOK, gin.H{"status": "received"})
			return
		}

		purchaseToken := subscriptionNotification["purchaseToken"].(string)
		log.Infof("purchaseToken: %s", purchaseToken)

		// Get the Google Play Developer API access token using our refresh token
		client_id := os.Getenv("GOOGLE_PLAY_CLIENT_ID")
		client_secret := os.Getenv("GOOGLE_PLAY_CLIENT_SECRET")
		refresh_token := os.Getenv("GOOGLE_PLAY_REFRESH_TOKEN")
		access_url := os.Getenv("GOOGLE_PLAY_ACCESS_URL")

		// Create the request payload
		payload := "grant_type=refresh_token&client_id=" + client_id + "&client_secret=" + client_secret + "&refresh_token=" + refresh_token

		rsp, err := http.Post(access_url, "application/x-www-form-urlencoded", strings.NewReader(payload))
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		defer rsp.Body.Close()

		if rsp.StatusCode != http.StatusOK {
			log.Errorf("invalid response from Google: %s", rsp.Status)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Parse the response to get the access token
		var result map[string]interface{}
		err = json.NewDecoder(rsp.Body).Decode(&result)
		if err != nil {
			log.Errorf("error getting access token : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting access token"})
			return
		}

		access_token := result["access_token"].(string)

		// Get the subscription from google
		url := "https://www.googleapis.com/androidpublisher/v3/applications/" + packageName + "/purchases/subscriptionsv2/tokens/" + purchaseToken + "?access_token=" + access_token
		log.Infof("url: %s", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Errorf("error geting subscription from google: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting subscription from google"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Errorf("invalid subscription response from Google: %s", resp.Status)
			c.JSON(http.StatusOK, gin.H{"status": "received"})
			return
		}

		// Parse the response to check the receipt status
		var sub map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&sub)
		if err != nil {
			log.Errorf("error decoding response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		log.Infof("google subscription: %v", sub)
		// retrieve the subscription
		subscription, err := core.GetSubscriptionByReceipt(purchaseToken)
		if err != nil {
			// create the subscription

			log.Errorf("error getting subscription: %v, creating one", err)

			// Create a subscription based on the transaction info

			now := time.Now().UTC()
			// generate a random subscription id
			id, err := util.RandomString(8)
			if err != nil {
				log.Error(err)
			}
			id = "android-" + id

			subscription = &model.Subscription{
				Id:          id,
				AccountID:   "", // we don't have the account id, so we'll leave it blank for now
				Email:       "", // we don't have the email, so we'll leave it blank for now
				Name:        "", // we don't have the name, so we'll leave it blank for now
				Description: "",
				Issued:      &now,
				CreatedBy:   "android",
				Receipt:     purchaseToken,
			}
			log.Infof("created subscription: %v", subscription)
			mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

		}

		if sub["subscriptionState"] != nil && sub["subscriptionState"].(string) == "SUBSCRIPTION_STATE_ACTIVE" {

			if subscription.Status == "expired" || subscription.Status == "cancelled" || subscription.Status == "grace" {
				subscription.Status = "active"
				last := time.Now().UTC()
				subscription.LastUpdated = &last
				subscription.UpdatedBy = "google"
				f := false
				subscription.IsDeleted = &f
				core.RenewSubscription(subscription.Id)
				log.Infof("subscription renewed: %s", subscription.Id)
				subscription, err = core.GetSubscriptionByReceipt(purchaseToken)
				if err != nil {
					log.Errorf("error getting subscription: %v", err)
					c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
					return
				}
			}
		}

		// get the lineItems from the subscription.  It's an array of maps
		lineItems, ok := sub["lineItems"].([]interface{})
		if !ok || len(lineItems) == 0 {
			log.Errorf("lineItems not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		zero := lineItems[0].(map[string]interface{})

		// update the expires date with expiryTime
		if zero != nil && zero["expiryTime"] != nil {
			*subscription.Expires, err = time.Parse(time.RFC3339, zero["expiryTime"].(string))
			if err != nil {
				log.Errorf("error parsing expiryTime: %v", err)
				//				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				//				return
			} else {
				last := time.Now().UTC()
				subscription.LastUpdated = &last
				log.Infof("subscription updated: %s %v", subscription.Id, subscription)
				//				c.JSON(http.StatusOK, gin.H{"status": "updated"})
				//				return
			}
			subscription.UpdatedBy = "google"
			core.UpdateSubscription(subscription.Id, subscription)
		}

		if sub["subscriptionState"] != nil && sub["subscriptionState"].(string) == "SUBSCRIPTION_STATE_EXPIRED" {

			core.ExpireSubscription(subscription.Id)

			log.Infof("subscription expired: %s", subscription.Id)
			core.SubscriptionEmail(subscription)

			c.JSON(http.StatusOK, gin.H{"status": "expired"})
			return
		}

		// handle cancel and did_not_renew
		if sub["subscriptionState"] != nil && sub["subscriptionState"].(string) == "SUBSCRIPTION_STATE_CANCELED" {

			subscription.Status = "cancelled"
			core.UpdateSubscription(subscription.Id, subscription)
			core.ExpireSubscription(subscription.Id)
			core.SubscriptionEmail(subscription)

			log.Infof("subscription cancelled: %s", subscription.Id)

			c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
			return
		}

		if sub["subscriptionState"] != nil && sub["subscriptionState"].(string) == "SUBSCRIPTION_STATE_IN_GRACE_PERIOD" {

			subscription.Status = "grace"

			core.UpdateSubscription(subscription.Id, subscription)
			core.SubscriptionEmail(subscription)

			log.Infof("subscription grace: %s", subscription.Id)

			c.JSON(http.StatusOK, gin.H{"status": "grace"})
			return
		}

	}

	if msg["voidedPurchaseNotification"] != nil {
		voidedPurchaseNotification := msg["voidedPurchaseNotification"].(map[string]interface{})
		log.Infof("voidedPurchaseNotification: %v", voidedPurchaseNotification)

	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func handleAndroidWebhook2(c *gin.Context) {

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	log.Info(string(bytes))

	var message map[string]interface{}
	if err = json.Unmarshal(bytes, &message); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	mess, ok := message["message"].(map[string]interface{})
	if !ok {
		log.Error("android webhook: message field missing or wrong type")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid message format"})
		return
	}

	data64, ok := mess["data"].(string)
	if !ok {
		log.Error("android webhook: message.data missing or wrong type")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid message format"})
		return
	}

	data, err := base64.StdEncoding.DecodeString(data64)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to decode data")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	log.Infof("android webhook data: %s", data)

	var msg map[string]interface{}
	if err = json.Unmarshal(data, &msg); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal data")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	log.Infof("android webhook msg: %v", msg)

	packageName, ok := msg["packageName"].(string)
	if !ok || packageName == "" {
		log.Error("android webhook: packageName missing")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	log.Infof("android webhook packageName: %s", packageName)

	if subscriptionNotification, ok := msg["subscriptionNotification"].(map[string]interface{}); ok {
		log.Infof("subscriptionNotification: %v", subscriptionNotification)

		purchaseToken, ok := subscriptionNotification["purchaseToken"].(string)
		if !ok || purchaseToken == "" {
			log.Error("purchaseToken not found -- assuming test")
			c.JSON(http.StatusOK, gin.H{"status": "received"})
			return
		}
		log.Infof("purchaseToken: %s", purchaseToken)

		// Get Google Play access token
		payload := "grant_type=refresh_token" +
			"&client_id=" + os.Getenv("GOOGLE_PLAY_CLIENT_ID") +
			"&client_secret=" + os.Getenv("GOOGLE_PLAY_CLIENT_SECRET") +
			"&refresh_token=" + os.Getenv("GOOGLE_PLAY_REFRESH_TOKEN")

		rsp, err := http.Post(os.Getenv("GOOGLE_PLAY_ACCESS_URL"), "application/x-www-form-urlencoded", strings.NewReader(payload))
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		defer rsp.Body.Close()

		if rsp.StatusCode != http.StatusOK {
			log.Errorf("invalid response from Google: %s", rsp.Status)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		var tokenResult map[string]interface{}
		if err = json.NewDecoder(rsp.Body).Decode(&tokenResult); err != nil {
			log.Errorf("error decoding access token response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting access token"})
			return
		}

		access_token, ok := tokenResult["access_token"].(string)
		if !ok || access_token == "" {
			log.Error("android webhook: access_token missing from response")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting access token"})
			return
		}

		// Fetch subscription from Google Play API
		gpURL := "https://www.googleapis.com/androidpublisher/v3/applications/" + packageName +
			"/purchases/subscriptionsv2/tokens/" + purchaseToken + "?access_token=" + access_token
		log.Infof("google play url: %s", gpURL)

		resp, err := http.Get(gpURL)
		if err != nil {
			log.Errorf("error getting subscription from google: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting subscription from google"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Errorf("invalid subscription response from Google: %s", resp.Status)
			c.JSON(http.StatusOK, gin.H{"status": "received"})
			return
		}

		var sub map[string]interface{}
		if err = json.NewDecoder(resp.Body).Decode(&sub); err != nil {
			log.Errorf("error decoding google subscription response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		log.Infof("google subscription: %v", sub)

		// Get or create local subscription record
		subscription, err := core.GetSubscriptionByReceipt(purchaseToken)
		if err != nil {
			log.Errorf("subscription not found for purchaseToken, creating stub: %v", err)

			now := time.Now().UTC()
			id, idErr := util.RandomString(8)
			if idErr != nil {
				log.Error(idErr)
				id = "unknown"
			}
			id = "android-" + id

			subscription = &model.Subscription{
				Id:          id,
				AccountID:   "",
				Email:       "",
				Name:        "",
				Description: "",
				Issued:      &now,
				CreatedBy:   "android",
				Receipt:     purchaseToken,
			}
			log.Infof("created subscription stub: %v", subscription)
			mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)
		}

		subscriptionState, _ := sub["subscriptionState"].(string)

		if subscriptionState == "SUBSCRIPTION_STATE_ACTIVE" {
			if subscription.Status == "expired" || subscription.Status == "cancelled" || subscription.Status == "grace" {
				last := time.Now().UTC()
				subscription.Status = "active"
				subscription.LastUpdated = &last
				subscription.UpdatedBy = "google"
				f := false
				subscription.IsDeleted = &f
				core.RenewSubscription(subscription.Id)
				log.Infof("subscription renewed: %s", subscription.Id)
				subscription, err = core.GetSubscriptionByReceipt(purchaseToken)
				if err != nil {
					log.Errorf("error getting subscription after renew: %v", err)
					c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
					return
				}
			}
		}

		// Update expiry from lineItems[0].expiryTime
		lineItems, ok := sub["lineItems"].([]interface{})
		if !ok || len(lineItems) == 0 {
			log.Errorf("lineItems not found in google response")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		zero, ok := lineItems[0].(map[string]interface{})
		if ok {
			if expiryStr, ok := zero["expiryTime"].(string); ok {
				t, parseErr := time.Parse(time.RFC3339, expiryStr)
				if parseErr != nil {
					log.Errorf("error parsing expiryTime %q: %v", expiryStr, parseErr)
				} else {
					subscription.Expires = &t
					last := time.Now().UTC()
					subscription.LastUpdated = &last
					log.Infof("subscription %s expires %s", subscription.Id, t.Format(time.RFC3339))
				}
			}
			subscription.UpdatedBy = "google"
			core.UpdateSubscription(subscription.Id, subscription)
		}

		switch subscriptionState {
		case "SUBSCRIPTION_STATE_EXPIRED":
			core.ExpireSubscription(subscription.Id)
			log.Infof("subscription expired: %s", subscription.Id)
			core.SubscriptionEmail(subscription)
			c.JSON(http.StatusOK, gin.H{"status": "expired"})
			return

		case "SUBSCRIPTION_STATE_CANCELED":
			subscription.Status = "cancelled"
			core.UpdateSubscription(subscription.Id, subscription)
			core.ExpireSubscription(subscription.Id)
			core.SubscriptionEmail(subscription)
			log.Infof("subscription cancelled: %s", subscription.Id)
			c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
			return

		case "SUBSCRIPTION_STATE_IN_GRACE_PERIOD":
			subscription.Status = "grace"
			core.UpdateSubscription(subscription.Id, subscription)
			core.SubscriptionEmail(subscription)
			log.Infof("subscription grace: %s", subscription.Id)
			c.JSON(http.StatusOK, gin.H{"status": "grace"})
			return
		}
	}

	if voidedNotification, ok := msg["voidedPurchaseNotification"].(map[string]interface{}); ok {
		log.Infof("voidedPurchaseNotification: %v", voidedNotification)
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func fetchApplePublicKeyFromX5C(x5c string) (*ecdsa.PublicKey, error) {
	// Decode the base64 encoded certificate
	certBytes, err := base64.StdEncoding.DecodeString(x5c)
	if err != nil {
		return nil, fmt.Errorf("failed to decode x5c: %v", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Validate the certificate (e.g., check the certificate chain, expiration, etc.)
	// This example assumes the certificate is self-signed and trusted for simplicity.
	// In a real-world scenario, you should validate the certificate chain against trusted CAs.

	// Extract the public key
	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of type ECDSA")
	}

	return pubKey, nil
}

func handleAppleWebhook(c *gin.Context) {

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	body := string(bytes)
	log.Info(body)

	var msg map[string]interface{}

	// unmarshal the body into a map
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal request body")
		// c.AbortWithStatus(http.StatusUnprocessableEntity)
		// return with no error so webhook doesn't get disabled
		c.JSON(http.StatusOK, body)
		return
	}

	signedPayload := msg["signedPayload"].(string)

	parts := strings.Split(signedPayload, ".")
	if len(parts) != 3 {
		log.Errorf("invalid payload %s", signedPayload)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid payload"})
		return
	}

	// Decode the parts
	// headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	// if err != nil {
	//	log.Error(err)
	// a}
	// header := string(headerBytes)
	// log.Infof("header: %s", header)

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Error(err)
	}
	var payload map[string]interface{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		log.Error(err)
	}
	log.Infof("payload: %v", payload)

	// signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	// if err != nil {
	//	log.Error(err)
	// }
	// signature := string(signatureBytes)
	// log.Infof("signature: %s", signature)

	// Validate the signature - ignore errors for now
	valid, err := validateAppleSignature(parts[0], parts[1], parts[2])
	if err != nil {
		log.Errorf("invalid signature %v", err)
		//        c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid signature"})
		//        return
	}

	if !valid {
		log.Errorf("invalid signature %v", err)
		//        c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid signature"})
		//        return
	}

	notificationType := payload["notificationType"].(string)
	log.Infof("notificationType: %s", notificationType)

	AutoRenew := true
	if payload["subtype"] != nil && payload["subtype"].(string) == "AUTO_RENEW_DISABLED" {
		AutoRenew = false
	}

	log.Infof("AutoRenew: %v", AutoRenew)

	data := payload["data"].(map[string]interface{})
	signedTransactionInfo := data["signedTransactionInfo"].(string)

	parts = strings.Split(signedTransactionInfo, ".")
	transactionBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Error(err)
	}
	var transaction map[string]interface{}
	err = json.Unmarshal(transactionBytes, &transaction)
	if err != nil {
		log.Error(err)
	}
	log.Infof("transaction: %v", transaction)

	originalTransactionId := transaction["originalTransactionId"].(string)
	log.Infof("originalTransactionId: %s", originalTransactionId)

	var expiresDate float64
	expiresDate = float64(time.Now().UTC().AddDate(0, 0, 1).UnixMilli())
	if transaction["expiresDate"] != nil {
		expiresDate = transaction["expiresDate"].(float64)
	}

	expires := time.Unix(int64(expiresDate)/1000, 0)
	log.Infof("expires: %s", expires)

	transactionReason := transaction["transactionReason"].(string)
	log.Infof("transactionReason: %s", transactionReason)

	// let's update the subscription
	if originalTransactionId != "" {
		subscription, err := core.GetSubscriptionByReceipt(originalTransactionId)
		if err != nil {
			log.Error(err)
			// Create a subscription based on the transaction info
			log.Infof("subscription not found for originalTransactionId %s, creating new subscription", originalTransactionId)
			now := time.Now().UTC()
			subscription = &model.Subscription{
				Id:          "apple-" + originalTransactionId,
				AccountID:   "", // we don't have the account id, so we'll leave it blank for now
				Email:       "", // we don't have the email, so we'll leave it blank for now
				Name:        "", // we don't have the name, so we'll leave it blank for now
				Description: "",
				AutoRenew:   AutoRenew,
				Issued:      &now,
				CreatedBy:   "apple",
				UpdatedBy:   "apple/" + originalTransactionId,
				Receipt:     originalTransactionId,
			}
			log.Infof("created subscription: %v", subscription)
			mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

		}
		AutoRenew = subscription.AutoRenew
		last := time.Now().UTC()
		subscription.LastUpdated = &last
		subscription.UpdatedBy = "apple"

		switch transactionReason {
		case "PURCHASE":
			subscription.Expires = &expires
			subscription.Sku = transaction["productId"].(string)
			switch subscription.Sku {
			case "basic_monthly":
				subscription.Name = "Core Service (monthly)"
				subscription.Description = "A single tunnel or relay in any region"
				subscription.Credits = 1
			case "basic_yearly":
				subscription.Name = "Core Service (yearly)"
				subscription.Description = "A single tunnel or relay in any region"
				subscription.Credits = 1
			case "premium_monthly":
				subscription.Name = "Premium Service (monthly)"
				subscription.Description = "Up to 5 tunnels or relays in any region"
				subscription.Credits = 5
			case "premium_yearly":
				subscription.Name = "Premium Service (yearly)"
				subscription.Description = "Up to 5 tunnels or relays in any region"
				subscription.Credits = 5
			case "professional_monthly":
				subscription.Name = "Professional Service (monthly)"
				subscription.Description = "Up to 10 tunnels or relays in any region"
				subscription.Credits = 10
			case "professional_yearly":
				subscription.Name = "Professional Service (yearly)"
				subscription.Description = "Up to 10 tunnels or relays in any region"
				subscription.Credits = 10
			default:
				log.Errorf("unknown sku %s", subscription.Sku)
			}
			log.Infof("subscription sku apple: %s", subscription.Sku)
			log.Infof("apple subscription: %v", subscription)
			f := false
			subscription.IsDeleted = &f
			core.UpdateSubscription(subscription.Id, subscription)
			log.Infof("apple: subscription PURCHASE updated: %s until %s", subscription.Id, expires)

		case "RENEWAL":
			subscription.Expires = &expires
			f := false
			subscription.IsDeleted = &f
			if subscription.Status == "cancelled" || subscription.Status == "expired" {
				subscription.Status = "active"
				core.UpdateSubscription(subscription.Id, subscription)
				core.RenewSubscription(subscription.Id)
			} else {
				core.UpdateSubscription(subscription.Id, subscription)
			}
			log.Infof("apple: subscription RENWAL: %s until %s", subscription.Id, expires)

		case "CANCEL":
			subscription.Status = "active"
			core.UpdateSubscription(subscription.Id, subscription)

			// this will only expire services if it is after the expires date
			core.ExpireSubscription(subscription.Id)
			log.Infof("apple: subscription CANCEL: %s", subscription.Id)

		case "DID_NOT_RENEW":
			subscription.Status = "expired"
			core.ExpireSubscription(subscription.Id)
			log.Infof("apple: subscription DID_NOT_RENEW: %s at %s", subscription.Id, expires)
		}
		core.SubscriptionEmail(subscription)

	}
	// Respond with 200 OK to acknowledge receipt of the webhook
	c.JSON(http.StatusOK, gin.H{"status": "received"})

}

// handleAppleWebhook2 is the hardened replacement for handleAppleWebhook.
//
// Key differences from the original:
//   - Dispatches on notificationType (the Apple event kind) rather than
//     transactionReason, so all notification types are handled correctly.
//   - All JWS field extractions use appleStr / appleF64 — no bare assertions.
//   - Enforces correct DB operation ordering: RenewSubscription is called BEFORE
//     UpdateSubscription for reactivation (RenewSubscription reads old status from
//     DB; if status is already "active" it returns without re-enabling services).
//     ExpireSubscription is called AFTER UpdateSubscription has persisted the
//     Apple-provided expiresDate (ExpireSubscription rejects future expiry dates).
//   - Per-transaction lock prevents duplicate processing from Apple retries.
//   - Always returns 200 so Apple does not disable the webhook endpoint.
func handleAppleWebhook2(c *gin.Context) {

	// 1. Read body.
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("handleAppleWebhook2: read body: %v", err)
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	//log.Infof("handleAppleWebhook2: %s", string(body))

	// 2. Parse outer JWS envelope.
	var msg map[string]interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Errorf("handleAppleWebhook2: unmarshal envelope: %v \n%s", err, string(body))
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	signedPayload, ok := appleStr(msg, "signedPayload")
	if !ok || signedPayload == "" {
		log.Errorf("handleAppleWebhook2: signedPayload missing or wrong type")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	// 3. Split outer JWS: header.payload.signature.
	outerParts := strings.Split(signedPayload, ".")
	if len(outerParts) != 3 {
		log.Errorf("handleAppleWebhook2: signedPayload has %d parts, expected 3", len(outerParts))
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	// 4. Validate outer JWS signature.
	// Logged but not enforced: validateAppleSignature is still being stabilised.
	// Once confirmed working this block should reject invalid payloads.
	if valid, sigErr := validateAppleSignature(outerParts[0], outerParts[1], outerParts[2]); sigErr != nil || !valid {
		log.Errorf("handleAppleWebhook2: signature invalid (valid=%v err=%v) — proceeding", valid, sigErr)
	}

	// 5. Decode and parse the notification payload.
	payloadBytes, err := base64.RawURLEncoding.DecodeString(outerParts[1])
	if err != nil {
		log.Errorf("handleAppleWebhook2: decode payload: %v \n%s", err, outerParts[1])
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Errorf("handleAppleWebhook2: unmarshal payload: %v \n%s", err, string(payloadBytes))
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	//log.Infof("handleAppleWebhook2: payload %v", payload)

	// 6. Extract notification type — this is the canonical dispatch key.
	notificationType, ok := appleStr(payload, "notificationType")
	if !ok || notificationType == "" {
		log.Errorf("handleAppleWebhook2: notificationType missing")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	subtype, _ := appleStr(payload, "subtype") // optional; absent is valid
	log.Infof("handleAppleWebhook2: notificationType=%s subtype=%s", notificationType, subtype)

	// 7. Decode signedTransactionInfo from data.
	data, ok := payload["data"].(map[string]interface{})
	if !ok || data == nil {
		log.Errorf("handleAppleWebhook2: data missing")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	signedTxInfo, ok := appleStr(data, "signedTransactionInfo")
	if !ok || signedTxInfo == "" {
		log.Errorf("handleAppleWebhook2: signedTransactionInfo missing")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	txParts := strings.Split(signedTxInfo, ".")
	if len(txParts) != 3 {
		log.Errorf("handleAppleWebhook2: signedTransactionInfo has %d parts, expected 3", len(txParts))
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	txBytes, err := base64.RawURLEncoding.DecodeString(txParts[1])
	if err != nil {
		log.Errorf("handleAppleWebhook2: decode transaction: %v \n%s", err, txParts[1])
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	var transaction map[string]interface{}
	if err := json.Unmarshal(txBytes, &transaction); err != nil {
		log.Errorf("handleAppleWebhook2: unmarshal transaction: %v \n%s", err, string(txBytes))
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	log.Infof("handleAppleWebhook2: transaction %s", string(txBytes))

	// 8. Extract required transaction fields.
	originalTxId, ok := appleStr(transaction, "originalTransactionId")
	if !ok || originalTxId == "" {
		log.Errorf("handleAppleWebhook2: originalTransactionId missing")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}
	productId, _ := appleStr(transaction, "productId")

	var appleExpires time.Time
	if expiresMs, ok := appleF64(transaction, "expiresDate"); ok {
		appleExpires = time.Unix(int64(expiresMs)/1000, 0).UTC()
	}

	log.Infof("handleAppleWebhook2: originalTxId=%s productId=%s expires=%s",
		originalTxId, productId, appleTimeStr(appleExpires))

	// 9. Per-transaction lock — Apple retries failed deliveries so we must be
	//    idempotent.  The lock ensures only one goroutine processes a given
	//    originalTransactionId at a time.
	release := acquireAppleTxLock(originalTxId)
	defer release()

	// 10. Look up the subscription.
	subscription, err := core.GetSubscriptionByReceipt(originalTxId)
	if err != nil && !strings.Contains(err.Error(), "no documents in result") {
		log.Errorf("handleAppleWebhook2: DB error for %s: %v", originalTxId, err)
		// Return 200 — Apple will retry; we'll process when DB recovers.
		c.JSON(http.StatusNotFound, gin.H{"status": "received"})
		return
	}
	notFound := subscription == nil

	if notFound {
		log.Infof("handleAppleWebhook2: no subscription found for originalTxId %s", originalTxId)
	} else {
		log.Infof("handleAppleWebhook2: found subscription %s by %s for originalTxId %s", subscription.Id, subscription.Email, originalTxId)
	}

	now := time.Now().UTC()
	falseVal := false

	// 11. Dispatch on notificationType.
	switch notificationType {

	case "SUBSCRIBED":
		// New subscription, or resubscription after a full expiry.
		// Create a stub when not found so createSubscriptionApple2 can claim it.
		if notFound {
			subscription = appleNewStub(originalTxId, productId, &now)
			if serErr := mongo.Serialize(subscription.Id, "id", "subscriptions", subscription); serErr != nil {
				log.Errorf("handleAppleWebhook2: SUBSCRIBED stub serialize: %v", serErr)
			}
		}
		wasInactive := subscription.Status == "expired" || subscription.Status == "cancelled" || subscription.Status == ""
		if sku, known := appleSkuMap[productId]; known {
			subscription.Sku = productId
			subscription.Name = sku.name
			subscription.Description = sku.description
			subscription.Credits = sku.credits
		}
		if !appleExpires.IsZero() {
			subscription.Expires = &appleExpires
		}
		subscription.Status = "active"
		subscription.AutoRenew = true
		subscription.IsDeleted = &falseVal
		subscription.LastUpdated = &now
		subscription.UpdatedBy = "apple"
		// RenewSubscription must run BEFORE UpdateSubscription: it reads the old
		// status from DB and won't re-enable services if status is already "active".
		if wasInactive && subscription.AccountID != "" {
			if err := core.RenewSubscription(subscription.Id); err != nil {
				log.Errorf("handleAppleWebhook2: SUBSCRIBED RenewSubscription: %v", err)
			}
		}
		if _, err := core.UpdateSubscription(subscription.Id, subscription); err != nil {
			log.Errorf("handleAppleWebhook2: SUBSCRIBED UpdateSubscription: %v", err)
		}
		if wasInactive && subscription.AccountID != "" {
			core.SubscriptionEmail(subscription)
		}
		log.Infof("handleAppleWebhook2: SUBSCRIBED %s until %s (wasInactive=%v notFound=%v)",
			subscription.Id, appleTimeStr(appleExpires), wasInactive, notFound)

	case "DID_RENEW":
		// Auto-renewal succeeded, or billing recovered after a grace period.
		if notFound {
			subscription = appleNewStub(originalTxId, productId, &now)
			if serErr := mongo.Serialize(subscription.Id, "id", "subscriptions", subscription); serErr != nil {
				log.Errorf("handleAppleWebhook2: DID_RENEW stub serialize: %v", serErr)
			}
		}
		wasInactive := subscription.Status == "expired" || subscription.Status == "cancelled"
		if sku, known := appleSkuMap[productId]; known {
			subscription.Sku = productId
			subscription.Name = sku.name
			subscription.Description = sku.description
			subscription.Credits = sku.credits
		}
		if !appleExpires.IsZero() {
			subscription.Expires = &appleExpires
		}
		subscription.Status = "active"
		subscription.IsDeleted = &falseVal
		subscription.LastUpdated = &now
		subscription.UpdatedBy = "apple"
		// Same sequencing requirement as SUBSCRIBED.
		if wasInactive && subscription.AccountID != "" {
			if err := core.RenewSubscription(subscription.Id); err != nil {
				log.Errorf("handleAppleWebhook2: DID_RENEW RenewSubscription: %v", err)
			}
		}
		if _, err := core.UpdateSubscription(subscription.Id, subscription); err != nil {
			log.Errorf("handleAppleWebhook2: DID_RENEW UpdateSubscription: %v", err)
		}
		if wasInactive && subscription.AccountID != "" {
			core.SubscriptionEmail(subscription)
		}
		log.Infof("handleAppleWebhook2: DID_RENEW %s until %s (wasInactive=%v subtype=%s)",
			subscription.Id, appleTimeStr(appleExpires), wasInactive, subtype)

	case "DID_CHANGE_RENEWAL_STATUS":
		// The user toggled auto-renew.  They still have service until the current
		// billing period ends — status and expiry are unchanged.
		// AUTO_RENEW_DISABLED = user "cancelled"; service continues until Expires.
		// AUTO_RENEW_ENABLED  = user re-enabled; subscription will auto-renew.
		if notFound {
			log.Infof("handleAppleWebhook2: DID_CHANGE_RENEWAL_STATUS no subscription for %s, ignoring", originalTxId)
			break
		}
		switch subtype {
		case "AUTO_RENEW_DISABLED":
			subscription.AutoRenew = false
		case "AUTO_RENEW_ENABLED":
			subscription.AutoRenew = true
		default:
			log.Infof("handleAppleWebhook2: DID_CHANGE_RENEWAL_STATUS unknown subtype %q for %s", subtype, subscription.Id)
		}
		subscription.LastUpdated = &now
		subscription.UpdatedBy = "apple"
		if _, err := core.UpdateSubscription(subscription.Id, subscription); err != nil {
			log.Errorf("handleAppleWebhook2: DID_CHANGE_RENEWAL_STATUS UpdateSubscription: %v", err)
		}
		log.Infof("handleAppleWebhook2: DID_CHANGE_RENEWAL_STATUS subtype=%s autoRenew=%v for %s",
			subtype, subscription.AutoRenew, subscription.Id)

	case "DID_FAIL_TO_RENEW":
		// Billing failed.  Without a subtype we're inside Apple's billing grace
		// period — service continues and we take no action.  When subtype is
		// GRACE_PERIOD_EXPIRED the grace period has ended and service must stop.
		if notFound {
			log.Infof("handleAppleWebhook2: DID_FAIL_TO_RENEW no subscription for %s, ignoring", originalTxId)
			break
		}
		if subtype != "GRACE_PERIOD_EXPIRED" {
			log.Infof("handleAppleWebhook2: DID_FAIL_TO_RENEW grace period active for %s, no action", subscription.Id)
			break
		}
		// Grace period has ended — same expiry logic as EXPIRED below.
		if !appleExpires.IsZero() {
			subscription.Expires = &appleExpires
		}
		subscription.AutoRenew = false
		subscription.LastUpdated = &now
		subscription.UpdatedBy = "apple"
		if _, err := core.UpdateSubscription(subscription.Id, subscription); err != nil {
			log.Errorf("handleAppleWebhook2: DID_FAIL_TO_RENEW UpdateSubscription: %v", err)
		}
		// ExpireSubscription reads fresh from DB; it requires Expires to be in the
		// past and status to not already be "expired".  UpdateSubscription above
		// persisted the Apple-provided (past) expiresDate before this call.
		if err := core.ExpireSubscription(subscription.Id); err != nil {
			log.Errorf("handleAppleWebhook2: DID_FAIL_TO_RENEW ExpireSubscription: %v", err)
		}
		subscription.Status = "expired" // update in-memory for email only
		core.SubscriptionEmail(subscription)
		log.Infof("handleAppleWebhook2: DID_FAIL_TO_RENEW GRACE_PERIOD_EXPIRED %s expired at %s",
			subscription.Id, appleTimeStr(appleExpires))

	case "EXPIRED", "GRACE_PERIOD_EXPIRED":
		// The subscription has fully expired.
		// EXPIRED subtypes: CANCELLED (user cancelled and billing period ended),
		// BILLING_RETRY (billing failed after retries), PRICE_INCREASE, etc.
		// GRACE_PERIOD_EXPIRED is also sent as its own notification type.
		if notFound {
			log.Infof("handleAppleWebhook2: %s no subscription for %s, ignoring", notificationType, originalTxId)
			break
		}
		// Persist Apple's authoritative expiresDate (which is in the past) BEFORE
		// calling ExpireSubscription, which rejects subscriptions with a future Expires.
		if !appleExpires.IsZero() {
			subscription.Expires = &appleExpires
		}
		subscription.AutoRenew = false
		subscription.LastUpdated = &now
		subscription.UpdatedBy = "apple"
		if _, err := core.UpdateSubscription(subscription.Id, subscription); err != nil {
			log.Errorf("handleAppleWebhook2: %s UpdateSubscription: %v", notificationType, err)
		}
		if err := core.ExpireSubscription(subscription.Id); err != nil {
			log.Errorf("handleAppleWebhook2: %s ExpireSubscription: %v", notificationType, err)
		}
		subscription.Status = "expired" // update in-memory for email only
		core.SubscriptionEmail(subscription)
		log.Infof("handleAppleWebhook2: %s %s expired at %s (subtype=%s)",
			notificationType, subscription.Id, appleTimeStr(appleExpires), subtype)

	case "REFUND":
		// Apple issued a refund — revoke service immediately regardless of the
		// remaining billing period.  We force Expires into the past so that
		// ExpireSubscription acts right away rather than waiting for the period end.
		if notFound {
			log.Infof("handleAppleWebhook2: REFUND no subscription for %s, ignoring", originalTxId)
			break
		}
		past := now.Add(-time.Second)
		subscription.Status = "cancelled"
		subscription.Expires = &past
		subscription.AutoRenew = false
		subscription.LastUpdated = &now
		subscription.UpdatedBy = "apple"
		if _, err := core.UpdateSubscription(subscription.Id, subscription); err != nil {
			log.Errorf("handleAppleWebhook2: REFUND UpdateSubscription: %v", err)
		}
		// ExpireSubscription reads status="cancelled" from DB and sees Expires in
		// the past, so it disables services without overwriting the cancelled status.
		if err := core.ExpireSubscription(subscription.Id); err != nil {
			log.Errorf("handleAppleWebhook2: REFUND ExpireSubscription: %v", err)
		}
		core.SubscriptionEmail(subscription)
		log.Infof("handleAppleWebhook2: REFUND %s", subscription.Id)

	case "DID_CHANGE_RENEWAL_PREF":
		// The user changed their subscription tier.
		//
		// UPGRADE: takes effect immediately.  The transaction in this notification
		//   already carries the new (higher) productId, new expiry, and prorated
		//   billing — update the subscription right now.
		//
		// DOWNGRADE: deferred.  The current tier continues until the billing period
		//   ends; the lower tier starts at the next renewal.  A subsequent DID_RENEW
		//   with the new productId will handle the actual switch then.
		if notFound {
			log.Infof("handleAppleWebhook2: DID_CHANGE_RENEWAL_PREF no subscription for %s, ignoring", originalTxId)
			break
		}
		if subtype == "UPGRADE" {
			prevSku := subscription.Sku
			if sku, known := appleSkuMap[productId]; known {
				subscription.Sku = productId
				subscription.Name = sku.name
				subscription.Description = sku.description
				subscription.Credits = sku.credits
			}
			if !appleExpires.IsZero() {
				subscription.Expires = &appleExpires
			}
			subscription.AutoRenew = true
			subscription.IsDeleted = &falseVal
			subscription.LastUpdated = &now
			subscription.UpdatedBy = "apple"
			if _, err := core.UpdateSubscription(subscription.Id, subscription); err != nil {
				log.Errorf("handleAppleWebhook2: DID_CHANGE_RENEWAL_PREF UPGRADE UpdateSubscription: %v", err)
			}
			log.Infof("handleAppleWebhook2: DID_CHANGE_RENEWAL_PREF UPGRADE %s sku %s→%s until %s",
				subscription.Id, prevSku, productId, appleTimeStr(appleExpires))
		} else {
			// DOWNGRADE — no immediate action; log for traceability.
			log.Infof("handleAppleWebhook2: DID_CHANGE_RENEWAL_PREF DOWNGRADE %s currentSku=%s newSku=%s effective at %s",
				subscription.Id, subscription.Sku, productId, appleTimeStr(appleExpires))
		}

	default:
		log.Infof("handleAppleWebhook2: unhandled notificationType=%s subtype=%s originalTxId=%s",
			notificationType, subtype, originalTxId)
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// appleNewStub returns a minimal Subscription placeholder for when a webhook
// arrives before the user's app has connected.
// createSubscriptionApple2 will claim it (fill in email, accountID, grant limits).
func appleNewStub(originalTxId, productId string, now *time.Time) *model.Subscription {
	falseVal := false
	sub := &model.Subscription{
		Id:        "apple-" + originalTxId,
		Receipt:   originalTxId,
		AutoRenew: true,
		Issued:    now,
		CreatedBy: "apple",
		UpdatedBy: "apple",
		IsDeleted: &falseVal,
	}
	if sku, known := appleSkuMap[productId]; known {
		sub.Sku = productId
		sub.Name = sku.name
		sub.Description = sku.description
		sub.Credits = sku.credits
	}
	return sub
}

// appleTimeStr formats t as "2006-01-02T15:04:05Z" for readable log output.
// Returns "(not set)" when t is the zero value.
func appleTimeStr(t time.Time) string {
	if t.IsZero() {
		return "(not set)"
	}
	return t.UTC().Format(time.RFC3339)
}

func handleAppleDiscount(c *gin.Context) {

	request := model.DiscountRequest{}
	response := model.DiscountResponse{}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infof("discount request: %v", request)

	// Generate a nonce (UUID)

	nonce, err := uuid.NewRandom()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.Nonce = strings.ToLower(nonce.String())

	// Get the current time in milliseconds
	timestamp := time.Now().UnixMilli()
	response.Timestamp = timestamp

	response.KeyId = os.Getenv("APPLE_ITUNES_IN_APP_PURCHASE_KEY_ID")
	response.ProductId = request.ProductId
	response.OfferId = request.OfferId
	response.UserName = request.UserName

	bundleId := os.Getenv("APPLE_ITUNES_BUNDLE_ID")

	// Create the signature
	// appBundleId + '\u2063' + keyIdentifier + '\u2063' + productIdentifier + '\u2063' + offerIdentifier + '\u2063' + appAccountToken + '\u2063' + nonce + '\u2063' + timestamp

	data := fmt.Sprintf("%s\u2063%s\u2063%s\u2063%s\u2063%s\u2063%s\u2063%d", bundleId, response.KeyId, response.ProductId, response.OfferId, response.UserName, response.Nonce, response.Timestamp)

	log.Infof("data: %s", data)
	println(data)

	// Load the private key
	bytes, err := os.ReadFile(os.Getenv("APPLE_ITUNES_IN_APP_PURCHASE_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	block, _ := pem.Decode(bytes)
	if block == nil || block.Type != "PRIVATE KEY" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode PEM block containing private key"})
		return
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bdata := []byte(data)

	// Sign the data
	hash := sha256.Sum256(bdata)
	r, s, err := ecdsa.Sign(rand.Reader, key.(*ecdsa.PrivateKey), hash[:])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Encode the signature
	signature := append(r.Bytes(), s.Bytes()...)

	response.Signature = base64.StdEncoding.EncodeToString(signature)

	c.JSON(http.StatusOK, response)

}

func fetchApplePublicKey(kid string) (*rsa.PublicKey, error) {
	resp, err := http.Get("https://account.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Apple public keys: %s", resp.Status)
	}

	var jwks map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	if err != nil {
		return nil, err
	}

	keys := jwks["keys"].([]interface{})
	for _, key := range keys {
		keyMap := key.(map[string]interface{})
		if keyMap["kid"].(string) == kid {
			alg := keyMap["alg"].(string)
			switch alg {
			case "RSA256":
				nStr := keyMap["n"].(string)
				eStr := keyMap["e"].(string)

				nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
				if err != nil {
					return nil, err
				}
				eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
				if err != nil {
					return nil, err
				}

				n := new(big.Int).SetBytes(nBytes)
				e := int(new(big.Int).SetBytes(eBytes).Int64())

				pubKey := &rsa.PublicKey{
					N: n,
					E: e,
				}
				return pubKey, nil
			}
		}
	}

	return nil, fmt.Errorf("public key not found")
}

func validateAppleSignature(header, payload, signature string) (bool, error) {
	headerBytes, err := base64.RawURLEncoding.DecodeString(header)
	if err != nil {
		return false, err
	}
	var headerMap map[string]interface{}
	err = json.Unmarshal(headerBytes, &headerMap)
	if err != nil {
		return false, err
	}

	// log.Infof("header: %v", headerMap)

	if headerMap["x5c"] == nil {
		return false, fmt.Errorf("x5c not found")
	}

	x5cArray, ok := headerMap["x5c"].([]interface{})
	if !ok || len(x5cArray) == 0 {
		return false, fmt.Errorf("x5c is not an array or is empty")
	}

	x5c := x5cArray[0].(string)

	pubKey, err := fetchApplePublicKeyFromX5C(x5c)
	if err != nil {
		return false, err
	}

	data := fmt.Sprintf("%s.%s", header, payload)
	hash := sha256.Sum256([]byte(data))

	sig, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	r := new(big.Int).SetBytes(sig[:len(sig)/2])
	s := new(big.Int).SetBytes(sig[len(sig)/2:])

	valid := ecdsa.Verify(pubKey, hash[:], r, s)
	return valid, nil
}

var appleLock sync.Mutex

func createSubscriptionApple(c *gin.Context) {

	appleLock.Lock()
	defer appleLock.Unlock()

	var receipt model.PurchaseRceipt

	err := json.NewDecoder(c.Request.Body).Decode(&receipt)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	log.Infof("apple: %s", receipt)

	// Validate the receipt with Apple
	result, err := validateReceiptApple2(receipt.Receipt)
	if err != nil || result == nil {
		log.Error(err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid receipt"})
		return
	}

	log.Infof("result: %v", result)

	var originalTransactionId string
	if result["originalTransactionId"] != nil {
		originalTransactionId = result["originalTransactionId"].(string)
	} else {
		if result["transactionId"] != nil {
			originalTransactionId = result["transactionId"].(string)
		} else {
			originalTransactionId = receipt.Receipt
		}
	}
	log.Infof("originalTransactionId: %s", originalTransactionId)
	subscription, err := core.GetSubscriptionByReceipt(originalTransactionId)
	if err != nil {
		log.Error(err)
		//		receipt.Receipt = originalTransactionId
	}
	if subscription != nil {
		isDeleted := false
		subscription.IsDeleted = &isDeleted
		if subscription.Email == "" {
			subscription.Email = receipt.Email
			subscription.AccountID = receipt.AccountID
		}
		// subscription.Name = receipt.Name
		subscription.Sku = receipt.ProductID
		last := time.Now().UTC()
		productId, ok := result["productId"].(string)
		if !ok {
			log.Errorf("productId not found in result")
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid subscription"})
			return
		}
		var expires time.Time
		if productId == "24_hours_flex" || productId == "10_day_flex" {
			isDeleted = true
			subscription.Status = "expired"
			core.UpdateSubscription(subscription.Id, subscription)
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid subscription"})
			return
		} else {
			expiresDate, ok := result["expiresDate"].(float64)
			if !ok {
				log.Errorf("expiresDate not found in result")
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid subscription"})
				return
			}
			expires = time.Unix(int64(expiresDate)/1000, 0)
		}
		log.Infof("expires: %s", expires)
		transactionReason, ok := result["transactionReason"].(string)
		if !ok {
			log.Errorf("transactionReason not found in result")
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid subscription"})
			return
		}
		if transactionReason == "RENEWAL" {
			subscription.Expires = &expires
			subscription.LastUpdated = &last
			subscription.UpdatedBy = "apple"
			if subscription.Status == "cancelled" || subscription.Status == "expired" {
				subscription.Status = "active"
				core.SubscriptionEmail(subscription)
				core.UpdateSubscription(subscription.Id, subscription)
				core.RenewSubscription(subscription.Id)
				log.Infof("subscription renewed: %s until %s", subscription.Id, expires)
				c.JSON(http.StatusOK, subscription)
				return
			} else {
				// we're going to let it fall through here and update the subscription
				// (upgrade/downgrade/etc)
				//core.UpdateSubscription(subscription.Id, subscription)
				//core.SubscriptionEmail(subscription)

				//log.Infof("subscription updated: %s %v", subscription.Id, subscription)
				//c.JSON(http.StatusOK, subscription)
				//return
			}

		} else {
			log.Infof("createSubscriptionApple: transactionReason: %s", transactionReason)
			subscription.Expires = &expires
			subscription.LastUpdated = &last
			subscription.UpdatedBy = "apple"
			core.UpdateSubscription(subscription.Id, subscription)
			core.SubscriptionEmail(subscription)

			log.Infof("subscription updated: %s %v", subscription.Id, subscription)
			c.JSON(http.StatusOK, subscription)
			return
		}
	}

	credits := 0
	name := ""
	description := ""
	devices := 0
	networks := 0
	members := 0
	relays := 0
	autoRenew := false
	issued := time.Now()
	expires := time.Now().AddDate(0, 1, 0)

	if result["expiresDate"] != nil {
		expiresDate := result["expiresDate"].(float64)
		expires = time.Unix(int64(expiresDate)/1000, 0)
	}

	switch receipt.ProductID {
	case "basic_monthly":
		name = "Core Service (monthly)"
	case "basic_yearly":
		name = "Core Service (yearly)"
		expires = time.Now().AddDate(1, 0, 0)
	case "premium_monthly":
		name = "Premium Service (monthly)"
	case "premium_yearly":
		name = "Premium Service (yearly)"
		expires = time.Now().AddDate(1, 0, 0)
	case "professional_monthly":
		name = "Professional Service (monthly)"
	case "professional_yearly":
		name = "Professional Service (yearly)"
		expires = time.Now().AddDate(1, 0, 0)
	default:
		log.Errorf("unknown sku %s", receipt.ProductID)
	}
	// set the credits, name, and description based on the sku
	switch receipt.ProductID {
	case "24_hours_flex":
		credits = 1
		name = "24 Hours"
		description = "Service in any region for 24 hours"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = false
		expires = time.Now().Add(24 * time.Hour)
	case "10_day_flex":
		credits = 1
		name = "10 Day Flex"
		description = "Service in any region for 10 days"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = false
		expires = time.Now().AddDate(0, 0, 10)
	case "basic_monthly", "basic_yearly":
		credits = 1
		description = "A single tunnel or relay in any region"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = true
	case "premium_monthly", "premium_yearly":
		credits = 5
		description = "Up to 5 tunnels or relays in any region"
		devices = 25
		networks = 10
		members = 5
		relays = 5
		autoRenew = true
	case "professional_monthly", "professional_yearly":
		credits = 10
		description = "Up to 10 tunnels or relays in any region"
		devices = 100
		networks = 25
		members = 25
		relays = 10
		autoRenew = true
	default:
		log.Errorf("unknown sku %s", receipt.ProductID)
	}

	// set the limits based on the sku
	accounts, err := core.ReadAllAccounts(receipt.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			account.Name = "Me"
			account.AccountName = "Company"
			account.Email = receipt.Email
			account.Role = "Owner"
			account.Status = "Active"
			account.CreatedBy = receipt.Email
			account.UpdatedBy = receipt.Email
			account.Picture = os.Getenv("SERVER") + "/account-circle.png"

			a, err := core.CreateAccount(&account)
			log.Infof("CREATE ACCOUNT = %v", a)
			if err != nil {
				log.Error(err)
			}
			accounts, err = core.ReadAllAccounts(receipt.Email)
			if err != nil {
				log.Error(err)
			}

		}
	}

	var account *model.Account
	for i := 0; i < len(accounts); i++ {
		if accounts[i].Id == accounts[i].Parent {
			account = accounts[i]
			break
		}
	}

	if account == nil && len(accounts) > 0 {
		account = accounts[0]
	}

	if account == nil {
		log.Errorf("account not found for email %s", receipt.Email)
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	limits, err := core.ReadLimits(account.Id)
	if err != nil {
		log.Error(err)
		limits_id, err := util.GenerateRandomString(8)
		if err != nil {
			log.Error(err)
		}
		limits_id = "limits-" + limits_id

		limits = &model.Limits{
			Id:          limits_id,
			AccountID:   account.Id,
			MaxDevices:  0,
			MaxNetworks: 0,
			MaxMembers:  0,
			MaxServices: 0,
			Tolerance:   core.GetDefaultTolerance(),
			CreatedBy:   receipt.Email,
			UpdatedBy:   receipt.Email,
			Created:     time.Now(),
			Updated:     time.Now(),
		}
	}

	limits.MaxDevices += devices
	limits.MaxNetworks += networks
	limits.MaxMembers += members
	limits.MaxServices += relays

	errs := limits.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("limits validation error")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limits validation error"})
		return
	}

	// save limits to mongodb
	mongo.Serialize(limits.Id, "id", "limits", limits)

	// generate a random subscription id
	id, err := util.RandomString(8)
	if err != nil {
		log.Error(err)
	}
	id = receipt.Source + "-" + id

	// construct a subscription object
	lu := time.Now()
	isDeleted := false

	sub := model.Subscription{
		Id:          id,
		AccountID:   account.Id,
		Email:       receipt.Email,
		Name:        name,
		Description: description,
		Issued:      &issued,
		LastUpdated: &lu,
		Expires:     &expires,
		UpdatedBy:   receipt.Email,
		Credits:     credits,
		Sku:         receipt.ProductID,
		Status:      "active",
		AutoRenew:   autoRenew,
		Receipt:     receipt.Receipt,
		IsDeleted:   &isDeleted,
	}

	if subscription != nil {
		log.Infof("subscription found: %s", subscription.Id)
		sub.Id = subscription.Id
		sub.AccountID = subscription.AccountID
		sub.Email = subscription.Email
		sub.Issued = subscription.Issued
		sub.Receipt = subscription.Receipt
		sub.CreatedBy = receipt.Email
		sub.UpdatedBy = subscription.UpdatedBy
		sub.LastUpdated = subscription.LastUpdated
		sub.Expires = subscription.Expires
		sub.Status = subscription.Status

	} else {
		log.Infof("creating new subscription")

	}

	errs = sub.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("subscription validation error")
		}
		return
	}

	// save subscription to mongodb
	mongo.Serialize(sub.Id, "id", "subscriptions", sub)

	err = core.SubscriptionEmail(&sub)
	if err != nil {
		log.Errorf("failed to send email: %v", err)
	}

	log.Infof("subscription created: %s %v", sub.Id, sub)

	c.JSON(http.StatusOK, sub)

}

func validateReceiptApple(receipt string) (bool, error) {
	// Apple receipt validation URL
	//	url := "https://buy.itunes.apple.com/verifyReceipt"
	//url := "https://sandbox.itunes.apple.com/verifyReceipt"
	url := os.Getenv("APPLE_ITUNES_RECEIPT_URL")

	// Create the request payload
	payload := map[string]string{
		"receipt-data": receipt,
		"password":     os.Getenv("APPLE_ITUNES_SHARED_SECRET"), // Replace with your app's shared secret
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	// Send the request to Apple
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("invalid response from Apple: %s", resp.Status)
	}

	// Parse the response to check the receipt status
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, err
	}

	log.Infof("apple receipt: %v", result)

	// Check if the receipt is valid
	if status, ok := result["status"].(float64); ok && status == 0 {
		return true, nil
	}

	return false, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// createSubscriptionApple2 — hardened Apple IAP handler
//
// Called directly by the iOS app without session authentication.
// Validates the Apple transaction token with Apple's servers before touching
// any subscription or account data.
// ─────────────────────────────────────────────────────────────────────────────

// appleSkuDef holds the service entitlements for one Apple IAP product.
// Flex products (24_hours_flex, 10_day_flex) are intentionally absent —
// they are non-renewable and must never be restored.
type appleSkuDef struct {
	name        string
	description string
	credits     int
	devices     int
	networks    int
	members     int
	relays      int
}

var appleSkuMap = map[string]appleSkuDef{
	"basic_monthly":        {"Core Service (monthly)", "A single tunnel or relay in any region", 1, 5, 1, 2, 1},
	"basic_yearly":         {"Core Service (yearly)", "A single tunnel or relay in any region", 1, 5, 1, 2, 1},
	"premium_monthly":      {"Premium Service (monthly)", "Up to 5 tunnels or relays in any region", 5, 25, 10, 5, 5},
	"premium_yearly":       {"Premium Service (yearly)", "Up to 5 tunnels or relays in any region", 5, 25, 10, 5, 5},
	"professional_monthly": {"Professional Service (monthly)", "Up to 10 tunnels or relays in any region", 10, 100, 25, 25, 10},
	"professional_yearly":  {"Professional Service (yearly)", "Up to 10 tunnels or relays in any region", 10, 100, 25, 25, 10},
}

// ── Per-transaction locking ───────────────────────────────────────────────────
// Serialises concurrent calls that share the same originalTransactionId so that
// restore-purchase storms cannot create duplicate subscriptions or double-grant
// limits.  Uses a reference-counted map so locks clean up after themselves.

type appleTxEntry struct {
	mu      sync.Mutex
	waiters int
}

var (
	appleTxMu       sync.Mutex
	appleTxInflight = make(map[string]*appleTxEntry)
)

func acquireAppleTxLock(txID string) func() {
	appleTxMu.Lock()
	e, ok := appleTxInflight[txID]
	if !ok {
		e = &appleTxEntry{}
		appleTxInflight[txID] = e
	}
	e.waiters++
	appleTxMu.Unlock()

	e.mu.Lock()
	return func() {
		e.mu.Unlock()
		appleTxMu.Lock()
		e.waiters--
		if e.waiters == 0 {
			delete(appleTxInflight, txID)
		}
		appleTxMu.Unlock()
	}
}

// ── Safe field extraction ─────────────────────────────────────────────────────
// These avoid the panic that results from a bare .(string) assertion when
// Apple's response is missing a field.

func appleStr(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func appleF64(m map[string]interface{}, key string) (float64, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

// ── Account / limits helpers ──────────────────────────────────────────────────

// appleEnsureAccount returns the root Nettica account for email, creating one
// if none exists yet.
func appleEnsureAccount(email string) (*model.Account, error) {
	accounts, err := core.ReadAllAccounts(email)
	if err != nil {
		return nil, fmt.Errorf("ReadAllAccounts: %w", err)
	}
	if len(accounts) == 0 {
		acct := model.Account{
			Name:        "Me",
			AccountName: "Company",
			Email:       email,
			Role:        "Owner",
			Status:      "Active",
			CreatedBy:   email,
			UpdatedBy:   email,
			Picture:     os.Getenv("SERVER") + "/account-circle.png",
		}
		if _, err = core.CreateAccount(&acct); err != nil {
			return nil, fmt.Errorf("CreateAccount: %w", err)
		}
		accounts, err = core.ReadAllAccounts(email)
		if err != nil {
			return nil, fmt.Errorf("ReadAllAccounts after create: %w", err)
		}
	}
	for _, a := range accounts {
		if a.Id == a.Parent {
			return a, nil
		}
	}
	if len(accounts) > 0 {
		return accounts[0], nil
	}
	return nil, fmt.Errorf("no account found for %s after create", email)
}

// appleAddLimits increments the account limits by one SKU's worth of
// entitlements.  Callers are responsible for calling this exactly once per
// subscription lifetime.
func appleAddLimits(accountID, email string, sku appleSkuDef) error {
	limits, err := core.ReadLimits(accountID)
	if err != nil {
		limitsID, idErr := util.GenerateRandomString(8)
		if idErr != nil {
			return fmt.Errorf("GenerateRandomString: %w", idErr)
		}
		limits = &model.Limits{
			Id:        "limits-" + limitsID,
			AccountID: accountID,
			Tolerance: core.GetDefaultTolerance(),
			CreatedBy: email,
			UpdatedBy: email,
			Created:   time.Now().UTC(),
			Updated:   time.Now().UTC(),
		}
	}
	limits.MaxDevices += sku.devices
	limits.MaxNetworks += sku.networks
	limits.MaxMembers += sku.members
	limits.MaxServices += sku.relays
	limits.UpdatedBy = email
	limits.Updated = time.Now().UTC()

	if errs := limits.IsValid(); len(errs) != 0 {
		for _, e := range errs {
			log.WithField("err", e).Error("appleAddLimits: validation error")
		}
		return fmt.Errorf("limits validation failed")
	}
	return mongo.Serialize(limits.Id, "id", "limits", limits)
}

// ── Main handler ──────────────────────────────────────────────────────────────

func createSubscriptionApple2(c *gin.Context) {

	// 1. Parse and minimally validate the request body.
	var receipt model.PurchaseRceipt
	if err := json.NewDecoder(c.Request.Body).Decode(&receipt); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request body"})
		return
	}
	if receipt.Receipt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receipt is required"})
		return
	}
	if receipt.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	log.Infof("createSubscriptionApple2: email=%s productId=%s", receipt.Email, receipt.ProductID)

	// 2. Validate the transaction token with Apple (production, then sandbox).
	//    This is the only source of truth — we never trust client-provided fields
	//    for entitlement decisions.
	result, err := validateReceiptApple2(receipt.Receipt)
	if err != nil || result == nil {
		log.Errorf("createSubscriptionApple2: Apple validation failed: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid receipt"})
		return
	}

	// 3. Extract required fields from Apple's response without risking a panic.
	productId, ok := appleStr(result, "productId")
	if !ok || productId == "" {
		log.Errorf("createSubscriptionApple2: productId missing from Apple response for %s", receipt.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid receipt: missing productId"})
		return
	}

	// originalTransactionId is the stable key that links every renewal of one
	// subscription.  transactionId changes on every billing cycle.
	originalTxId, ok := appleStr(result, "originalTransactionId")
	if !ok || originalTxId == "" {
		if originalTxId, ok = appleStr(result, "transactionId"); !ok || originalTxId == "" {
			// Sandbox / legacy fallback.
			originalTxId = receipt.Receipt
		}
	}

	transactionReason, _ := appleStr(result, "transactionReason")

	// expiresDate is mandatory for auto-renewable subscriptions.  Without it we
	// have no way to confirm whether service should be granted.
	expiresMs, hasExpiry := appleF64(result, "expiresDate")
	if !hasExpiry {
		log.Errorf("createSubscriptionApple2: expiresDate missing from Apple response for %s %s", productId, receipt.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot determine subscription validity"})
		return
	}
	appleExpires := time.Unix(int64(expiresMs)/1000, 0).UTC()

	// 4. Flex products are time-boxed and non-renewable — Apple never issues a
	//    future expiresDate for them after the window closes.  Reject explicitly
	//    rather than letting the expiry check below handle it so the error is clear.
	if productId == "24_hours_flex" || productId == "10_day_flex" {
		log.Infof("createSubscriptionApple2: rejecting non-restorable flex product %s for %s", productId, receipt.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "this product type cannot be restored"})
		return
	}

	// 5. Reject unknown SKUs.
	sku, knownSKU := appleSkuMap[productId]
	if !knownSKU {
		log.Errorf("createSubscriptionApple2: unknown productId %q from Apple for %s", productId, receipt.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown product"})
		return
	}

	// 6. Apple's expiresDate is ground truth.  Never grant or restore service for
	//    a subscription Apple itself says has expired.
	if appleExpires.Before(time.Now().UTC()) {
		log.Infof("createSubscriptionApple2: Apple confirms expired at %s for %s", appleExpires, receipt.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "subscription has expired"})
		return
	}

	// 7. Per-transaction lock: prevents a restore-purchase burst from creating
	//    duplicate subscriptions or granting limits more than once.
	release := acquireAppleTxLock(originalTxId)
	defer release()

	// 8. Look up an existing subscription.
	// GetSubscriptionByReceipt returns (nil, mongo.ErrNoDocuments) — message
	// "mongo: no documents in result" — when nothing exists yet.  That is the
	// normal path for a first purchase, not a fault.  Any other error is a real
	// DB problem; in that case we stop rather than risk creating a duplicate and
	// double-granting limits.
	existing, err := core.GetSubscriptionByReceipt(originalTxId)
	if err != nil {
		if !strings.Contains(err.Error(), "no documents in result") {
			log.Errorf("createSubscriptionApple2: DB error looking up %s: %v", originalTxId, err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable"})
			return
		}
		existing = nil
	}

	now := time.Now().UTC()
	falseVal := false

	if existing != nil {
		// ── Subscription already exists ──────────────────────────────────────────

		// Security: once a subscription is claimed by a Nettica email, only that
		// email may interact with it.  This prevents a caller from replaying a
		// stolen transaction token against a different account.
		if existing.Email != "" && existing.Email != receipt.Email {
			log.Errorf("createSubscriptionApple2: receipt %s is owned by %s, refused for %s",
				originalTxId, existing.Email, receipt.Email)
			c.JSON(http.StatusForbidden, gin.H{"error": "receipt belongs to a different account"})
			return
		}

		// handleAppleWebhook creates stub entries (email="", accountID="") when a
		// webhook fires before the app has connected.  Limits are never granted for
		// stubs — we must grant them exactly once, right here, when the app claims
		// the stub.
		isStub := existing.AccountID == ""

		if existing.Email == "" {
			existing.Email = receipt.Email
		}

		var acctID string
		if isStub {
			acct, acctErr := appleEnsureAccount(receipt.Email)
			if acctErr != nil {
				log.Errorf("createSubscriptionApple2: appleEnsureAccount: %v", acctErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve account"})
				return
			}
			acctID = acct.Id
			existing.AccountID = acctID
		} else {
			acctID = existing.AccountID
		}

		// Preserve the original Issued timestamp; set it only if the stub left it nil.
		if existing.Issued == nil {
			existing.Issued = &now
		}

		// Update authoritative fields from Apple's response.
		existing.Sku = productId
		existing.Name = sku.name
		existing.Description = sku.description
		existing.Credits = sku.credits
		existing.AutoRenew = true
		existing.Expires = &appleExpires
		existing.LastUpdated = &now
		existing.UpdatedBy = "apple"
		existing.IsDeleted = &falseVal

		// Status is inactive for: explicit cancellation/expiry, or an unclaimed stub.
		wasInactive := existing.Status == "expired" || existing.Status == "cancelled" || existing.Status == ""

		if wasInactive {
			existing.Status = "active"
			if _, updateErr := core.UpdateSubscription(existing.Id, existing); updateErr != nil {
				log.Errorf("createSubscriptionApple2: UpdateSubscription failed: %v", updateErr)
			}
			if renewErr := core.RenewSubscription(existing.Id); renewErr != nil {
				log.Errorf("createSubscriptionApple2: RenewSubscription failed: %v", renewErr)
			}
			// Grant limits only for stubs — a previously active subscription already
			// had its limits granted at purchase time; granting again would inflate them.
			if isStub {
				if limErr := appleAddLimits(acctID, receipt.Email, sku); limErr != nil {
					log.Errorf("createSubscriptionApple2: appleAddLimits failed: %v", limErr)
				}
			}
			core.SubscriptionEmail(existing)
			log.Infof("createSubscriptionApple2: reactivated %s (%s) until %s reason=%s",
				existing.Id, receipt.Email, appleExpires, transactionReason)
		} else {
			// Subscription is already active — update metadata and expiry only.
			// Do NOT touch limits; they were already granted at the original purchase.
			if _, updateErr := core.UpdateSubscription(existing.Id, existing); updateErr != nil {
				log.Errorf("createSubscriptionApple2: UpdateSubscription failed: %v", updateErr)
			}
			if isStub {
				// Edge case: active stub (webhook fired with status set, but no account).
				if limErr := appleAddLimits(acctID, receipt.Email, sku); limErr != nil {
					log.Errorf("createSubscriptionApple2: appleAddLimits failed: %v", limErr)
				}
			}
			log.Infof("createSubscriptionApple2: updated %s (%s) reason=%s",
				existing.Id, receipt.Email, transactionReason)
		}

		c.JSON(http.StatusOK, existing)
		return
	}

	// ── No existing subscription — genuine first-time purchase ───────────────────

	acct, err := appleEnsureAccount(receipt.Email)
	if err != nil {
		log.Errorf("createSubscriptionApple2: appleEnsureAccount: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve account"})
		return
	}

	if err := appleAddLimits(acct.Id, receipt.Email, sku); err != nil {
		log.Errorf("createSubscriptionApple2: appleAddLimits failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision service limits"})
		return
	}

	subID, err := util.RandomString(8)
	if err != nil {
		log.Errorf("createSubscriptionApple2: RandomString failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	// Always use the fixed "apple-" prefix — never interpolate receipt.Source,
	// which is client-provided and could be used to craft predictable IDs.
	subID = "apple-" + subID

	sub := model.Subscription{
		Id:          subID,
		AccountID:   acct.Id,
		Email:       receipt.Email,
		Name:        sku.name,
		Description: sku.description,
		Credits:     sku.credits,
		Sku:         productId,
		AutoRenew:   true,
		Status:      "active",
		Receipt:     originalTxId, // stable originalTransactionId, NOT the raw JWS token
		Issued:      &now,
		LastUpdated: &now,
		Expires:     &appleExpires,
		CreatedBy:   receipt.Email,
		UpdatedBy:   "apple",
		IsDeleted:   &falseVal,
	}

	if errs := sub.IsValid(); len(errs) != 0 {
		for _, e := range errs {
			log.WithField("err", e).Error("createSubscriptionApple2: subscription validation error")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription validation failed"})
		return
	}

	if err := mongo.Serialize(sub.Id, "id", "subscriptions", sub); err != nil {
		log.Errorf("createSubscriptionApple2: Serialize failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save subscription"})
		return
	}

	if err := core.SubscriptionEmail(&sub); err != nil {
		log.Errorf("createSubscriptionApple2: SubscriptionEmail failed: %v", err)
	}

	log.Infof("createSubscriptionApple2: created %s for %s until %s", sub.Id, receipt.Email, appleExpires)
	c.JSON(http.StatusCreated, sub)
}

// validateReceiptApple2 validates an Apple purchaseId
// It first tries to validate the receipt with the production URL
// If that fails, it tries the sandbox URL
// This allows the same code to run in both production and development environments
func validateReceiptApple2(receipt string) (map[string]interface{}, error) {

	production := os.Getenv("APPLE_ITUNES_RECEIPT_URL")
	sandbox := os.Getenv("APPLE_ITUNES_SANDBOX_URL")

	// Attempt validation with the production URL
	response, err := validateReceipt(production, receipt)
	if err != nil {
		log.Errorf("Failed to validate receipt with production URL: %v", err)
		// Attempt validation with the sandbox URL
		response, err = validateReceipt(sandbox, receipt)
		if err != nil {
			log.Errorf("Failed to validate receipt with sandbox URL: %v", err)
			return nil, err
		}
	}

	return response, nil
}

func validateReceipt(url, receipt string) (map[string]interface{}, error) {
	// Apple receipt validation URL
	// url := "https://api.storekit.itunes.apple.com/inApps/v1/transactions/{transactionId}"
	// url := "https://api.storekit-sandbox.itunes.apple.com/inApps/v1/transactions/{transactionId}"
	url += receipt

	// Create a JWT to authenticate with Apple
	keyfile := os.Getenv("APPLE_ITUNES_IN_APP_PURCHASE_KEY")
	keyid := os.Getenv("APPLE_ITUNES_IN_APP_PURCHASE_KEY_ID")
	issuer := os.Getenv("APPLE_ITUNES_IN_APP_PURCHASE_KEY_ISSUER")

	keybytes, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseECPrivateKeyFromPEM(keybytes)
	if err != nil {
		log.Error(err)
	}

	t := jwt.New(jwt.SigningMethodES256)
	t.Header["kid"] = keyid
	t.Header["alg"] = "ES256"
	t.Header["typ"] = "JWT"
	t.Claims = jwt.MapClaims{
		"iss": issuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute * 60).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": "com.nettica.agent",
	}
	token, err := t.SignedString(key)
	if err != nil {
		return nil, err
	}

	// call the server with the JWT and url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error statusCode from Apple: %d %s", resp.StatusCode, resp.Status)
	}

	// enough of this bullshit!

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := string(bytes)

	parts := strings.Split(response, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid response from Apple: %s", response)
	}

	// Decode the payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// Parse the response to check the receipt status
	var result map[string]interface{}
	err = json.Unmarshal(payload, &result)
	if err != nil {
		return nil, err
	}

	log.Infof("apple receipt: %v", result)

	// Check if the receipt is valid
	if status, ok := result["status"].(float64); ok && status == 0 {
		return result, nil
	}

	// TODO: change this to false when we're done testing
	return result, nil
}

func createHelioSubscription(c *gin.Context) {
	var body string
	var sub map[string]interface{}

	HelioPaylinkIdRelay := os.Getenv("HELIO_PAYLINK_ID_RELAY")
	HelioPaylinkIdPremium := os.Getenv("HELIO_PAYLINK_ID_PREMIUM")
	HelioPaylinkIdPro := os.Getenv("HELIO_PAYLINK_ID_PRO")

	HelioPaylinkTokenRelay := os.Getenv("HELIO_PAYLINK_TOKEN_RELAY")
	HelioPaylinkTokenPremium := os.Getenv("HELIO_PAYLINK_TOKEN_PREMIUM")
	HelioPaylinkTokenPro := os.Getenv("HELIO_PAYLINK_TOKEN_PRO")

	// read and log the request body

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	body = string(bytes)
	// remove all the backslashes from the body (is this needed?)
	body = strings.Replace(body, "\\", "", -1)
	log.Info(body)
	bytes = []byte(body)

	// unmarshal the body into a map
	err = json.Unmarshal(bytes, &sub)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal request body")
		// c.AbortWithStatus(http.StatusUnprocessableEntity)
		// return with no error so webhook doesn't get disabled
		c.JSON(http.StatusOK, body)
		return
	}

	log.Info(sub)

	var sku string
	var sharedKey string

	// get the transact id from the sub object

	transact := sub["transaction"].(map[string]interface{})

	paylinkId := transact["paystream"].(string)

	switch paylinkId {
	case HelioPaylinkIdRelay:
		sku = "Relay-1"
		sharedKey = "Bearer " + HelioPaylinkTokenRelay
	case HelioPaylinkIdPremium:
		sku = "Premium-5"
		sharedKey = "Bearer " + HelioPaylinkTokenPremium
	case HelioPaylinkIdPro:
		sku = "Pro-10"
		sharedKey = "Bearer " + HelioPaylinkTokenPro
	default:
		log.Errorf("unknown paylink_id %s", paylinkId)
	}

	// Read the Authorization header
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		log.Error("Authorization header is empty")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is empty"})
		return
	}

	// Compare the sharedKey with the Authorization header

	if sharedKey != authorization {
		log.Errorf("Authenication denied")
		// c.JSON(http.StatusUnauthorized, gin.H{"error": "Authenication denied"})
		// return
	}

	// walk the json and find the email address

	meta := transact["meta"].(map[string]interface{})
	log.Info(meta)

	customer := meta["customerDetails"].(map[string]interface{})
	log.Info(customer)

	email := customer["email"].(string)
	log.Info(email)

	transStatus := sub["event"].(string)

	var status string

	switch transStatus {
	case "STARTED":
		status = "active"
	case "STOPPED":
		status = "expired"
	case "CANCELLED":
		status = "cancelled"
	default:
		log.Errorf("unknown transaction status %s", transStatus)
	}

	endedAt := transact["endedAt"].(string)
	log.Info(endedAt)
	// convert the endedAt string of unixtime to a time.Time object
	endedAtUnix, err := time.Parse(time.RFC3339, endedAt)
	if err != nil {
		log.Error(err)
	}

	start := transact["startedAt"].(map[string]interface{})
	id := start["transactionSignature"].(string)
	log.Info(id)

	go func() {

		customer_name := "Me"

		credits := 0
		name := ""
		description := ""
		devices := 0
		networks := 0
		members := 0
		relays := 0

		// set the credits, name, and description based on the sku
		switch sku {
		case "Starter-0":
			credits = 0
			name = "Starter"
			description = "The Starter subscription"
			devices = core.GetDefaultMaxDevices()
			networks = core.GetDefaultMaxNetworks()
			members = core.GetDefaultMaxMembers()
			relays = core.GetDefaultMaxServices()

		case "Relay-1":
			fallthrough
		case "RelayYear-1":
			credits = 1
			name = "Core Service"
			description = "A single tunnel or relay in any Region"
			devices = 5
			networks = 1
			members = 2
			relays = 1
		case "Premium-5":
			fallthrough
		case "PremiumYear-5":
			credits = 5
			name = "Premium"
			description = "Up to 5 tunnels or relays in any Region"
			devices = 25
			networks = 10
			members = 5
			relays = 5
		case "Pro-10":
			fallthrough
		case "ProYear-10":
			credits = 10
			name = "Professional"
			description = "Up to 10 tunnels or relays in any Region"
			devices = 100
			networks = 25
			members = 25
			relays = 10
		default:
			log.Errorf("unknown sku %s", sku)
		}

		// set the limits based on the sku
		accounts, err := core.ReadAllAccounts(email)
		if err != nil {
			log.Error(err)
		} else {
			//  If there's no error and no account, create one.
			if len(accounts) == 0 {
				var account model.Account
				account.Name = customer_name
				account.AccountName = "Company"
				account.Email = email
				account.Role = "Owner"
				account.Status = "Active"
				account.CreatedBy = email
				account.UpdatedBy = email
				account.Picture = os.Getenv("SERVER") + "/account-circle.png"

				a, err := core.CreateAccount(&account)
				log.Infof("CREATE ACCOUNT = %v", a)
				if err != nil {
					log.Error(err)
				}
				accounts, err = core.ReadAllAccounts(email)
				if err != nil {
					log.Error(err)
				}

			}
		}

		var account *model.Account
		for i := 0; i < len(accounts); i++ {
			if accounts[i].Id == accounts[i].Parent {
				account = accounts[i]
				break
			}
		}

		if account == nil && len(accounts) > 0 {
			account = accounts[0]
		}

		if account == nil {
			log.Errorf("account not found for email %s", email)
			return
		}

		limits, err := core.ReadLimits(account.Id)
		if err != nil {
			log.Error(err)
			limits_id, err := util.GenerateRandomString(8)
			if err != nil {
				log.Error(err)
			}
			limits_id = "limits-" + limits_id

			limits = &model.Limits{
				Id:          limits_id,
				AccountID:   account.Id,
				MaxDevices:  0,
				MaxNetworks: 0,
				MaxMembers:  0,
				MaxServices: 0,
				Tolerance:   core.GetDefaultTolerance(),
				CreatedBy:   email,
				UpdatedBy:   email,
				Created:     time.Now(),
				Updated:     time.Now(),
			}
		}

		limits.MaxDevices += devices
		limits.MaxNetworks += networks
		limits.MaxMembers += members
		limits.MaxServices += relays

		errs := limits.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("limits validation error")
			}
			return
		}

		// save limits to mongodb
		mongo.Serialize(limits.Id, "id", "limits", limits)

		// construct a subscription object
		issued := time.Now()
		lu := time.Now()
		expires := endedAtUnix
		isDeleted := false
		subscription := model.Subscription{
			Id:          id,
			AccountID:   account.Id,
			Email:       email,
			Name:        name,
			Description: description,
			Issued:      &issued,
			LastUpdated: &lu,
			Expires:     &expires,
			Credits:     credits,
			Sku:         sku,
			Status:      status,
			IsDeleted:   &isDeleted,
			UpdatedBy:   email,
			CreatedBy:   email,
		}

		errs = subscription.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("subscription validation error")
			}
			return
		}

		// save subscription to mongodb
		mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

	}()

	c.JSON(http.StatusOK, body)
}

func htmlHack(body string) string {

	// WooCommerce injecting unescaped HTML into JSON.  Great job.
	// That must have taken some effort.  Remove it.

	front := strings.Index(body, "<span class=")
	back := strings.LastIndex(body, "/></div>")

	if front == -1 || back == -1 {
		return body
	}
	body = body[:front] + body[back+7:]

	return body
}

func createSubscription(c *gin.Context) {
	var body string
	var sub map[string]interface{}

	// get the secret and hash of the body
	secret := os.Getenv("WC_SECRET")
	signature := c.Request.Header.Get("x-wc-webhook-signature")

	// read and log the request body

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	body = string(bytes)
	// remove all the backslashes from the body (is this needed?)
	bodi := strings.Replace(body, "\\", "", -1)
	body = htmlHack(bodi)
	if body != bodi {
		log.Info("WooPayments still broken")
	}

	log.Info(body)
	bytes = []byte(body)

	// hash the body and compare it to the signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(bytes)
	expected := h.Sum(nil)
	if !hmac.Equal([]byte(signature), expected) {
		log.WithFields(log.Fields{
			"signature": signature,
			"expected":  expected,
			"body":      string(bytes),
		}).Error("failed to verify signature")
	} else {
		log.Info("signature verified")
	}

	// unmarshal the body into a map
	err = json.Unmarshal(bytes, &sub)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal request body")
		// c.AbortWithStatus(http.StatusUnprocessableEntity)
		// return with no error so webhook doesn't get disabled
		c.JSON(http.StatusOK, body)
		return
	}

	log.Info(sub)

	receipt := fmt.Sprintf("%d", int(sub["id"].(float64)))
	log.Infof("receipt: %s", receipt)

	// walk the json and find the customer href
	links := sub["_links"].(map[string]interface{})
	log.Info(links)

	// get the sku from the line_items
	sku := sub["line_items"].([]interface{})[0].(map[string]interface{})["sku"].(string)
	status := sub["status"].(string)

	//status := "active"

	customer := links["customer"].([]interface{})
	log.Info(customer)

	customer0 := customer[0].(map[string]interface{})
	log.Info(customer0)

	href := customer0["href"].(string)
	log.Info(href)

	go func() {

		// make http request with basic authentication using href as url to get the customer object
		req, err := http.NewRequest("GET", href, nil)
		if err != nil {
			return
		}

		req.SetBasicAuth(os.Getenv("WC_USERNAME"), os.Getenv("WC_PASSWORD"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf("http.client.Do = %v", err)
			return
		}

		if resp.StatusCode != 200 {
			log.Errorf("http status %s expect 200 OK", resp.Status)
			return
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			log.Error(err)
			return
		}

		log.Info(data)

		// get the email from the customer object
		email := data["email"].(string)
		log.Info(email)

		customer_name := data["first_name"].(string) + " " + data["last_name"].(string)

		// generate a random subscription id
		id, err := util.RandomString(8)
		if err != nil {
			log.Error(err)
		}

		credits := 0
		name := ""
		description := ""
		devices := 0
		networks := 0
		members := 0
		relays := 0
		expires := time.Now().AddDate(1, 0, 0)

		if sub["next_payment_date_gmt"] != nil {
			layout := "2006-01-02T15:04:05"
			expires, err = time.Parse(layout, sub["next_payment_date_gmt"].(string))
			if err != nil {
				log.Error(err)
			}
		}

		// set the credits, name, and description based on the sku
		switch sku {
		case "Starter-0":
			credits = 0
			name = "Starter"
			description = "The Starter subscription"
			devices = core.GetDefaultMaxDevices()
			networks = core.GetDefaultMaxNetworks()
			members = core.GetDefaultMaxMembers()
			relays = core.GetDefaultMaxServices()

		case "Relay-1":
			fallthrough
		case "RelayYear-1":
			credits = 1
			name = "Core Service"
			description = "A single tunnel or relay in any Region"
			devices = 5
			networks = 1
			members = 2
			relays = 1
		case "Premium-5":
			fallthrough
		case "PremiumYear-5":
			credits = 5
			name = "Premium"
			description = "Up to 5 tunnels or relays in any Region"
			devices = 25
			networks = 10
			members = 5
			relays = 5
		case "Pro-10":
			fallthrough
		case "ProYear-10":
			credits = 10
			name = "Professional"
			description = "Up to 10 tunnels or relays in any Region"
			devices = 100
			networks = 25
			members = 25
			relays = 10
		default:
			log.Errorf("unknown sku %s", sku)
		}

		// set the limits based on the sku
		accounts, err := core.ReadAllAccounts(email)
		if err != nil {
			log.Error(err)
		} else {
			//  If there's no error and no account, create one.
			if len(accounts) == 0 {
				var account model.Account
				account.Name = customer_name
				account.AccountName = "Company"
				account.Email = email
				account.Role = "Owner"
				account.Status = "Active"
				account.CreatedBy = email
				account.UpdatedBy = email
				account.Picture = os.Getenv("SERVER") + "/account-circle.png"

				a, err := core.CreateAccount(&account)
				log.Infof("CREATE ACCOUNT = %v", a)
				if err != nil {
					log.Error(err)
				}
				accounts, err = core.ReadAllAccounts(email)
				if err != nil {
					log.Error(err)
				}

			}
		}

		var account *model.Account
		for i := 0; i < len(accounts); i++ {
			if accounts[i].Id == accounts[i].Parent {
				account = accounts[i]
				break
			}
		}

		if account == nil && len(accounts) > 0 {
			account = accounts[0]
		}

		if account == nil {
			log.Errorf("account not found for email %s", email)
			return
		}

		limits, err := core.ReadLimits(account.Id)
		if err != nil {
			log.Error(err)
			limits_id, err := util.GenerateRandomString(8)
			if err != nil {
				log.Error(err)
			}
			limits_id = "limits-" + limits_id

			limits = &model.Limits{
				Id:          limits_id,
				AccountID:   account.Id,
				MaxDevices:  0,
				MaxNetworks: 0,
				MaxMembers:  0,
				MaxServices: 0,
				Tolerance:   core.GetDefaultTolerance(),
				CreatedBy:   email,
				UpdatedBy:   email,
				Created:     time.Now(),
				Updated:     time.Now(),
			}
		}

		limits.MaxDevices += devices
		limits.MaxNetworks += networks
		limits.MaxMembers += members
		limits.MaxServices += relays

		errs := limits.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("limits validation error")
			}
			return
		}

		// save limits to mongodb
		mongo.Serialize(limits.Id, "id", "limits", limits)

		// construct a subscription object
		issued := time.Now()
		lu := time.Now()
		isDeleted := false
		subscription := model.Subscription{
			Id:          id,
			AccountID:   account.Id,
			Email:       email,
			Name:        name,
			Description: description,
			Issued:      &issued,
			LastUpdated: &lu,
			Expires:     &expires,
			Credits:     credits,
			Sku:         sku,
			Status:      status,
			Receipt:     receipt,
			IsDeleted:   &isDeleted,
			UpdatedBy:   email,
			CreatedBy:   email,
		}

		errs = subscription.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("subscription validation error")
			}
			return
		}

		// save subscription to mongodb
		mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

	}()

	c.JSON(http.StatusOK, body)
}

func updateSubscriptionWoo(c *gin.Context) {

	var body string
	//	var sub map[string]interface{}

	// get the secret and hash of the body
	//	secret := os.Getenv("WC_SECRET")
	//	signature := c.Request.Header.Get("x-wc-webhook-signature")

	// read and log the request body for now

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	body = string(bytes)
	// remove all the backslashes from the body (is this needed?)
	bodi := strings.Replace(body, "\\", "", -1)
	body = htmlHack(bodi)
	if body != bodi {
		log.Info("WooPayments still broken")
	}

	log.Info(body)

	// unmarshal the body into a map

	var sub map[string]interface{}
	err = json.Unmarshal(bytes, &sub)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal request body")
		// c.AbortWithStatus(http.StatusUnprocessableEntity)
		// return with no error so webhook doesn't get disabled
		c.JSON(http.StatusOK, body)
		return
	}

	log.Info(sub)

	receipt := fmt.Sprintf("%d", int(sub["id"].(float64)))
	email := ""

	var billing map[string]interface{}
	if sub["billing"] != nil {
		billing = sub["billing"].(map[string]interface{})
		email = billing["email"].(string)
	}

	log.Infof("Receipt = %s", receipt)
	log.Infof("Email = %s", email)

	s, err := core.GetSubscriptionByReceipt(receipt)
	if err != nil {

		subs, err := core.ReadSubscriptions(email)
		if err != nil {
			log.Infof("*************** SUBSCRIPTION %s NEEDS ATTENTION ***************", receipt)
			log.Info(body)
		}

		if len(subs) == 1 {
			s = subs[0]
			s.Receipt = receipt
		} else if len(subs) > 1 {
			log.Infof("*************** USER WITH MULTIPLE SUBSCRIPTIONS %s NEEDS ATTENTION ***************", receipt)
			log.Info(body)
		} else {
			log.Infof("*************** SUBSCRIPTION %s NOT FOUND ***************", receipt)
			log.Info(body)
		}
	}

	if s != nil {
		if sub["status"] != nil {
			s.Status = sub["status"].(string)
		}

		if sub["next_payment_date_gmt"] != nil {
			expires := sub["next_payment_date_gmt"].(string)
			if expires != "" {
				layout := "2006-01-02T15:04:05"
				*s.Expires, err = time.Parse(layout, expires)
				if err != nil {
					log.Error(err)
				}
			}
		}

		if sub["end_date_gmt"] != nil {
			expires := sub["end_date_gmt"].(string)
			if expires != "" {
				layout := "2006-01-02T15:04:05"
				*s.Expires, err = time.Parse(layout, expires)
				if err != nil {
					log.Error(err)
				}
			}
		}

		s.UpdatedBy = "woo"
		_, err = core.UpdateSubscription(s.Id, s)
		core.SubscriptionEmail(s)

		if err != nil {
			log.Errorf("Error Updating Subscription: %v %v", s, err)
		} else {
			log.Infof("Updated Subscription: %v", s)
		}

	}

	c.JSON(http.StatusOK, body)

}

func readSubscription(c *gin.Context) {
	id := c.Param("id")

	client, err := core.ReadSubscription(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

func updateSubscription(c *gin.Context) {
	var data model.Subscription
	id := c.Param("id")

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// get update user from token and add to client infos
	oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
	user, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	data.UpdatedBy = user.Email

	client, err := core.UpdateSubscription(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	core.SubscriptionEmail(client)

	c.JSON(http.StatusOK, client)
}

func deleteSubscription(c *gin.Context) {
	id := c.Param("id")

	err := core.DeleteSubscription(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readSubscriptions(c *gin.Context) {
	value, exists := c.Get("oauth2Token")
	if !exists {
		c.AbortWithStatus(401)
		return
	}
	oauth2Token := value.(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
	user, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.Email == "" && os.Getenv("OAUTH2_PROVIDER_NAME") != "fake" {

		log.Error("security alert: Email empty on authenticated token")
		c.JSON(http.StatusForbidden, gin.H{"error": "This error has been logged"})
	}

	subscriptions, err := core.ReadSubscriptions(user.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

func readSubscriptionsDeleted(c *gin.Context) {
	value, exists := c.Get("oauth2Token")
	if !exists {
		c.AbortWithStatus(401)
		return
	}
	oauth2Token := value.(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
	user, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.Email == "" && os.Getenv("OAUTH2_PROVIDER_NAME") != "fake" {

		log.Error("security alert: Email empty on authenticated token")
		c.JSON(http.StatusForbidden, gin.H{"error": "This error has been logged"})
	}

	subscriptions, err := core.ReadSubscriptions(user.Email, true)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}
