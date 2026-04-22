package errs

import "net/http"

// Error codes related to pay later installment plans
var (
	ErrPayLaterPlanIdInvalid          = NewNormalError(NormalSubcategoryPayLater, 0, http.StatusBadRequest, "pay later plan id is invalid")
	ErrPayLaterPlanNotFound           = NewNormalError(NormalSubcategoryPayLater, 1, http.StatusBadRequest, "pay later plan not found")
	ErrPayLaterInstallmentIdInvalid   = NewNormalError(NormalSubcategoryPayLater, 2, http.StatusBadRequest, "pay later installment id is invalid")
	ErrPayLaterInstallmentNotFound    = NewNormalError(NormalSubcategoryPayLater, 3, http.StatusBadRequest, "pay later installment not found")
	ErrPayLaterInstallmentAlreadyPaid = NewNormalError(NormalSubcategoryPayLater, 4, http.StatusBadRequest, "pay later installment is already paid")
	ErrPayLaterPlanAlreadyCompleted   = NewNormalError(NormalSubcategoryPayLater, 5, http.StatusBadRequest, "pay later plan is already completed")
	ErrPayLaterPlanAlreadyCancelled   = NewNormalError(NormalSubcategoryPayLater, 6, http.StatusBadRequest, "pay later plan is already cancelled")
	ErrPayLaterInstallmentCountInvalid = NewNormalError(NormalSubcategoryPayLater, 7, http.StatusBadRequest, "installment count must be between 1 and 12")
)
