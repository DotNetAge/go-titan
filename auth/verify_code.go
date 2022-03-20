package auth

import (
	"time"
)

type VerifyCode struct {
	ClientID string    `json:"client_id" gorm:"primaryKey;type:varchar(50)" pg:"client_id,type:varchar(50),pk"`
	Mobile   string    `json:"mobile" gorm:"primaryKey;type:varchar(20)" pg:"mobile,type:varchar(20),pk"`
	Code     string    `json:"code" pg:"code"`
	ExpireAt time.Time `json:"expire_at" pg:"expire_at"`
}

func NewVerifyCode(clientId, mobile string, duration time.Duration) *VerifyCode {
	return &VerifyCode{
		ClientID: clientId,
		Mobile:   mobile,
		Code:     GenVerifyCode(),
		ExpireAt: time.Now().Add(duration),
	}
}