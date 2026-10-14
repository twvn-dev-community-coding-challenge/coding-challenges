package sms

import (
	"time"

	"github.com/dotdak/sms-otp/internal/providers"
)

// MessageListParams filters and paginates ListMessages. Zero values mean "no filter" for each field.
// Limit 0 means no limit (return all matches). HTTP handlers should set a positive default and cap.
type MessageListParams struct {
	Status providers.MessageStatus
	Phone  string
	Since  *time.Time
	Limit  int
	Offset int
}
