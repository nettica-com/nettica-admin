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
		g.POST("/apple", createSubscriptionApple)
		g.POST("/apple/webhook", handleAppleWebhook)
		g.POST("/apple/discount", handleAppleDiscount)
		g.POST("/android", createSubscriptionAndroid)
		g.POST("/android/webhook", handleAndroidWebhook)
		g.GET("/offers/:id", getOffers)
		g.GET("/:id", readSubscription)
		g.PATCH("/:id", updateSubscription)
		g.DELETE("/:id", deleteSubscription)
		g.GET("", readSubscriptions)
	}
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
		c.JSON(http.StatusOK, s)
		return
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
		name = "Basic Service (monthly)"
	case "basic_yearly":
		name = "Basic Service (yearly)"
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
		if sub["subscriptionState"] != nil && sub["subscriptionState"].(string) == "SUBSCRIPTION_STATE_ACTIVE" {
			// retrieve the subscription
			subscription, err := core.GetSubscriptionByReceipt(purchaseToken)
			if err != nil {
				// create the subscription

				log.Errorf("error getting subscription: %v", err)
				c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
				return
			}

			if subscription.Status == "expired" {
				subscription.Status = "active"
				last := time.Now().UTC()
				subscription.LastUpdated = &last
				subscription.UpdatedBy = "google"
				core.RenewSubscription(subscription.Id)
				log.Infof("subscription renewed: %s", subscription.Id)
				subscription, err = core.GetSubscriptionByReceipt(purchaseToken)
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
			if zero["expiryTime"] != nil {
				*subscription.Expires, err = time.Parse(time.RFC3339, zero["expiryTime"].(string))
				if err != nil {
					log.Errorf("error parsing expiryTime: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					return
				}
				last := time.Now().UTC()
				subscription.LastUpdated = &last
				subscription.UpdatedBy = "google"
				core.UpdateSubscription(subscription.Id, subscription)
				log.Infof("subscription updated: %s %v", subscription.Id, subscription)
				c.JSON(http.StatusOK, gin.H{"status": "updated"})
				return
			}
		}

		if sub["subscriptionState"] != nil && sub["subscriptionState"].(string) == "SUBSCRIPTION_STATE_EXPIRED" {
			// retrieve the subscription
			subscription, err := core.GetSubscriptionByReceipt(purchaseToken)
			if err != nil {
				log.Errorf("error getting subscription: %v", err)
				c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
				return
			}

			core.ExpireSubscription(subscription.Id)

			log.Infof("subscription expired: %s", subscription.Id)

			c.JSON(http.StatusOK, gin.H{"status": "expired"})
			return
		}

		// handle cancel and did_not_renew

	}

	if msg["voidedPurchaseNotification"] != nil {
		voidedPurchaseNotification := msg["voidedPurchaseNotification"].(map[string]interface{})
		log.Infof("voidedPurchaseNotification: %v", voidedPurchaseNotification)

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
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		log.Error(err)
	}
	header := string(headerBytes)
	log.Infof("header: %s", header)

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

	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		log.Error(err)
	}
	signature := string(signatureBytes)
	log.Infof("signature: %s", signature)

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
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}
		last := time.Now().UTC()
		subscription.LastUpdated = &last
		subscription.UpdatedBy = "apple"

		switch transactionReason {
		case "PURCHASE":
			subscription.Expires = &expires
			core.UpdateSubscription(subscription.Id, subscription)
			log.Infof("apple: subscription PURCHASE updated: %s until %s", subscription.Id, expires)

		case "RENEWAL":
			subscription.Expires = &expires
			if subscription.Status == "cancelled" || subscription.Status == "expired" {
				subscription.Status = "active"
				core.UpdateSubscription(subscription.Id, subscription)
				core.RenewSubscription(subscription.Id)
			} else {
				core.UpdateSubscription(subscription.Id, subscription)
			}
			log.Infof("apple: subscription RENWAL: %s until %s", subscription.Id, expires)

		case "CANCEL":
			subscription.Status = "cancelled"
			core.UpdateSubscription(subscription.Id, subscription)
			log.Infof("apple: subscription CANCEL: %s", subscription.Id)

		case "DID_NOT_RENEW":
			subscription.Status = "expired"
			core.UpdateSubscription(subscription.Id, subscription)
			core.ExpireSubscription(subscription.Id)
			log.Infof("apple: subscription DID_NOT_RENEW: %s at %s", subscription.Id, expires)
		}

	}
	// Respond with 200 OK to acknowledge receipt of the webhook
	c.JSON(http.StatusOK, gin.H{"status": "received"})

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
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
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

	log.Infof("header: %v", headerMap)

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

	_, err = core.GetSubscriptionByReceipt(receipt.Receipt)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Subscription already exists"})
		return
	}

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
		originalTransactionId = ""
	}
	log.Infof("originalTransactionId: %s", originalTransactionId)
	if originalTransactionId != "" {
		subscription, err := core.GetSubscriptionByReceipt(originalTransactionId)
		if err != nil {
			log.Error(err)
			receipt.Receipt = originalTransactionId
		} else {
			last := time.Now().UTC()
			productId := result["productId"].(string)
			var expires time.Time
			if productId == "24_hours_flex" || productId == "10_day_flex" {
				expires = time.Now().Add(24 * time.Hour)
				if productId == "10_day_flex" {
					expires = time.Now().AddDate(0, 0, 10)
				}
			} else {
				expiresDate := result["expiresDate"].(float64)
				expires = time.Unix(int64(expiresDate)/1000, 0)
			}
			log.Infof("expires: %s", expires)
			transactionReason := result["transactionReason"].(string)
			if transactionReason == "RENEWAL" {
				subscription.Expires = &expires
				subscription.LastUpdated = &last
				if subscription.Status == "cancelled" || subscription.Status == "expired" {
					subscription.Status = "active"
					core.UpdateSubscription(subscription.Id, subscription)
					core.RenewSubscription(subscription.Id)
					log.Infof("subscription renewed: %s until %s", subscription.Id, expires)
				} else {
					core.UpdateSubscription(subscription.Id, subscription)
					log.Infof("subscription updated: %s %v", subscription.Id, subscription)
					c.JSON(http.StatusOK, subscription)
					return
				}

			} else {
				log.Infof("createSubscriptionApple: transactionReason: %s", transactionReason)
				subscription.Expires = &expires
				subscription.LastUpdated = &last
				core.UpdateSubscription(subscription.Id, subscription)
				log.Infof("subscription updated: %s %v", subscription.Id, subscription)
				c.JSON(http.StatusOK, subscription)
				return
			}
			log.Infof("*** subscription: %v", subscription)
			log.Infof("*** result: %v", result)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
			// handle cancel and did_not_renew
		}
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
	expires := time.Now().AddDate(0, 2, 0)

	switch receipt.ProductID {
	case "basic_monthly":
		name = "Basic Service (monthly)"
	case "basic_yearly":
		name = "Basic Service (yearly)"
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
	subscription := model.Subscription{
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
	return

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
			name = "Relay Service"
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

	receipt := sub["id"].(string)
	log.Infof("receipt: %s", receipt)

	// walk the json and find the customer href
	links := sub["_links"].(map[string]interface{})
	log.Info(links)

	// get the sku from the line_items
	sku := sub["line_items"].([]interface{})[0].(map[string]interface{})["sku"].(string)
	//status := sub["status"].(string)

	status := "active"

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
			name = "Relay Service"
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

	receipt := sub["id"].(string)
	email := ""

	var billing map[string]interface{}
	if sub["billing"] != nil {
		billing = sub["billing"].(map[string]interface{})
		email = billing["email"].(string)
	}

	log.Infof("Receipt = %s", receipt)
	log.Infof("Email = %s", email)

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
