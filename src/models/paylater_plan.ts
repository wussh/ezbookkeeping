export const PayLaterPlanStatus = {
    Active: 1,
    Completed: 2,
    Cancelled: 3
} as const;

export type PayLaterPlanStatusType = typeof PayLaterPlanStatus[keyof typeof PayLaterPlanStatus];

export const PayLaterInstallmentStatus = {
    Pending: 1,
    Paid: 2,
    Overdue: 3
} as const;

export type PayLaterInstallmentStatusType = typeof PayLaterInstallmentStatus[keyof typeof PayLaterInstallmentStatus];

export interface PayLaterInstallmentResponse {
    readonly id: string;
    readonly planId: string;
    readonly installmentNumber: number;
    readonly dueDate: number;
    readonly amount: number;
    readonly paidTransactionId?: string;
    readonly status: PayLaterInstallmentStatusType;
}

export interface PayLaterInstallmentPlanInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly totalAmount: number;
    readonly installmentCount: number;
    readonly purchaseDate: number;
    readonly accountId: string;
    readonly paymentAccountId: string;
    readonly categoryId: string;
    readonly tagIds: string[];
    readonly comment: string;
    readonly status: PayLaterPlanStatusType;
    readonly installments?: PayLaterInstallmentResponse[];
}

export interface PayLaterPlanCreateRequest {
    name: string;
    totalAmount: number;
    installmentCount: number;
    purchaseDate: number;
    accountId: string;
    paymentAccountId: string;
    categoryId: string;
    tagIds: string[];
    comment: string;
    clientSessionId: string;
}

export interface PayLaterInstallmentPayRequest {
    id: string;
    paymentTime: number;
    utcOffset: number;
}

export interface PayLaterPlanDeleteRequest {
    id: string;
}

export class PayLaterPlan implements PayLaterInstallmentPlanInfoResponse {
    public id: string;
    public name: string;
    public totalAmount: number;
    public installmentCount: number;
    public purchaseDate: number;
    public accountId: string;
    public paymentAccountId: string;
    public categoryId: string;
    public tagIds: string[];
    public comment: string;
    public status: PayLaterPlanStatusType;
    public installments?: PayLaterInstallmentResponse[];

    private constructor(
        id: string,
        name: string,
        totalAmount: number,
        installmentCount: number,
        purchaseDate: number,
        accountId: string,
        paymentAccountId: string,
        categoryId: string,
        tagIds: string[],
        comment: string,
        status: PayLaterPlanStatusType,
        installments?: PayLaterInstallmentResponse[]
    ) {
        this.id = id;
        this.name = name;
        this.totalAmount = totalAmount;
        this.installmentCount = installmentCount;
        this.purchaseDate = purchaseDate;
        this.accountId = accountId;
        this.paymentAccountId = paymentAccountId;
        this.categoryId = categoryId;
        this.tagIds = tagIds;
        this.comment = comment;
        this.status = status;
        this.installments = installments;
    }

    public get paidInstallmentsCount(): number {
        if (!this.installments) return 0;
        return this.installments.filter(i => i.status === PayLaterInstallmentStatus.Paid).length;
    }

    public static of(response: PayLaterInstallmentPlanInfoResponse): PayLaterPlan {
        return new PayLaterPlan(
            response.id,
            response.name,
            response.totalAmount,
            response.installmentCount,
            response.purchaseDate,
            response.accountId,
            response.paymentAccountId,
            response.categoryId,
            response.tagIds ?? [],
            response.comment,
            response.status,
            response.installments
        );
    }
}
