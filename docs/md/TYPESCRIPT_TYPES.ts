// TYPESCRIPT TYPES FOR FRONTEND
// Copy this file to your frontend project: src/types/api.ts

// ============================================
// ENUMS
// ============================================

export type BudgetPeriod = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly' | 'custom';
export type BudgetStatus = 'active' | 'warning' | 'critical' | 'exceeded' | 'expired';
export type AlertThreshold = '50' | '75' | '90' | '100';

export type GoalType = 'savings' | 'debt' | 'investment' | 'purchase' | 'emergency' | 'retirement' | 'education' | 'other';
export type GoalStatus = 'active' | 'completed' | 'paused' | 'cancelled' | 'overdue';
export type GoalPriority = 'low' | 'medium' | 'high' | 'critical';
export type ContributionFrequency = 'one_time' | 'daily' | 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly';

export type DebtType = 'credit_card' | 'loan' | 'mortgage' | 'student_loan' | 'personal_loan' | 'car_loan' | 'other';
export type DebtStatus = 'active' | 'paid_off' | 'defaulted' | 'refinanced';
export type PaymentFrequency = 'weekly' | 'biweekly' | 'monthly';

export type TransactionType = 'income' | 'expense' | 'transfer';
export type TransactionStatus = 'completed' | 'pending' | 'cancelled' | 'failed';

export type AccountType = 'bank' | 'cash' | 'credit_card' | 'investment' | 'crypto' | 'other';
export type AccountStatus = 'active' | 'inactive' | 'closed';

export type CategoryType = 'income' | 'expense';

export type IncomeType = 'salary' | 'freelance' | 'business' | 'investment' | 'rental' | 'pension' | 'other';
export type IncomeFrequency = 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly';

// ============================================
// MAIN ENTITIES
// ============================================

export interface Budget {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  amount: number;
  currency: string;
  period: BudgetPeriod;
  start_date: string;
  end_date?: string;
  category_id?: string;
  account_id?: string;
  spent_amount: number;
  remaining_amount: number;
  percentage_spent: number;
  status: BudgetStatus;
  last_calculated_at?: string;
  enable_alerts: boolean;
  alert_thresholds: AlertThreshold[];
  alerted_at?: string;
  notification_sent: boolean;
  allow_rollover: boolean;
  rollover_amount: number;
  carry_over_percent?: number;
  auto_adjust: boolean;
  auto_adjust_percentage?: number;
  auto_adjust_based_on?: string;
  created_at: string;
  updated_at: string;
}

export interface Goal {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  type: GoalType;
  priority: GoalPriority;
  target_amount: number;
  current_amount: number;
  currency: string;
  start_date: string;
  target_date?: string;
  completed_at?: string;
  percentage_complete: number;
  remaining_amount: number;
  status: GoalStatus;
  suggested_contribution?: number;
  contribution_frequency?: ContributionFrequency;
  auto_contribute: boolean;
  auto_contribute_amount?: number;
  auto_contribute_account_id?: string;
  linked_account_id?: string;
  enable_reminders: boolean;
  reminder_frequency?: string;
  last_reminder_sent_at?: string;
  milestones?: string;
  notes?: string;
  tags?: string;
  created_at: string;
  updated_at: string;
}

export interface Debt {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  type: DebtType;
  principal_amount: number;
  current_balance: number;
  currency: string;
  interest_rate: number;
  minimum_payment: number;
  payment_frequency?: PaymentFrequency;
  next_payment_date?: string;
  start_date: string;
  paid_off_date?: string;
  total_paid: number;
  total_interest_paid: number;
  remaining_amount: number;
  percentage_paid: number;
  status: DebtStatus;
  creditor_name?: string;
  account_number?: string;
  linked_account_id?: string;
  auto_pay: boolean;
  auto_pay_amount?: number;
  enable_reminders: boolean;
  reminder_days_before?: number;
  last_reminder_sent_at?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  account_id: string;
  category_id?: string;
  type: TransactionType;
  amount: number;
  currency: string;
  description?: string;
  transaction_date: string;
  status: TransactionStatus;
  reference_number?: string;
  merchant_name?: string;
  location?: string;
  notes?: string;
  tags?: string;
  is_recurring: boolean;
  recurring_frequency?: string;
  parent_transaction_id?: string;
  linked_transaction_id?: string;
  linked_budget_id?: string;
  linked_goal_id?: string;
  linked_debt_id?: string;
  attachments?: string;
  created_at: string;
  updated_at: string;
}

export interface Account {
  id: string;
  user_id: string;
  name: string;
  type: AccountType;
  balance: number;
  currency: string;
  status: AccountStatus;
  bank_name?: string;
  account_number?: string;
  routing_number?: string;
  iban?: string;
  swift_code?: string;
  branch?: string;
  credit_limit?: number;
  interest_rate?: number;
  opening_date?: string;
  closing_date?: string;
  is_primary: boolean;
  include_in_net_worth: boolean;
  notes?: string;
  icon?: string;
  color?: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  user_id?: string;
  name: string;
  type: CategoryType;
  parent_id?: string;
  icon?: string;
  color?: string;
  is_active: boolean;
  is_system: boolean;
  display_order?: number;
  budget_amount?: number;
  description?: string;
  children?: Category[];
  created_at: string;
  updated_at: string;
}

