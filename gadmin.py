import os
import requests
from dotenv import load_dotenv
import sys

# Load environment variables from .env file
load_dotenv()

# Get the Google Play credentials from environment variables
client_id = os.getenv('GOOGLE_PLAY_CLIENT_ID')
client_secret = os.getenv('GOOGLE_PLAY_CLIENT_SECRET')
refresh_token = os.getenv('GOOGLE_PLAY_REFRESH_TOKEN')
access_url = os.getenv('GOOGLE_PLAY_ACCESS_URL')

# Get command line arguments
args = sys.argv

# Ensure the correct number of arguments are provided
if len(args) != 3:
    print("Usage: gadmin.py <product_id> <subscription_token>")
    sys.exit(1)

# Assign command line arguments to variables
product_id = args[1]
subscription_token = args[2]


# Create the request payload
payload = {
    'grant_type': 'refresh_token',
    'client_id': client_id,
    'client_secret': client_secret,
    'refresh_token': refresh_token
}

# Make the HTTP request to get the auth token
response = requests.post(access_url, data=payload)

# Check if the request was successful
if response.status_code == 200:
    auth_token = response.json().get('access_token')
    print(f'Auth Token: {auth_token}')
else:
    print(f'Failed to get auth token: {response.status_code}')
    print(response.json())

# Define the URL to get subscription details
subscription_details_url = f"https://www.googleapis.com/androidpublisher/v3/applications/com.nettica.agent/purchases/subscriptions/{product_id}/tokens/{subscription_token}"

# Set the headers for the request
headers = {
    'Authorization': f'Bearer {auth_token}',
    'Accept': 'application/json'
}

# Make the HTTP request to get the subscription details
response = requests.get(subscription_details_url, headers=headers)

# Check if the request was successful
if response.status_code == 200:
    subscription_details = response.json()
    print(f'Subscription Details: {subscription_details}')
else:
    print(f'Failed to get subscription details: {response.status_code}')
    print(response.json())

# Extract the latestOrderId from the subscription details
order_id = subscription_details.get('orderId')

if not order_id:
    print('Failed to get orderId from subscription details')
    sys.exit(1)

# Define the URL to get order details
order_details_url = f"https://www.googleapis.com/androidpublisher/v3/applications/com.nettica.agent/purchases/orders/{order_id}"

# Make the HTTP request to get the order details
response = requests.get(order_details_url, headers=headers)

# Check if the request was successful
if response.status_code == 200:
    order_details = response.json()
    print(f'Order Details: {order_details}')
else:
    print(f'Failed to get order details: {response.status_code}')
    print(response.json())