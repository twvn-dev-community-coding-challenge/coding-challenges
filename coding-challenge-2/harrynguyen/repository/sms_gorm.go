package repository

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
	"gorm.io/gorm"
)

type smsMessageModel struct {
	ID                string `gorm:"primaryKey;type:varchar(64)"`
	ProviderMessageID string `gorm:"column:provider_message_id;type:varchar(255);index"`
	SendSource        string `gorm:"column:send_source;type:varchar(32);index"`
	Country           string `gorm:"type:varchar(8);not null"`
	PhoneNumber       string `gorm:"type:varchar(32);index;not null"`
	Content           string `gorm:"type:text;not null"`
	Carrier           string `gorm:"type:varchar(64);not null"`
	Provider          string `gorm:"type:varchar(64);not null"`
	Status            string `gorm:"type:varchar(32);index;not null"`
	EstimatedCost     float64
	ActualCost        float64
	CreatedAt         time.Time `gorm:"index"`
	UpdatedAt         time.Time
}

func (*smsMessageModel) TableName() string { return "sms_messages" }

type smsStatusLogModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(64)"`
	SMSID     string    `gorm:"column:sms_id;type:varchar(64);index:idx_sms_status_logs_sms_id;not null"`
	Status    string    `gorm:"type:varchar(32);not null"`
	Timestamp time.Time `gorm:"index"`
	Metadata  string    `gorm:"type:text"`
}

func (*smsStatusLogModel) TableName() string { return "sms_status_logs" }

// SMSGormRepository persists SMS messages and status logs in Postgres via GORM.
type SMSGormRepository struct {
	db        *gorm.DB
	createSeq atomic.Uint64
}

// NewSMSGormRepository returns a sms.Repository backed by Postgres.
func NewSMSGormRepository(db *gorm.DB) *SMSGormRepository {
	return &SMSGormRepository{db: db}
}

// AutoMigrateSMS creates or updates sms_messages and sms_status_logs tables.
func AutoMigrateSMS(db *gorm.DB) error {
	return db.AutoMigrate(&smsMessageModel{}, &smsStatusLogModel{})
}

func (r *SMSGormRepository) Create(ctx context.Context, msg *providers.SMSMessage) error {
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("sms_%d_%d", time.Now().UnixNano(), r.createSeq.Add(1))
	}
	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = now
	}
	row := domainToModel(msg)
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *SMSGormRepository) Update(ctx context.Context, msg *providers.SMSMessage) error {
	res := r.db.WithContext(ctx).Model(&smsMessageModel{}).Where("id = ?", msg.ID).Updates(map[string]interface{}{
		"provider_message_id": msg.MessageID,
		"country":             msg.Country,
		"phone_number":        msg.PhoneNumber,
		"content":             msg.Content,
		"carrier":             string(msg.Carrier),
		"provider":            string(msg.Provider),
		"status":              string(msg.Status),
		"estimated_cost":      msg.EstimatedCost,
		"actual_cost":         msg.ActualCost,
		"updated_at":          time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return sms.ErrNotFound
	}
	msg.UpdatedAt = time.Now()
	return nil
}

func (r *SMSGormRepository) GetByID(ctx context.Context, id string) (*providers.SMSMessage, error) {
	var row smsMessageModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sms.ErrNotFound
		}
		return nil, err
	}
	return modelToDomain(&row), nil
}

func (r *SMSGormRepository) GetByMessageID(ctx context.Context, messageID string) (*providers.SMSMessage, error) {
	var row smsMessageModel
	if err := r.db.WithContext(ctx).Where("provider_message_id = ?", messageID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sms.ErrNotFound
		}
		return nil, err
	}
	return modelToDomain(&row), nil
}

func (r *SMSGormRepository) ListMessages(ctx context.Context, p sms.MessageListParams) ([]*providers.SMSMessage, error) {
	q := r.db.WithContext(ctx).Model(&smsMessageModel{})
	if p.Status != "" {
		q = q.Where("status = ?", string(p.Status))
	}
	if p.Phone != "" {
		q = q.Where("phone_number LIKE ?", "%"+p.Phone+"%")
	}
	if p.Since != nil {
		q = q.Where("created_at >= ?", *p.Since)
	}
	q = q.Order("created_at DESC")
	if p.Limit > 0 {
		q = q.Limit(p.Limit)
	}
	if p.Offset > 0 {
		q = q.Offset(p.Offset)
	}
	var rows []smsMessageModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*providers.SMSMessage, 0, len(rows))
	for i := range rows {
		out = append(out, modelToDomain(&rows[i]))
	}
	return out, nil
}

func (r *SMSGormRepository) AddStatusLog(ctx context.Context, log *providers.StatusLog) error {
	if log.ID == "" {
		log.ID = fmt.Sprintf("log_%d", time.Now().UnixNano())
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	row := smsStatusLogModel{
		ID:        log.ID,
		SMSID:     log.SMSID,
		Status:    string(log.Status),
		Timestamp: log.Timestamp,
		Metadata:  log.Metadata,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *SMSGormRepository) GetStatusLogs(ctx context.Context, smsID string) ([]*providers.StatusLog, error) {
	var rows []smsStatusLogModel
	if err := r.db.WithContext(ctx).Where("sms_id = ?", smsID).Order("timestamp ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*providers.StatusLog, 0, len(rows))
	for i := range rows {
		out = append(out, &providers.StatusLog{
			ID:        rows[i].ID,
			SMSID:     rows[i].SMSID,
			Status:    providers.MessageStatus(rows[i].Status),
			Timestamp: rows[i].Timestamp,
			Metadata:  rows[i].Metadata,
		})
	}
	return out, nil
}

func domainToModel(msg *providers.SMSMessage) smsMessageModel {
	return smsMessageModel{
		ID:                msg.ID,
		ProviderMessageID: msg.MessageID,
		SendSource:        msg.SendSource,
		Country:           msg.Country,
		PhoneNumber:       msg.PhoneNumber,
		Content:           msg.Content,
		Carrier:           string(msg.Carrier),
		Provider:          string(msg.Provider),
		Status:            string(msg.Status),
		EstimatedCost:     msg.EstimatedCost,
		ActualCost:        msg.ActualCost,
		CreatedAt:         msg.CreatedAt,
		UpdatedAt:         msg.UpdatedAt,
	}
}

func modelToDomain(m *smsMessageModel) *providers.SMSMessage {
	return &providers.SMSMessage{
		ID:            m.ID,
		MessageID:     m.ProviderMessageID,
		SendSource:    m.SendSource,
		Country:       m.Country,
		PhoneNumber:   m.PhoneNumber,
		Content:       m.Content,
		Carrier:       providers.Carrier(m.Carrier),
		Provider:      providers.Provider(m.Provider),
		Status:        providers.MessageStatus(m.Status),
		EstimatedCost: m.EstimatedCost,
		ActualCost:    m.ActualCost,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
