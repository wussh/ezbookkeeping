package models

// PayLaterInstallmentStatus represents the status of a single installment
type PayLaterInstallmentStatus byte

// Pay later installment statuses
const (
	PAY_LATER_INSTALLMENT_STATUS_PENDING  PayLaterInstallmentStatus = 1
	PAY_LATER_INSTALLMENT_STATUS_PAID     PayLaterInstallmentStatus = 2
	PAY_LATER_INSTALLMENT_STATUS_OVERDUE  PayLaterInstallmentStatus = 3
)

// PayLaterInstallment represents a single installment within a pay later plan stored in database
type PayLaterInstallment struct {
	InstallmentId     int64                     `xorm:"PK"`
	PlanId            int64                     `xorm:"INDEX(IDX_pay_later_installment_plan_id_uid) NOT NULL"`
	Uid               int64                     `xorm:"INDEX(IDX_pay_later_installment_plan_id_uid) NOT NULL"`
	InstallmentNumber int32                     `xorm:"NOT NULL"`
	DueDate           int64                     `xorm:"NOT NULL"`
	Amount            int64                     `xorm:"NOT NULL"`
	PaidTransactionId int64                     `xorm:"NOT NULL"`
	Status            PayLaterInstallmentStatus `xorm:"NOT NULL"`
	CreatedUnixTime   int64
	UpdatedUnixTime   int64
}

// PayLaterInstallmentPayRequest represents all parameters of pay installment request
type PayLaterInstallmentPayRequest struct {
	Id          int64 `json:"id,string" binding:"required,min=1"`
	PaymentTime int64 `json:"paymentTime" binding:"required,min=1"`
	UtcOffset   int16 `json:"utcOffset" binding:"min=-720,max=840"`
}

// PayLaterInstallmentResponse represents a single installment response
type PayLaterInstallmentResponse struct {
	Id                string                    `json:"id"`
	PlanId            string                    `json:"planId"`
	InstallmentNumber int32                     `json:"installmentNumber"`
	DueDate           int64                     `json:"dueDate"`
	Amount            int64                     `json:"amount"`
	PaidTransactionId string                    `json:"paidTransactionId,omitempty"`
	Status            PayLaterInstallmentStatus `json:"status"`
}
