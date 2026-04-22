import { ref } from 'vue';
import { defineStore } from 'pinia';

import {
    type PayLaterInstallmentPlanInfoResponse,
    type PayLaterPlanCreateRequest,
    type PayLaterInstallmentPayRequest,
    type PayLaterPlanDeleteRequest,
    PayLaterPlan
} from '@/models/paylater_plan.ts';

import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';
import { getCurrentUnixTime, getTimezoneOffsetMinutes } from '@/lib/datetime.ts';

export const usePayLaterStore = defineStore('payLater', () => {
    const allPlans = ref<PayLaterPlan[]>([]);
    const plansLoaded = ref(false);

    function loadPlanList(plans: PayLaterInstallmentPlanInfoResponse[]): void {
        allPlans.value = plans.map(PayLaterPlan.of);
        plansLoaded.value = true;
    }

    function addPlanToList(plan: PayLaterInstallmentPlanInfoResponse): void {
        allPlans.value.unshift(PayLaterPlan.of(plan));
    }

    function removePlanFromList(planId: string): void {
        const index = allPlans.value.findIndex(p => p.id === planId);
        if (index >= 0) {
            allPlans.value.splice(index, 1);
        }
    }

    function resetPayLaterPlans(): void {
        allPlans.value = [];
        plansLoaded.value = false;
    }

    function loadAllPlans({ force }: { force?: boolean } = {}): Promise<PayLaterPlan[]> {
        if (!force && plansLoaded.value) {
            return Promise.resolve(allPlans.value);
        }

        return new Promise((resolve, reject) => {
            services.getAllPayLaterPlans({}).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve pay later plan list' });
                    return;
                }

                loadPlanList(data.result);
                resolve(allPlans.value);
            }).catch(error => {
                logger.error('failed to load pay later plan list', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve pay later plan list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function getPlan({ planId }: { planId: string }): Promise<PayLaterPlan> {
        return new Promise((resolve, reject) => {
            services.getPayLaterPlan({ id: planId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve pay later plan' });
                    return;
                }

                resolve(PayLaterPlan.of(data.result));
            }).catch(error => {
                logger.error('failed to get pay later plan', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve pay later plan' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function createPlan(req: PayLaterPlanCreateRequest): Promise<PayLaterPlan> {
        return new Promise((resolve, reject) => {
            services.addPayLaterPlan(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to create pay later plan' });
                    return;
                }

                addPlanToList(data.result);
                resolve(PayLaterPlan.of(data.result));
            }).catch(error => {
                logger.error('failed to create pay later plan', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to create pay later plan' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function payInstallment(req: PayLaterInstallmentPayRequest): Promise<void> {
        return new Promise((resolve, reject) => {
            services.payPayLaterInstallment(req).then(response => {
                const data = response.data;

                if (!data || !data.success) {
                    reject({ message: 'Unable to pay installment' });
                    return;
                }

                // Invalidate cache so next load picks up latest state
                plansLoaded.value = false;
                resolve();
            }).catch(error => {
                logger.error('failed to pay pay later installment', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to pay installment' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deletePlan(req: PayLaterPlanDeleteRequest): Promise<void> {
        return new Promise((resolve, reject) => {
            services.deletePayLaterPlan(req).then(response => {
                const data = response.data;

                if (!data || !data.success) {
                    reject({ message: 'Unable to delete pay later plan' });
                    return;
                }

                removePlanFromList(req.id);
                resolve();
            }).catch(error => {
                logger.error('failed to delete pay later plan', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete pay later plan' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function getCurrentUtcOffset(): number {
        return getTimezoneOffsetMinutes(getCurrentUnixTime());
    }

    return {
        allPlans,
        plansLoaded,
        loadAllPlans,
        getPlan,
        createPlan,
        payInstallment,
        deletePlan,
        resetPayLaterPlans,
        getCurrentUtcOffset
    };
});
