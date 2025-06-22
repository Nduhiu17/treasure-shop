package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// PaymentService will handle payment processing logic
// You would integrate with a third-party payment gateway here

type PaymentService struct {
	// Add any dependencies for your payment gateway (e.g., API keys)
}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (s *PaymentService) ProcessPayment(orderID string, paymentInfo map[string]interface{}) (bool, error) {
	// In a real application, you would interact with a payment gateway API
	// For example, using Stripe:
	// stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	// _, err := charge.New(params)
	// if err != nil {
	// 	return false, err
	// }
	// return true, nil

	// For now, simulate success
	println("Processing payment for order:", orderID, "with info:", paymentInfo)
	return true, nil
}

type PayPalOrder struct {
	Intent        string `json:"intent"`
	PurchaseUnits []struct {
		Amount struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"amount"`
	} `json:"purchase_units"`
	ApplicationContext *struct {
		BrandName          string `json:"brand_name"`
		Locale             string `json:"locale"`
		ShippingPreference string `json:"shipping_preference"`
		UserAction         string `json:"user_action"`
		ReturnURL          string `json:"return_url"`
		CancelURL          string `json:"cancel_url"`
	} `json:"application_context,omitempty"`
}

type PayPalOrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

type PayPalCaptureResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				} `json:"amount"`
				CreateTime string `json:"create_time"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

func getPayPalAccessToken() (string, error) {
	paypalClientID := os.Getenv("PAYPAL_CLIENT_ID")
	paypalClientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")
	paypalBaseURL := os.Getenv("PAYPAL_BASE_URL")

	reqBody := bytes.NewBufferString("grant_type=client_credentials")
	req, err := http.NewRequest("POST", paypalBaseURL+"/v1/oauth2/token", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create access token request: %w", err)
	}
	req.SetBasicAuth(paypalClientID, paypalClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send access token request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("paypal access token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to decode access token response: %w", err)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}
	return accessToken, nil
}

func (s *PaymentService) CreatePayPalOrder(amount float64, currency, returnURL, cancelURL string) (string, error) {
	paypalBaseURL := os.Getenv("PAYPAL_BASE_URL")
	accessToken, err := getPayPalAccessToken()
	if err != nil {
		return "", err
	}
	orderRequest := PayPalOrder{
		Intent: "CAPTURE",
		PurchaseUnits: []struct {
			Amount struct {
				CurrencyCode string `json:"currency_code"`
				Value        string `json:"value"`
			} `json:"amount"`
		}{
			{
				Amount: struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				}{
					CurrencyCode: currency,
					Value:        strconv.FormatFloat(amount, 'f', 2, 64),
				},
			},
		},
		ApplicationContext: &struct {
			BrandName          string `json:"brand_name"`
			Locale             string `json:"locale"`
			ShippingPreference string `json:"shipping_preference"`
			UserAction         string `json:"user_action"`
			ReturnURL          string `json:"return_url"`
			CancelURL          string `json:"cancel_url"`
		}{
			BrandName:          "Academic Codebase",
			Locale:             "en-US",
			ShippingPreference: "NO_SHIPPING",
			UserAction:         "PAY_NOW",
			ReturnURL:          returnURL,
			CancelURL:          cancelURL,
		},
	}
	jsonOrderRequest, err := json.Marshal(orderRequest)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", paypalBaseURL+"/v2/checkout/orders", bytes.NewBuffer(jsonOrderRequest))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("PayPal order creation failed: %s", string(bodyBytes))
	}
	var orderResponse PayPalOrderResponse
	if err := json.Unmarshal(bodyBytes, &orderResponse); err != nil {
		return "", err
	}
	return orderResponse.ID, nil
}

func (s *PaymentService) CapturePayPalOrder(orderID string) (string, error) {
	paypalBaseURL := os.Getenv("PAYPAL_BASE_URL")
	accessToken, err := getPayPalAccessToken()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", paypalBaseURL+"/v2/checkout/orders/"+orderID+"/capture", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("PayPal capture failed: %s", string(bodyBytes))
	}
	var captureResponse PayPalCaptureResponse
	if err := json.Unmarshal(bodyBytes, &captureResponse); err != nil {
		return "", err
	}
	return captureResponse.Status, nil
}
