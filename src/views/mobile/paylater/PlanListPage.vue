<template>
    <f7-page ptr @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt('Pay Later Plans')"></f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="plus" href="/paylater/plans/add"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-top skeleton-text" v-if="loading">
            <f7-list-item title="Plan Name"
                          :key="itemIdx" v-for="itemIdx in [1, 2, 3]">
                <template #media>
                    <f7-icon f7="creditcard"></f7-icon>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-top" v-if="!loading && !plans.length">
            <f7-list-item :title="tt('No available pay later plans')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-top" v-if="!loading && plans.length">
            <f7-list-item swipeout
                          :title="plan.name"
                          :footer="getPlanProgress(plan)"
                          :key="plan.id"
                          v-for="plan in plans"
                          @click="viewDetail(plan)">
                <template #media>
                    <f7-icon f7="creditcard" :color="getPlanStatusColor(plan)"></f7-icon>
                </template>
                <template #after>
                    <span :class="'text-' + getPlanStatusColor(plan)">{{ getPlanStatusLabel(plan) }}</span>
                </template>
                <f7-swipeout-actions right>
                    <f7-swipeout-button color="red" class="padding-horizontal" @click="remove(plan, false)">
                        <f7-icon f7="trash"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>
        </f7-list>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this plan?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="remove(planToDelete, true)">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import { usePayLaterStore } from '@/stores/payLater.ts';
import { PayLaterPlan, PayLaterPlanStatus } from '@/models/paylater_plan.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showToast, routeBackOnError } = useI18nUIComponents();

const payLaterStore = usePayLaterStore();

const loading = ref<boolean>(true);
const loadingError = ref<unknown | null>(null);
const planToDelete = ref<PayLaterPlan | null>(null);
const showDeleteActionSheet = ref<boolean>(false);

const plans = computed<PayLaterPlan[]>(() => payLaterStore.allPlans);

function getPlanStatusLabel(plan: PayLaterPlan): string {
    if (plan.status === PayLaterPlanStatus.Completed) return tt('Completed');
    if (plan.status === PayLaterPlanStatus.Cancelled) return tt('Cancelled');
    return tt('Active');
}

function getPlanStatusColor(plan: PayLaterPlan): string {
    if (plan.status === PayLaterPlanStatus.Completed) return 'green';
    if (plan.status === PayLaterPlanStatus.Cancelled) return 'gray';
    return 'blue';
}

function getPlanProgress(plan: PayLaterPlan): string {
    if (!plan.installments) {
        return `${plan.installmentCount} ${tt('installments')}`;
    }
    const paid = plan.paidInstallmentsCount;
    return `${paid} / ${plan.installmentCount} ${tt('installments paid')}`;
}

function viewDetail(plan: PayLaterPlan): void {
    props.f7router.navigate(`/paylater/plans/detail?id=${plan.id}`);
}

function remove(plan: PayLaterPlan | null, confirm: boolean): void {
    if (!plan) return;

    if (!confirm) {
        planToDelete.value = plan;
        showDeleteActionSheet.value = true;
        return;
    }

    showDeleteActionSheet.value = false;
    planToDelete.value = null;
    showLoading();

    payLaterStore.deletePlan({ id: plan.id }).then(() => {
        hideLoading();
    }).catch(error => {
        hideLoading();
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function init(): void {
    loading.value = true;

    payLaterStore.loadAllPlans({ force: false }).then(() => {
        loading.value = false;
    }).catch(error => {
        if (error.processed) {
            loading.value = false;
        } else {
            loadingError.value = error;
            showToast(error.message || error);
        }
    });
}

function reload(done?: () => void): void {
    payLaterStore.loadAllPlans({ force: true }).then(() => {
        done?.();
    }).catch(error => {
        done?.();
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function onPageAfterIn(): void {
    routeBackOnError(props.f7router, loadingError);
}

init();
</script>
