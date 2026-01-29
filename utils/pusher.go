package utils

import (
	"log"

	"github.com/gadhittana01/cases-modules/utils"
	pusher "github.com/pusher/pusher-http-go/v5"
)

// NewPusherClient creates a new Pusher client from config
func NewPusherClient(config *utils.Config) *pusher.Client {
	if config.PusherAppID == "" || config.PusherKey == "" || config.PusherSecret == "" {
		log.Println("Warning: Pusher credentials not configured, Pusher events will be disabled")
		return nil
	}

	client := pusher.Client{
		AppID:   config.PusherAppID,
		Key:     config.PusherKey,
		Secret:  config.PusherSecret,
		Cluster: config.PusherCluster,
		Secure:  true,
	}

	return &client
}

// EmitPaymentStatus emits a payment status update via Pusher
func EmitPaymentStatus(client *pusher.Client, channel, event string, data interface{}) error {
	if client == nil {
		return nil // Silently skip if Pusher not configured
	}

	err := client.Trigger(channel, event, data)
	if err != nil {
		log.Printf("Failed to emit Pusher event: %v", err)
		return err
	}

	return nil
}
