package sms

import (
	"context"
	"testing"
	"time"

	"github.com/dotdak/sms-otp/internal/providers"
)

func TestInMemRepository_ListMessages_phoneFilter(t *testing.T) {
	repo := NewInMemRepository()
	ctx := context.Background()
	m := &providers.SMSMessage{
		ID: "m1", Country: "PH", PhoneNumber: "6391712345000", Content: "x",
		Carrier: providers.CarrierGlobe, Provider: providers.ProviderMessageBird,
		Status: providers.StatusQueue, CreatedAt: time.Now(),
	}
	if err := repo.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	out, err := repo.ListMessages(ctx, MessageListParams{Phone: "1712345", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("len=%d", len(out))
	}
}

func TestInMemRepository_ListMessages_offsetAndLimit(t *testing.T) {
	repo := NewInMemRepository()
	ctx := context.Background()
	t0 := time.Now().Add(-2 * time.Hour)
	for i, id := range []string{"a", "b", "c"} {
		m := &providers.SMSMessage{
			ID: id, Country: "PH", PhoneNumber: "639171111000" + id, Content: "x",
			Carrier: providers.CarrierGlobe, Provider: providers.ProviderMessageBird,
			Status: providers.StatusQueue, CreatedAt: t0.Add(time.Duration(i) * time.Minute),
		}
		if err := repo.Create(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	// Offset past all rows
	empty, err := repo.ListMessages(ctx, MessageListParams{Limit: 10, Offset: 99})
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("offset 99: len=%d", len(empty))
	}
	// Negative offset treated as 0
	all, err := repo.ListMessages(ctx, MessageListParams{Limit: 10, Offset: -3})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("negative offset: len=%d", len(all))
	}
}

func TestInMemRepository_ListMessages_phoneFilterCaseInsensitive(t *testing.T) {
	repo := NewInMemRepository()
	ctx := context.Background()
	m := &providers.SMSMessage{
		ID: "m2", Country: "PH", PhoneNumber: "6391712345000", Content: "x",
		Carrier: providers.CarrierGlobe, Provider: providers.ProviderMessageBird,
		Status: providers.StatusQueue, CreatedAt: time.Now(),
	}
	if err := repo.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	out, err := repo.ListMessages(ctx, MessageListParams{Phone: "1712345", Limit: 10})
	if err != nil || len(out) != 1 {
		t.Fatalf("err=%v len=%d", err, len(out))
	}
	out2, err := repo.ListMessages(ctx, MessageListParams{Phone: "NOTINSIDE", Limit: 10})
	if err != nil || len(out2) != 0 {
		t.Fatalf("non-matching filter: err=%v len=%d", err, len(out2))
	}
}

func TestInMemRepository_ListMessages_emptyPhoneMeansNoFilter(t *testing.T) {
	repo := NewInMemRepository()
	ctx := context.Background()
	if err := repo.Create(ctx, &providers.SMSMessage{
		ID: "m3", Country: "PH", PhoneNumber: "6391700000001", Content: "x",
		Carrier: providers.CarrierGlobe, Provider: providers.ProviderMessageBird,
		Status: providers.StatusQueue, CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	out, err := repo.ListMessages(ctx, MessageListParams{Limit: 10})
	if err != nil || len(out) != 1 {
		t.Fatalf("err=%v len=%d", err, len(out))
	}
}
