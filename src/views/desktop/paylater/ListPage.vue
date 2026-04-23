<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Pay Later Plans') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating" @click="showCreateDialog = true">{{ tt('Add') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :disabled="loading || updating"
                               :loading="loading" @click="reload">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                    </div>
                </template>

                <v-table class="pay-later-plans-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>{{ tt('Plan Name') }}</th>
                        <th>{{ tt('Total Amount') }}</th>
                        <th>{{ tt('Progress') }}</th>
                        <th>{{ tt('Status') }}</th>
                        <th>{{ tt('Operation') }}</th>
                    </tr>
                    </thead>

                    <tbody v-if="loading && !plans.length">
                    <tr :key="idx" v-for="idx in [1, 2, 3]">
                        <td colspan="5" class="px-0">
                            <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && !plans.length">
                    <tr>
                        <td colspan="5">{{ tt('No available pay later plans') }}</td>
                    </tr>
                    </tbody>

                    <tbody>
                    <tr v-for="plan in plans" :key="plan.id" class="text-sm cursor-pointer"
                        @click="viewDetail(plan)">
                        <td>{{ plan.name }}</td>
                        <td>{{ plan.totalAmount }}</td>
                        <td>
                            <v-progress-linear
                                :model-value="(plan.paidInstallmentsCount / plan.installmentCount) * 100"
                                color="blue"
                                height="8"
                                rounded
                            ></v-progress-linear>
                            <small>{{ plan.paidInstallmentsCount }} / {{ plan.installmentCount }}</small>
                        </td>
                        <td>
                            <v-chip :color="getPlanStatusColor(plan)" size="small">
                                {{ getPlanStatusLabel(plan) }}
                            </v-chip>
                        </td>
                        <td @click.stop="">
                            <v-btn density="comfortable" variant="text" color="error"
                                   :prepend-icon="mdiDeleteOutline"
                                   :loading="planRemoving[plan.id]"
                                   :disabled="loading || updating"
                                   @click="remove(plan)">
                                <template #loader>
                                    <v-progress-circular indeterminate size="20" width="2"/>
                                </template>
                                {{ tt('Delete') }}
                            </v-btn>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </v-card>
        </v-col>
    </v-row>

    <!-- Create Dialog -->
    <v-dialog v-model="showCreateDialog" max-width="560">
        <v-card :title="tt('Add Pay Later Plan')">
            <v-card-text>
                <v-text-field :label="tt('Plan Name')" v-model="newPlan.name" density="compact" variant="outlined" class="mb-3"/>
                <v-text-field :label="tt('Total Amount')" v-model.number="newPlan.totalAmount" type="number" min="1" density="compact" variant="outlined" class="mb-3"/>
                <v-text-field :label="tt('Installment Count (1-12)')" v-model.number="newPlan.installmentCount" type="number" min="1" max="12" density="compact" variant="outlined" class="mb-3"/>
                <v-text-field :label="tt('Purchase Date')" v-model="newPlan.purchaseDateString" type="date" density="compact" variant="outlined" class="mb-3"/>
                <v-text-field :label="tt('Comment')" v-model="newPlan.comment" density="compact" variant="outlined"/>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn @click="showCreateDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" :disabled="createInputIsEmpty || creating" :loading="creating" @click="createPlan">{{ tt('Save') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Detail Dialog -->
    <v-dialog v-model="showDetailDialog" max-width="700">
        <v-card :title="selectedPlan ? selectedPlan.name : tt('Plan Detail')">
            <v-card-text v-if="selectedPlan">
                <v-list lines="two" v-if="selectedPlan.installments">
                    <v-list-item
                        v-for="inst in selectedPlan.installments"
                        :key="inst.id"
                        :title="`#${inst.installmentNumber} — ${formatDate(inst.dueDate)}`"
                        :subtitle="`${inst.amount} — ${getInstallmentStatusLabel(inst)}`"
                    >
                        <template #append>
                            <v-btn v-if="inst.status !== PayLaterInstallmentStatus.Paid"
                                   color="primary" size="small" variant="tonal"
                                   :loading="paying[inst.id]"
                                   @click="payInstallment(inst)">
                                {{ tt('Pay Now') }}
                            </v-btn>
                            <v-icon v-else color="success" :icon="mdiCheckCircleOutline"/>
                        </template>
                    </v-list-item>
                </v-list>
                <div v-else class="text-center pa-4">
                    <v-progress-circular indeterminate/>
                </div>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn @click="showDetailDialog = false">{{ tt('Close') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar"/>
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { usePayLaterStore } from '@/stores/payLater.ts';
import { PayLaterPlan, PayLaterPlanStatus, PayLaterInstallmentStatus, type PayLaterInstallmentResponse } from '@/models/paylater_plan.ts';
import { getCurrentUnixTime, getTimezoneOffsetMinutes } from '@/lib/datetime.ts';
import { generateRandomUUID } from '@/lib/misc.ts';

import {
    mdiRefresh,
    mdiDeleteOutline,
    mdiCheckCircleOutline
} from '@mdi/js';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const payLaterStore = usePayLaterStore();

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);
const creating = ref<boolean>(false);
const showCreateDialog = ref<boolean>(false);
const showDetailDialog = ref<boolean>(false);
const selectedPlan = ref<PayLaterPlan | null>(null);
const planRemoving = ref<Record<string, boolean>>({});
const paying = ref<Record<string, boolean>>({});

const newPlan = ref({
    name: '',
    totalAmount: 0,
    installmentCount: 1,
    purchaseDateString: new Date().toISOString().substring(0, 10),
    comment: '',
    accountId: '0',
    paymentAccountId: '0',
    categoryId: '0',
    tagIds: [] as string[]
});

const plans = computed<PayLaterPlan[]>(() => payLaterStore.allPlans);
const createInputIsEmpty = computed<boolean>(() => {
    return !newPlan.value.name.trim() || newPlan.value.totalAmount <= 0 || newPlan.value.installmentCount < 1 || newPlan.value.installmentCount > 12;
});

function getPlanStatusLabel(plan: PayLaterPlan): string {
    if (plan.status === PayLaterPlanStatus.Completed) return tt('Completed');
    if (plan.status === PayLaterPlanStatus.Cancelled) return tt('Cancelled');
    return tt('Active');
}

function getPlanStatusColor(plan: PayLaterPlan): string {
    if (plan.status === PayLaterPlanStatus.Completed) return 'success';
    if (plan.status === PayLaterPlanStatus.Cancelled) return 'default';
    return 'primary';
}

function getInstallmentStatusLabel(inst: PayLaterInstallmentResponse): string {
    if (inst.status === PayLaterInstallmentStatus.Paid) return tt('Paid');
    if (inst.status === PayLaterInstallmentStatus.Overdue) return tt('Overdue');
    return tt('Pending');
}

function formatDate(unixTime: number): string {
    return new Date(unixTime * 1000).toLocaleDateString();
}

function viewDetail(plan: PayLaterPlan): void {
    selectedPlan.value = null;
    showDetailDialog.value = true;

    payLaterStore.getPlan({ planId: plan.id }).then(result => {
        selectedPlan.value = result;
    }).catch(error => {
        showDetailDialog.value = false;
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function createPlan(): void {
    if (createInputIsEmpty.value) return;

    creating.value = true;
    const purchaseDate = Math.floor(new Date(newPlan.value.purchaseDateString).getTime() / 1000);

    payLaterStore.createPlan({
        name: newPlan.value.name.trim(),
        totalAmount: Number(newPlan.value.totalAmount),
        installmentCount: Number(newPlan.value.installmentCount),
        purchaseDate: purchaseDate,
        accountId: newPlan.value.accountId,
        paymentAccountId: newPlan.value.paymentAccountId,
        categoryId: newPlan.value.categoryId,
        tagIds: newPlan.value.tagIds,
        comment: newPlan.value.comment,
        clientSessionId: generateRandomUUID()
    }).then(() => {
        creating.value = false;
        showCreateDialog.value = false;
        resetNewPlan();
        snackbar.value?.showMessage('Pay later plan has been created');
    }).catch(error => {
        creating.value = false;
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function resetNewPlan(): void {
    newPlan.value = {
        name: '',
        totalAmount: 0,
        installmentCount: 1,
        purchaseDateString: new Date().toISOString().substring(0, 10),
        comment: '',
        accountId: '0',
        paymentAccountId: '0',
        categoryId: '0',
        tagIds: []
    };
}

function payInstallment(inst: PayLaterInstallmentResponse): void {
    paying.value[inst.id] = true;
    const now = getCurrentUnixTime();
    const utcOffset = getTimezoneOffsetMinutes(now);

    payLaterStore.payInstallment({
        id: inst.id,
        paymentTime: now,
        utcOffset: utcOffset
    }).then(() => {
        paying.value[inst.id] = false;
        snackbar.value?.showMessage('Installment has been paid');
        // Refresh detail
        if (selectedPlan.value) {
            payLaterStore.getPlan({ planId: selectedPlan.value.id }).then(result => {
                selectedPlan.value = result;
            });
        }
    }).catch(error => {
        paying.value[inst.id] = false;
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(plan: PayLaterPlan): void {
    confirmDialog.value?.open('Are you sure you want to delete this plan?').then(() => {
        updating.value = true;
        planRemoving.value[plan.id] = true;

        payLaterStore.deletePlan({ id: plan.id }).then(() => {
            updating.value = false;
            planRemoving.value[plan.id] = false;
            snackbar.value?.showMessage('Pay later plan has been deleted');
        }).catch(error => {
            updating.value = false;
            planRemoving.value[plan.id] = false;
            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function reload(): void {
    loading.value = true;

    payLaterStore.loadAllPlans({ force: true }).then(() => {
        loading.value = false;
        snackbar.value?.showMessage('Pay later plan list has been updated');
    }).catch(error => {
        loading.value = false;
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function init(): void {
    loading.value = true;

    payLaterStore.loadAllPlans({ force: false }).then(() => {
        loading.value = false;
    }).catch(error => {
        loading.value = false;
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

init();
</script>
