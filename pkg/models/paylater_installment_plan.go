package models

// PayLaterPlanStatus represents the status of a pay later installment plan
type PayLaterPlanStatus byte

// Pay later plan statuses
const (
	PAY_LATER_PLAN_STATUS_ACTIVE    PayLaterPlanStatus = 1
	PAY_LATER_PLAN_STATUS_COMPLETED PayLaterPlanStatus = 2
	PAY_LATER_PLAN_STATUS_CANCELLED PayLaterPlanStatus = 3
)

// PayLaterInstallmentPlan represents a pay later installment plan stored in database
type PayLaterInstallmentPlan struct {
	PlanId            int64              `xorm:"PK"`
	Uid               int64              `xorm:"INDEX(IDX_pay_later_plan_uid_deleted_status) NOT NULL"`
	Deleted           bool               `xorm:"INDEX(IDX_pay_later_plan_uid_deleted_status) NOT NULL"`
	Status            PayLaterPlanStatus `xorm:"INDEX(IDX_pay_later_plan_uid_deleted_status) NOT NULL"`
	Name              string             `xorm:"VARCHAR(64) NOT NULL"`
	TotalAmount       int64              `xorm:"NOT NULL"`
	InstallmentCount  int32              `xorm:"NOT NULL"`
	PurchaseDate      int64              `xorm:"NOT NULL"`
	AccountId         int64              `xorm:"NOT NULL"`
	PaymentAccountId  int64              `xorm:"NOT NULL"`
	CategoryId        int64              `xorm:"NOT NULL"`
	TagIds            string             `xorm:"VARCHAR(255) NOT NULL"`
	Comment           string             `xorm:"VARCHAR(255) NOT NULL"`
	CreatedUnixTime   int64
	UpdatedUnixTime   int64
	DeletedUnixTime   int64
}

// PayLaterInstallmentPlanCreateRequest represents all parameters of pay later plan creation request
type PayLaterInstallmentPlanCreateRequest struct {
	Name             string   `json:"name" binding:"required,notBlank,max=64"`
	TotalAmount      int64    `json:"totalAmount" binding:"required,min=1"`
	InstallmentCount int32    `json:"installmentCount" binding:"required,min=1,max=12"`
	PurchaseDate     int64    `json:"purchaseDate" binding:"required,min=1"`
	AccountId        int64    `json:"accountId,string" binding:"required,min=1"`
	PaymentAccountId int64    `json:"paymentAccountId,string" binding:"required,min=1"`
	CategoryId       int64    `json:"categoryId,string" binding:"min=0"`
	TagIds           []string `json:"tagIds"`
	Comment          string   `json:"comment" binding:"max=255"`
	ClientSessionId  string   `json:"clientSessionId"`
}

// PayLaterInstallmentPlanDeleteRequest represents all parameters of pay later plan delete request
type PayLaterInstallmentPlanDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// PayLaterInstallmentPlanGetRequest represents all parameters of pay later plan get request
type PayLaterInstallmentPlanGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// PayLaterInstallmentPlanListRequest represents all parameters of pay later plan list request
type PayLaterInstallmentPlanListRequest struct {
	Status PayLaterPlanStatus `form:"status" binding:"min=0"`
}

// PayLaterInstallmentPlanInfoResponse represents a pay later plan response
type PayLaterInstallmentPlanInfoResponse struct {
	Id               string                         `json:"id"`
	Name             string                         `json:"name"`
	TotalAmount      int64                          `json:"totalAmount"`
	InstallmentCount int32                          `json:"installmentCount"`
	PurchaseDate     int64                          `json:"purchaseDate"`
	AccountId        string                         `json:"accountId"`
	PaymentAccountId string                         `json:"paymentAccountId"`
	CategoryId       string                         `json:"categoryId"`
	TagIds           []string                       `json:"tagIds"`
	Comment          string                         `json:"comment"`
	Status           PayLaterPlanStatus             `json:"status"`
	Installments     []*PayLaterInstallmentResponse `json:"installments,omitempty"`
}
