<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': submitting }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt('Add Pay Later Plan')"></f7-nav-title>
            <f7-nav-right :class="{ 'disabled': inputIsEmpty || submitting }">
                <f7-link icon-f7="checkmark_alt" @click="save"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list form strong inset dividers class="margin-vertical">
            <f7-list-input
                :label="tt('Plan Name')"
                type="text"
                :placeholder="tt('e.g. iPhone purchase')"
                v-model:value="plan.name"
                clear-button
            ></f7-list-input>

            <f7-list-input
                :label="tt('Total Amount')"
                type="number"
                :placeholder="tt('Total amount')"
                v-model:value="plan.totalAmount"
                min="1"
            ></f7-list-input>

            <f7-list-input
                :label="tt('Installment Count')"
                type="number"
                :placeholder="tt('Number of monthly installments (1-12)')"
                v-model:value="plan.installmentCount"
                min="1"
                max="12"
            ></f7-list-input>

            <f7-list-input
                :label="tt('Purchase Date')"
                type="date"
                v-model:value="purchaseDateString"
            ></f7-list-input>

            <f7-list-input
                :label="tt('Comment')"
                type="textarea"
                :placeholder="tt('Optional description')"
                v-model:value="plan.comment"
            ></f7-list-input>
        </f7-list>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import { usePayLaterStore } from '@/stores/payLater.ts';
import { generateRandomUUID } from '@/lib/misc.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showAlert } = useI18nUIComponents();
const payLaterStore = usePayLaterStore();

const plan = ref({
    name: '',
    totalAmount: 0,
    installmentCount: 1,
    comment: '',
    accountId: '0',
    paymentAccountId: '0',
    categoryId: '0',
    tagIds: [] as string[]
});

const purchaseDateString = ref<string>(new Date().toISOString().substring(0, 10));
const submitting = ref<boolean>(false);

const inputIsEmpty = computed<boolean>(() => {
    return !plan.value.name.trim() || plan.value.totalAmount <= 0 || plan.value.installmentCount < 1 || plan.value.installmentCount > 12;
});

function save(): void {
    if (inputIsEmpty.value) return;

    submitting.value = true;
    showLoading();

    const purchaseDate = Math.floor(new Date(purchaseDateString.value).getTime() / 1000);

    payLaterStore.createPlan({
        name: plan.value.name.trim(),
        totalAmount: Number(plan.value.totalAmount),
        installmentCount: Number(plan.value.installmentCount),
        purchaseDate: purchaseDate,
        accountId: plan.value.accountId,
        paymentAccountId: plan.value.paymentAccountId,
        categoryId: plan.value.categoryId,
        tagIds: plan.value.tagIds,
        comment: plan.value.comment,
        clientSessionId: generateRandomUUID()
    }).then(() => {
        submitting.value = false;
        hideLoading();
        props.f7router.back();
    }).catch(error => {
        submitting.value = false;
        hideLoading();
        if (!error.processed) {
            showAlert(error.message || error);
        }
    });
}

function onPageAfterIn(): void {
    // nothing special needed
}
</script>
