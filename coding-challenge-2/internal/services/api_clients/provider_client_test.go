package api_clients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProviderAndClient_Twilio(t *testing.T) {
	client, err := GetProviderAndClient(Twilio)
	require.NoError(t, err)
	assert.NotNil(t, client)
	_, ok := client.(*TwilioAPIClient)
	assert.True(t, ok)
}

func TestGetProviderAndClient_Vonage(t *testing.T) {
	client, err := GetProviderAndClient(Vonage)
	require.NoError(t, err)
	assert.NotNil(t, client)
	_, ok := client.(*VonageAPIClient)
	assert.True(t, ok)
}

func TestGetProviderAndClient_Unsupported(t *testing.T) {
	_, err := GetProviderAndClient("provider-unknown")
	assert.Error(t, err)
}

func TestTwilioAPIClient_GetCostEstimation(t *testing.T) {
	client := &TwilioAPIClient{}
	ctx := context.Background()

	cost, err := client.GetCostEstimation(ctx, EstimationRequest{
		ProviderID:     string(Twilio),
		CarrierID:      "carrier-001",
		RecipientPhone: "931234567",
		Content:        "Hello",
	})

	require.NoError(t, err)
	assert.Equal(t, "500", cost.Amount)
	assert.Equal(t, "VND", cost.Currency)
}

func TestTwilioAPIClient_Send(t *testing.T) {
	client := &TwilioAPIClient{}
	ctx := context.Background()

	err := client.Send(ctx, SendRequest{
		MessageID:      "msg-001",
		Content:        "Hello",
		RecipientPhone: "931234567",
	})
	assert.NoError(t, err)
}

func TestVonageAPIClient_GetCostEstimation(t *testing.T) {
	client := &VonageAPIClient{}
	ctx := context.Background()

	cost, err := client.GetCostEstimation(ctx, EstimationRequest{
		ProviderID:     string(Vonage),
		CarrierID:      "carrier-001",
		RecipientPhone: "931234567",
		Content:        "Hello",
	})

	require.NoError(t, err)
	assert.Equal(t, "1000", cost.Amount)
	assert.Equal(t, "VND", cost.Currency)
}

func TestVonageAPIClient_Send(t *testing.T) {
	client := &VonageAPIClient{}
	ctx := context.Background()

	err := client.Send(ctx, SendRequest{
		MessageID:      "msg-001",
		Content:        "Hello",
		RecipientPhone: "931234567",
	})
	assert.NoError(t, err)
}