export interface BudgetProfile {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  category_id?: string;
  min_amount: number;
  max_amount: number;
  target_amount?: number;
  constraint_type: 'hard' | 'soft' | 'aspirational';
  priority: number;
  is_flexible: boolean;
  is_active: boolean;
  version: number;
  effective_from: string;
  effective_to?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface IncomeProfile {
  id: string;
  user_id: string;
  name: string;
  type: IncomeType;
  amount: number;
  currency: string;
  frequency: IncomeFrequency;
  start_date: string;
  end_date?: string;
  is_active: boolean;
  is_verified: boolean;
  verified_at?: string;
  source_name?: string;
  account_id?: string;
  tax_rate?: number;
  is_taxable: boolean;
  is_guaranteed: boolean;
  confidence_level?: number;
  version: number;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// ============================================
// RESPONSE TYPES
// ============================================

export interface BudgetSummary {
  total_budgets: number;
  active_budgets: number;
  exceeded_budgets: number;
  warning_budgets: number;
  total_amount: number;
  total_spent: number;
  total_remaining: number;
  average_percentage: number;
  budgets_by_category?: Record<string, CategoryBudgetSum>;
}

export interface CategoryBudgetSum {
  category_id: string;
  category_name: string;
  amount: number;
  spent: number;
  remaining: number;
  percentage: number;
}

export interface GoalSummary {
  total_goals: number;
  active_goals: number;
  completed_goals: number;
  overdue_goals: number;
  total_target_amount: number;
  total_current_amount: number;
  total_remaining: number;
  average_progress: number;
  goals_by_type?: Record<string, GoalTypeSum>;
  goals_by_priority?: Record<string, number>;
}

export interface GoalTypeSum {
  count: number;
  target_amount: number;
  current_amount: number;
  progress: number;
}

export interface GoalProgress {
  goal_id: string;
  name: string;
  type: GoalType;
  priority: GoalPriority;
  target_amount: number;
  current_amount: number;
  remaining_amount: number;
  percentage_complete: number;
  status: GoalStatus;
  start_date: string;
  target_date?: string;
  days_elapsed: number;
  days_remaining?: number;
  time_progress?: number;
  on_track?: boolean;
  suggested_contribution?: number;
  projected_completion_date?: string;
}

export interface DebtSummary {
  total_debts: number;
  active_debts: number;
  paid_off_debts: number;
  overdue_debts: number;
  total_principal_amount: number;
  total_current_balance: number;
  total_paid: number;
  total_remaining: number;
  total_interest_paid: number;
  average_progress: number;
  debts_by_type?: Record<string, DebtTypeSum>;
  debts_by_status?: Record<string, number>;
}

export interface DebtTypeSum {
  count: number;
  principal_amount: number;
  current_balance: number;
  total_paid: number;
  progress: number;
}

// ============================================
// REQUEST TYPES
// ============================================

export interface CreateBudgetRequest {
  name: string;
  description?: string;
  amount: number;
  period: BudgetPeriod;
  start_date: string;
  end_date?: string;
  category_id?: string;
  account_id?: string;
  enable_alerts?: boolean;
  alert_thresholds?: AlertThreshold[];
  allow_rollover?: boolean;
  carry_over_percent?: number;
}

export interface CreateGoalRequest {
  name: string;
  description?: string;
  type: GoalType;
  priority: GoalPriority;
  target_amount: number;
  start_date: string;
  target_date?: string;
  contribution_frequency?: ContributionFrequency;
  linked_account_id?: string;
}

export interface CreateDebtRequest {
  name: string;
  description?: string;
  type: DebtType;
  principal_amount: number;
  interest_rate: number;
  minimum_payment: number;
  payment_frequency?: PaymentFrequency;
  start_date: string;
  creditor_name?: string;
  linked_account_id?: string;
}

export interface CreateTransactionRequest {
  account_id: string;
  category_id?: string;
  type: TransactionType;
  amount: number;
  description?: string;
  transaction_date: string;
  merchant_name?: string;
  notes?: string;
}

export interface ContributeToGoalRequest {
  amount: number;
}

export interface AddDebtPaymentRequest {
  amount: number;
}

// ============================================
// API RESPONSE WRAPPER
// ============================================

export interface ApiResponse<T> {
  data: T;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: any;
  };
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

// ============================================
// QUERY PARAMETERS
// ============================================

export interface PaginationParams {
  limit?: number;
  offset?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

export interface TransactionFilters extends PaginationParams {
  type?: TransactionType;
  category_id?: string;
  account_id?: string;
  start_date?: string;
  end_date?: string;
  min_amount?: number;
  max_amount?: number;
}

export interface BudgetFilters extends PaginationParams {
  status?: BudgetStatus;
  period?: BudgetPeriod;
  category_id?: string;
  account_id?: string;
}

export interface GoalFilters extends PaginationParams {
  status?: GoalStatus;
  type?: GoalType;
  priority?: GoalPriority;
}
