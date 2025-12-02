package repository

import (
	"errors"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
)

type SMSRepository struct{}

func NewSMSRepository() *SMSRepository {
	return &SMSRepository{}
}

func (r *SMSRepository) Create(code *models.SMSCode) error {
	query := `
		INSERT INTO sms_codes (id, phone, code, used, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	code.ID = uuid.New()
	code.CreatedAt = time.Now()

	row := database.DB.Raw(
		query, code.ID, code.Phone, code.Code, code.Used, code.ExpiresAt, code.CreatedAt,
	).Row()
	return row.Scan(&code.ID)
}

func (r *SMSRepository) GetValidCode(phone, code string) (*models.SMSCode, error) {
	smsCode := &models.SMSCode{}
	query := `
		SELECT id, phone, code, used, expires_at, created_at
		FROM sms_codes
		WHERE phone = $1 AND code = $2 AND used = FALSE AND expires_at > $3
		ORDER BY created_at DESC
		LIMIT 1
	`
	now := time.Now()
	result := database.DB.Raw(query, phone, code, now).Scan(smsCode)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("invalid or expired code")
	}
	return smsCode, nil
}

func (r *SMSRepository) MarkAsUsed(id uuid.UUID) error {
	query := `UPDATE sms_codes SET used = TRUE WHERE id = $1`
	result := database.DB.Exec(query, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("code not found")
	}
	return nil
}

func (r *SMSRepository) GetRecentCodeCount(phone string, since time.Time) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*)
		FROM sms_codes
		WHERE phone = $1 AND created_at > $2
	`
	err := database.DB.Raw(query, phone, since).Scan(&count).Error
	return count, err
}


