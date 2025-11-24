import apiClient from './client'

export interface PackageInput {
  name: string
  regime: 'sueldos_salarios' | 'resico'
  currency: 'MXN' | 'USD'
  exchange_rate?: number
  payment_frequency: 'hourly' | 'daily' | 'weekly' | 'biweekly' | 'monthly'
  hours_per_week?: number
  gross_monthly_salary: number
  has_aguinaldo?: boolean
  aguinaldo_days?: number
  has_vales_despensa?: boolean
  vales_despensa_amount?: number
  has_prima_vacacional?: boolean
  vacation_days?: number
  prima_vacacional_percent?: number
  has_fondo_ahorro?: boolean
  fondo_ahorro_percent?: number
  unpaid_vacation_days?: number
  other_benefits?: OtherBenefit[]
  has_equity?: boolean
  initial_equity_usd?: number
  has_refreshers?: boolean
  refresher_min_usd?: number
  refresher_max_usd?: number
}

export interface OtherBenefit {
  name: string
  amount: number
  tax_free: boolean
  currency: 'MXN' | 'USD'
  cadence: 'monthly' | 'annual'
  is_percentage: boolean
}

export interface SalaryCalculation {
  gross_salary: number
  net_salary: number
  isr_tax: number
  subsidio_empleo: number
  imss_worker: number
  sbc: number
  yearly_gross_base: number
  yearly_gross: number
  yearly_net: number
  monthly_adjusted: number
  aguinaldo_gross?: number
  aguinaldo_isr?: number
  aguinaldo_net?: number
  prima_vacacional_gross?: number
  prima_vacacional_isr?: number
  prima_vacacional_net?: number
  fondo_ahorro_yearly?: number
  infonavit_employer_annual?: number
  imss_employer_annual?: number
  other_benefits?: Array<{
    name: string
    amount: number
    isr: number
    net: number
    tax_free: boolean
    cadence: string
  }>
}

export interface ComparisonRequest {
  packages: PackageInput[]
}

export interface ComparisonResponse {
  results: Array<{
    package_name: string
    calculation: SalaryCalculation
  }>
  best_package: {
    package_name: string
    calculation: SalaryCalculation
  }
  fiscal_year: {
    year: number
    uma_monthly: number
    usd_mxn_rate: number
  }
}

export const calculatorAPI = {
  async comparePackages(data: ComparisonRequest): Promise<ComparisonResponse> {
    const response = await apiClient.post('/api/v1/compare', data)
    return response.data
  },

  async exportPDF(): Promise<Blob> {
    const response = await apiClient.get('/api/v1/export-pdf', {
      responseType: 'blob',
    })
    return response.data
  },

  async clearSession() {
    return apiClient.post('/api/v1/clear-session')
  },
}
