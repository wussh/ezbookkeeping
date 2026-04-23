<template>
    <f7-page ptr @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="plan ? plan.name : tt('Plan Detail')"></f7-nav-title>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-top skeleton-text" v-if="loading">
            <f7-list-item title="Installment" :key="idx" v-for="idx in [1, 2, 3]"></f7-list-item>
        </f7-list>

        <template v-if="!loading && plan">
            <f7-block-title>{{ tt('Plan Info') }}</f7-block-title>
            <f7-list strong inset dividers class="margin-top">
                <f7-list-item :title="tt('Total Amount')" :after="String(plan.totalAmount)"></f7-list-item>
                <f7-list-item :title="tt('Installment Count')" :after="String(plan.installmentCount)"></f7-list-item>
                <f7-list-item :title="tt('Status')" :after="getPlanStatusLabel(plan)"></f7-list-item>
                <f7-list-item v-if="plan.comment" :title="tt('Comment')" :after="plan.comment"></f7-list-item>
            </f7-list>

            <f7-block-title>{{ tt('Installment Schedule') }}</f7-block-title>
            <f7-list strong inset dividers class="margin-top">
                <f7-list-item
                    v-for="inst in plan.installments"
                    :key="inst.id"
                    :title="`#${inst.installmentNumber} — ${formatDate(inst.dueDate)}`"
                    :after="String(inst.amount)"
                    :footer="getInstallmentStatusLabel(inst)"
                >
                    <template #after>
                        <div class="d-flex align-items-center gap-8">
                            <span>{{ inst.amount }}</span>
                            <f7-button
                                v-if="inst.status !== PayLaterInstallmentStatus.Paid"
                                small fill color="blue"
                                :disabled="paying[inst.id]"
                                @click.stop="payInstallment(inst)"
                            >{{ tt('Pay Now') }}</f7-button>
                            <f7-icon v-else f7="checkmark_circle_fill" color="green"></f7-icon>
                        </div>
                    </template>
                </f7-list-item>
            </f7-list>
        </template>

        <f7-actions close-by-outside-click close-on-escape :opened="showPayConfirm" @actions:closed="showPayConfirm = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to mark this installment as paid?') }}</f7-actions-label>
                <f7-actions-button color="blue" @click="confirmPay">{{ tt('Pay Now') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import { usePayLaterStore } from '@/stores/payLater.ts';
import { PayLaterPlan, PayLaterPlanStatus, PayLaterInstallmentStatus, type PayLaterInstallmentResponse } from '@/models/paylater_plan.ts';
import { getCurrentUnixTime, getTimezoneOffsetMinutes } from '@/lib/datetime.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showToast, routeBackOnError } = useI18nUIComponents();
const payLaterStore = usePayLaterStore();

const plan = ref<PayLaterPlan | null>(null);
const loading = ref<boolean>(true);
const loadingError = ref<unknown | null>(null);
const paying = ref<Record<string, boolean>>({});
const selectedInstallment = ref<PayLaterInstallmentResponse | null>(null);
const showPayConfirm = ref<boolean>(false);

function getPlanStatusLabel(p: PayLaterPlan): string {
    if (p.status === PayLaterPlanStatus.Completed) return tt('Completed');
    if (p.status === PayLaterPlanStatus.Cancelled) return tt('Cancelled');
    return tt('Active');
}

function getInstallmentStatusLabel(inst: PayLaterInstallmentResponse): string {
    if (inst.status === PayLaterInstallmentStatus.Paid) return tt('Paid');
    if (inst.status === PayLaterInstallmentStatus.Overdue) return tt('Overdue');
    return tt('Pending');
}

function formatDate(unixTime: number): string {
    return new Date(unixTime * 1000).toLocaleDateString();
}

function payInstallment(inst: PayLaterInstallmentResponse): void {
    selectedInstallment.value = inst;
    showPayConfirm.value = true;
}

function confirmPay(): void {
    showPayConfirm.value = false;
    const inst = selectedInstallment.value;
    if (!inst) return;

    paying.value[inst.id] = true;
    showLoading();

    const now = getCurrentUnixTime();
    const utcOffset = getTimezoneOffsetMinutes(now);

    payLaterStore.payInstallment({
        id: inst.id,
        paymentTime: now,
        utcOffset: utcOffset
    }).then(() => {
        paying.value[inst.id] = false;
        hideLoading();
        loadPlan(true);
    }).catch(error => {
        paying.value[inst.id] = false;
        hideLoading();
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function loadPlan(force = false): void {
    const planId = props.f7route.query?.['id'] as string;
    if (!planId) {
        loadingError.value = { message: 'Plan ID is missing' };
        return;
    }

    loading.value = !force;

    payLaterStore.getPlan({ planId }).then(result => {
        plan.value = result;
        loading.value = false;
    }).catch(error => {
        loading.value = false;
        if (error.processed) {
            return;
        }
        loadingError.value = error;
        showToast(error.message || error);
    });
}

function reload(done?: () => void): void {
    loadPlan(true);
    done?.();
}

function onPageAfterIn(): void {
    routeBackOnError(props.f7router, loadingError);
}

loadPlan();
</script>
