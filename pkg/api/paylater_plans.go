package api

import (
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// PayLaterPlansApi represents the pay later plan api
type PayLaterPlansApi struct {
	ApiUsingConfig
	payLaterPlans *services.PayLaterPlanService
}

// Initialize a pay later plan api singleton instance
var (
	PayLaterPlans = &PayLaterPlansApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		payLaterPlans: services.PayLaterPlans,
	}
)

// PayLaterPlanListHandler returns pay later plan list of current user
func (a *PayLaterPlansApi) PayLaterPlanListHandler(c *core.WebContext) (any, *errs.Error) {
	var listReq models.PayLaterInstallmentPlanListRequest
	err := c.ShouldBindQuery(&listReq)

	if err != nil {
		log.Warnf(c, "[paylater_plans.PayLaterPlanListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	plans, err := a.payLaterPlans.GetAllPlansByUid(c, uid, listReq.Status)

	if err != nil {
		log.Errorf(c, "[paylater_plans.PayLaterPlanListHandler] failed to get plans for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	planResps := make([]*models.PayLaterInstallmentPlanInfoResponse, len(plans))

	for i, plan := range plans {
		planResps[i] = toPayLaterPlanInfoResponse(plan, nil)
	}

	return planResps, nil
}

// PayLaterPlanGetHandler returns one specific pay later plan of current user
func (a *PayLaterPlansApi) PayLaterPlanGetHandler(c *core.WebContext) (any, *errs.Error) {
	var getReq models.PayLaterInstallmentPlanGetRequest
	err := c.ShouldBindQuery(&getReq)

	if err != nil {
		log.Warnf(c, "[paylater_plans.PayLaterPlanGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	plan, err := a.payLaterPlans.GetPlanByPlanId(c, uid, getReq.Id)

	if err != nil {
		log.Errorf(c, "[paylater_plans.PayLaterPlanGetHandler] failed to get plan \"id:%d\" for user \"uid:%d\", because %s", getReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	installments, err := a.payLaterPlans.GetInstallmentsByPlanId(c, uid, plan.PlanId)

	if err != nil {
		log.Errorf(c, "[paylater_plans.PayLaterPlanGetHandler] failed to get installments for plan \"id:%d\", because %s", plan.PlanId, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return toPayLaterPlanInfoResponse(plan, installments), nil
}

// PayLaterPlanCreateHandler creates a new pay later installment plan
func (a *PayLaterPlansApi) PayLaterPlanCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var createReq models.PayLaterInstallmentPlanCreateRequest
	err := c.ShouldBindJSON(&createReq)

	if err != nil {
		log.Warnf(c, "[paylater_plans.PayLaterPlanCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	plan := &models.PayLaterInstallmentPlan{
		Uid:              uid,
		Name:             createReq.Name,
		TotalAmount:      createReq.TotalAmount,
		InstallmentCount: createReq.InstallmentCount,
		PurchaseDate:     createReq.PurchaseDate,
		AccountId:        createReq.AccountId,
		PaymentAccountId: createReq.PaymentAccountId,
		CategoryId:       createReq.CategoryId,
		TagIds:           strings.Join(createReq.TagIds, ","),
		Comment:          createReq.Comment,
	}

	installments, err := a.payLaterPlans.CreatePlan(c, plan)

	if err != nil {
		log.Errorf(c, "[paylater_plans.PayLaterPlanCreateHandler] failed to create plan for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[paylater_plans.PayLaterPlanCreateHandler] user \"uid:%d\" has created pay later plan \"id:%d\"", uid, plan.PlanId)
	return toPayLaterPlanInfoResponse(plan, installments), nil
}

// PayLaterInstallmentPayHandler marks an installment as paid
func (a *PayLaterPlansApi) PayLaterInstallmentPayHandler(c *core.WebContext) (any, *errs.Error) {
	var payReq models.PayLaterInstallmentPayRequest
	err := c.ShouldBindJSON(&payReq)

	if err != nil {
		log.Warnf(c, "[paylater_plans.PayLaterInstallmentPayHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	_, err = a.payLaterPlans.MarkInstallmentPaid(c, uid, payReq.Id, payReq.PaymentTime, payReq.UtcOffset)

	if err != nil {
		log.Errorf(c, "[paylater_plans.PayLaterInstallmentPayHandler] failed to mark installment \"id:%d\" as paid for user \"uid:%d\", because %s", payReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[paylater_plans.PayLaterInstallmentPayHandler] user \"uid:%d\" has paid installment \"id:%d\"", uid, payReq.Id)
	return true, nil
}

// PayLaterPlanDeleteHandler soft-deletes a pay later plan
func (a *PayLaterPlansApi) PayLaterPlanDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var deleteReq models.PayLaterInstallmentPlanDeleteRequest
	err := c.ShouldBindJSON(&deleteReq)

	if err != nil {
		log.Warnf(c, "[paylater_plans.PayLaterPlanDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	err = a.payLaterPlans.DeletePlan(c, uid, deleteReq.Id)

	if err != nil {
		log.Errorf(c, "[paylater_plans.PayLaterPlanDeleteHandler] failed to delete plan \"id:%d\" for user \"uid:%d\", because %s", deleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[paylater_plans.PayLaterPlanDeleteHandler] user \"uid:%d\" has deleted pay later plan \"id:%d\"", uid, deleteReq.Id)
	return true, nil
}

func toPayLaterPlanInfoResponse(plan *models.PayLaterInstallmentPlan, installments []*models.PayLaterInstallment) *models.PayLaterInstallmentPlanInfoResponse {
	tagIds := []string{}
	if plan.TagIds != "" {
		tagIds = strings.Split(plan.TagIds, ",")
	}

	resp := &models.PayLaterInstallmentPlanInfoResponse{
		Id:               utils.Int64ToString(plan.PlanId),
		Name:             plan.Name,
		TotalAmount:      plan.TotalAmount,
		InstallmentCount: plan.InstallmentCount,
		PurchaseDate:     plan.PurchaseDate,
		AccountId:        utils.Int64ToString(plan.AccountId),
		PaymentAccountId: utils.Int64ToString(plan.PaymentAccountId),
		CategoryId:       utils.Int64ToString(plan.CategoryId),
		TagIds:           tagIds,
		Comment:          plan.Comment,
		Status:           plan.Status,
	}

	if installments != nil {
		resp.Installments = make([]*models.PayLaterInstallmentResponse, len(installments))

		for i, inst := range installments {
			paidTxId := ""
			if inst.PaidTransactionId > 0 {
				paidTxId = utils.Int64ToString(inst.PaidTransactionId)
			}

			resp.Installments[i] = &models.PayLaterInstallmentResponse{
				Id:                utils.Int64ToString(inst.InstallmentId),
				PlanId:            utils.Int64ToString(inst.PlanId),
				InstallmentNumber: inst.InstallmentNumber,
				DueDate:           inst.DueDate,
				Amount:            inst.Amount,
				PaidTransactionId: paidTxId,
				Status:            inst.Status,
			}
		}
	}

	return resp
}
