package services

import (
	"strings"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// PayLaterPlanService represents the pay later installment plan service
type PayLaterPlanService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a pay later plan service singleton instance
var (
	PayLaterPlans = &PayLaterPlanService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllPlansByUid returns all pay later plans of user
func (s *PayLaterPlanService) GetAllPlansByUid(c core.Context, uid int64, status models.PayLaterPlanStatus) ([]*models.PayLaterInstallmentPlan, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var plans []*models.PayLaterInstallmentPlan
	sess := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false)

	if status > 0 {
		sess = sess.And("status=?", status)
	}

	err := sess.OrderBy("created_unix_time desc").Find(&plans)

	return plans, err
}

// GetPlanByPlanId returns a pay later plan model according to plan id
func (s *PayLaterPlanService) GetPlanByPlanId(c core.Context, uid int64, planId int64) (*models.PayLaterInstallmentPlan, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if planId <= 0 {
		return nil, errs.ErrPayLaterPlanIdInvalid
	}

	plan := &models.PayLaterInstallmentPlan{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND plan_id=?", uid, false, planId).Get(plan)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrPayLaterPlanNotFound
	}

	return plan, nil
}

// GetInstallmentsByPlanId returns all installments for a plan
func (s *PayLaterPlanService) GetInstallmentsByPlanId(c core.Context, uid int64, planId int64) ([]*models.PayLaterInstallment, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if planId <= 0 {
		return nil, errs.ErrPayLaterPlanIdInvalid
	}

	var installments []*models.PayLaterInstallment
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND plan_id=?", uid, planId).OrderBy("installment_number asc").Find(&installments)

	return installments, err
}

// CreatePlan creates a new pay later installment plan along with its installments
func (s *PayLaterPlanService) CreatePlan(c core.Context, plan *models.PayLaterInstallmentPlan) ([]*models.PayLaterInstallment, error) {
	if plan.Uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if plan.InstallmentCount < 1 || plan.InstallmentCount > 12 {
		return nil, errs.ErrPayLaterInstallmentCountInvalid
	}

	now := time.Now().Unix()

	planUuid := s.GenerateUuid(uuid.UUID_TYPE_PAY_LATER)
	if planUuid < 1 {
		return nil, errs.ErrSystemIsBusy
	}

	installmentUuids := s.GenerateUuids(uuid.UUID_TYPE_PAY_LATER, uint16(plan.InstallmentCount))
	if len(installmentUuids) < int(plan.InstallmentCount) {
		return nil, errs.ErrSystemIsBusy
	}

	plan.PlanId = planUuid
	plan.Status = models.PAY_LATER_PLAN_STATUS_ACTIVE
	plan.Deleted = false
	plan.CreatedUnixTime = now
	plan.UpdatedUnixTime = now

	// Calculate per-installment amount; remainder goes to first installment
	baseAmount := plan.TotalAmount / int64(plan.InstallmentCount)
	remainder := plan.TotalAmount % int64(plan.InstallmentCount)

	// Parse purchase date into time.Time to add months
	purchaseTime := time.Unix(plan.PurchaseDate, 0).UTC()

	installments := make([]*models.PayLaterInstallment, plan.InstallmentCount)

	for i := int32(0); i < plan.InstallmentCount; i++ {
		amount := baseAmount
		if i == 0 {
			amount += remainder
		}

		dueTime := purchaseTime.AddDate(0, int(i+1), 0)

		installments[i] = &models.PayLaterInstallment{
			InstallmentId:     installmentUuids[i],
			PlanId:            plan.PlanId,
			Uid:               plan.Uid,
			InstallmentNumber: i + 1,
			DueDate:           dueTime.Unix(),
			Amount:            amount,
			PaidTransactionId: 0,
			Status:            models.PAY_LATER_INSTALLMENT_STATUS_PENDING,
			CreatedUnixTime:   now,
			UpdatedUnixTime:   now,
		}
	}

	return installments, s.UserDataDB(plan.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(plan)
		if err != nil {
			return err
		}

		for _, inst := range installments {
			_, err = sess.Insert(inst)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// MarkInstallmentPaid marks a single installment as paid and creates the underlying expense transaction
func (s *PayLaterPlanService) MarkInstallmentPaid(c core.Context, uid int64, installmentId int64, paymentTime int64, utcOffset int16) (*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if installmentId <= 0 {
		return nil, errs.ErrPayLaterInstallmentIdInvalid
	}

	// Load installment
	installment := &models.PayLaterInstallment{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND installment_id=?", uid, installmentId).Get(installment)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrPayLaterInstallmentNotFound
	}

	if installment.Status == models.PAY_LATER_INSTALLMENT_STATUS_PAID {
		return nil, errs.ErrPayLaterInstallmentAlreadyPaid
	}

	// Load plan to get category, tags, accounts
	plan := &models.PayLaterInstallmentPlan{}
	has, err = s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND plan_id=?", uid, false, installment.PlanId).Get(plan)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrPayLaterPlanNotFound
	}

	if plan.Status == models.PAY_LATER_PLAN_STATUS_CANCELLED {
		return nil, errs.ErrPayLaterPlanAlreadyCancelled
	}

	// Generate transaction UUIDs
	transactionUuids := s.GenerateUuids(uuid.UUID_TYPE_TRANSACTION, 2)
	if len(transactionUuids) < 2 {
		return nil, errs.ErrSystemIsBusy
	}

	// Build tag index UUIDs
	var tagIds []int64
	if plan.TagIds != "" {
		for _, tagIdStr := range strings.Split(plan.TagIds, ",") {
			tagId, convErr := utils.StringToInt64(tagIdStr)
			if convErr == nil && tagId > 0 {
				tagIds = append(tagIds, tagId)
			}
		}
	}

	tagIndexUuids := s.GenerateUuids(uuid.UUID_TYPE_TAG_INDEX, uint16(len(tagIds)))
	if len(tagIndexUuids) < len(tagIds) {
		return nil, errs.ErrSystemIsBusy
	}

	now := time.Now().Unix()
	transactionTime := utils.GetMinTransactionTimeFromUnixTime(paymentTime)

	// Expense from PaymentAccountId (money leaving the payment account)
	expenseTx := &models.Transaction{
		TransactionId:        transactionUuids[0],
		Uid:                  uid,
		Deleted:              false,
		Type:                 models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
		CategoryId:           plan.CategoryId,
		AccountId:            plan.PaymentAccountId,
		TransactionTime:      transactionTime,
		TimezoneUtcOffset:    utcOffset,
		Amount:               installment.Amount,
		RelatedId:            transactionUuids[1],
		RelatedAccountId:     plan.AccountId,
		RelatedAccountAmount: installment.Amount,
		HideAmount:           false,
		Comment:              plan.Name,
		CreatedUnixTime:      now,
		UpdatedUnixTime:      now,
	}

	// Paired transfer-in to the pay later account (reducing its liability balance)
	incomeTx := &models.Transaction{
		TransactionId:        transactionUuids[1],
		Uid:                  uid,
		Deleted:              false,
		Type:                 models.TRANSACTION_DB_TYPE_TRANSFER_IN,
		CategoryId:           plan.CategoryId,
		AccountId:            plan.AccountId,
		TransactionTime:      transactionTime + 1,
		TimezoneUtcOffset:    utcOffset,
		Amount:               installment.Amount,
		RelatedId:            transactionUuids[0],
		RelatedAccountId:     plan.PaymentAccountId,
		RelatedAccountAmount: installment.Amount,
		HideAmount:           false,
		Comment:              plan.Name,
		CreatedUnixTime:      now,
		UpdatedUnixTime:      now,
	}

	transactionTagIndexes := make([]*models.TransactionTagIndex, len(tagIds))
	for i, tagId := range tagIds {
		transactionTagIndexes[i] = &models.TransactionTagIndex{
			TagIndexId:      tagIndexUuids[i],
			Uid:             uid,
			Deleted:         false,
			TagId:           tagId,
			TransactionId:   expenseTx.TransactionId,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}
	}

	userDataDb := s.UserDataDB(uid)

	err = userDataDb.DoTransaction(c, func(sess *xorm.Session) error {
		// Insert both transactions
		_, err2 := sess.Insert(expenseTx)
		if err2 != nil {
			return err2
		}

		_, err2 = sess.Insert(incomeTx)
		if err2 != nil {
			return err2
		}

		// Insert tag indexes for the expense transaction
		for _, tagIndex := range transactionTagIndexes {
			_, err2 = sess.Insert(tagIndex)
			if err2 != nil {
				return err2
			}
		}

		// Update account balances
		_, err2 = sess.Exec("UPDATE accounts SET balance=balance-? WHERE uid=? AND account_id=?",
			installment.Amount, uid, plan.PaymentAccountId)
		if err2 != nil {
			return err2
		}

		_, err2 = sess.Exec("UPDATE accounts SET balance=balance-? WHERE uid=? AND account_id=?",
			installment.Amount, uid, plan.AccountId)
		if err2 != nil {
			return err2
		}

		// Mark installment as paid
		installment.Status = models.PAY_LATER_INSTALLMENT_STATUS_PAID
		installment.PaidTransactionId = expenseTx.TransactionId
		installment.UpdatedUnixTime = now

		_, err2 = sess.ID(installmentId).Cols("status", "paid_transaction_id", "updated_unix_time").Update(installment)
		if err2 != nil {
			return err2
		}

		// Check if all installments are paid; if so, complete the plan
		count, err2 := sess.Where("plan_id=? AND uid=? AND status!=?", plan.PlanId, uid, models.PAY_LATER_INSTALLMENT_STATUS_PAID).Count(&models.PayLaterInstallment{})
		if err2 != nil {
			return err2
		}

		if count == 0 {
			plan.Status = models.PAY_LATER_PLAN_STATUS_COMPLETED
			plan.UpdatedUnixTime = now

			_, err2 = sess.ID(plan.PlanId).Cols("status", "updated_unix_time").Update(plan)
			if err2 != nil {
				return err2
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return expenseTx, nil
}

// DeletePlan soft-deletes a pay later plan
func (s *PayLaterPlanService) DeletePlan(c core.Context, uid int64, planId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if planId <= 0 {
		return errs.ErrPayLaterPlanIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Where("uid=? AND deleted=? AND plan_id=?", uid, false, planId).
			Cols("deleted", "deleted_unix_time", "updated_unix_time").
			Update(&models.PayLaterInstallmentPlan{
				Deleted:         true,
				DeletedUnixTime: now,
				UpdatedUnixTime: now,
			})

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrPayLaterPlanNotFound
		}

		return nil
	})
}

// MarkOverdueInstallments marks all pending installments past their due date as overdue
func (s *PayLaterPlanService) MarkOverdueInstallments(c core.Context) error {
	now := time.Now().Unix()
	dbCount := s.UserDataDBCount()

	for i := 0; i < dbCount; i++ {
		db := s.UserDataDBByIndex(i)

		var pendingInstallments []*models.PayLaterInstallment
		err := db.NewSession(c).Where("status=? AND due_date<?", models.PAY_LATER_INSTALLMENT_STATUS_PENDING, now).Find(&pendingInstallments)

		if err != nil {
			log.Errorf(c, "[paylater_plans.MarkOverdueInstallments] failed to find overdue installments, because %s", err.Error())
			continue
		}

		for _, inst := range pendingInstallments {
			err = db.DoTransaction(c, func(sess *xorm.Session) error {
				inst.Status = models.PAY_LATER_INSTALLMENT_STATUS_OVERDUE
				inst.UpdatedUnixTime = now
				_, err2 := sess.ID(inst.InstallmentId).Cols("status", "updated_unix_time").Update(inst)
				return err2
			})

			if err != nil {
				log.Errorf(c, "[paylater_plans.MarkOverdueInstallments] failed to mark installment \"id:%d\" as overdue, because %s", inst.InstallmentId, err.Error())
			}
		}
	}

	return nil
}
