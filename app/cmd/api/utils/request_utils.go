package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendPost(userID uint, oldIP string, newIP string, webhookURL string) {
	payload := map[string]string{
		"event":   "new_ip_detected",
		"user_id": fmt.Sprintf("%d", userID),
		"old_ip":  oldIP,
		"new_ip":  newIP,
		"message": "Attempt to refresh tokens from new IP address",
	}
	jsonData, _ := json.Marshal(payload)
	_, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		println("Failed to send webhook:", err.Error())
	}
}
