import { useState } from 'react'
import { PackageInput, OtherBenefit } from '../api/calculator'

interface PackageFormProps {
  package: PackageInput
  index: number
  borderColor: string
  showRemoveButton: boolean
  fiscalYear: any
  onChange: (pkg: PackageInput) => void
  onRemove?: () => void
}

export default function PackageForm({
  package: pkg,
  borderColor,
  showRemoveButton,
  onChange,
  onRemove,
}: PackageFormProps) {
  const [showEquity, setShowEquity] = useState(pkg.has_equity || false)
  const [showRefreshers, setShowRefreshers] = useState(pkg.has_refreshers || false)

  const updateField = (field: keyof PackageInput, value: any) => {
    onChange({ ...pkg, [field]: value })
  }

  const formatNumber = (value: string) => {
    const cleanValue = value.replace(/[^\d.]/g, '')
    const parts = cleanValue.split('.')
    parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    return parts.length > 1 ? parts[0] + '.' + parts[1].substring(0, 2) : parts[0]
  }

  const addBenefit = () => {
    const newBenefit: OtherBenefit = {
      name: '',
      amount: 0,
      tax_free: false,
      currency: 'MXN',
      cadence: 'monthly',
      is_percentage: false,
    }
    updateField('other_benefits', [...(pkg.other_benefits || []), newBenefit])
  }

  const removeBenefit = (benefitIndex: number) => {
    const newBenefits = [...(pkg.other_benefits || [])]
    newBenefits.splice(benefitIndex, 1)
    updateField('other_benefits', newBenefits)
  }

  const updateBenefit = (benefitIndex: number, field: keyof OtherBenefit, value: any) => {
    const newBenefits = [...(pkg.other_benefits || [])]
    newBenefits[benefitIndex] = { ...newBenefits[benefitIndex], [field]: value }
    updateField('other_benefits', newBenefits)
  }

  const toggleRegime = (regime: string) => {
    if (regime === 'resico') {
      updateField('regime', 'resico')
      updateField('has_aguinaldo', false)
      updateField('has_vales_despensa', false)
      updateField('has_prima_vacacional', false)
      updateField('has_fondo_ahorro', false)
    } else {
      updateField('regime', 'sueldos_salarios')
      updateField('currency', 'MXN')
      updateField('has_aguinaldo', true)
      updateField('has_vales_despensa', true)
      updateField('has_prima_vacacional', true)
      updateField('has_fondo_ahorro', true)
    }
  }

  const getSalaryLabel = () => {
    switch (pkg.payment_frequency) {
      case 'hourly': return '💰 Tarifa Por Hora'
      case 'daily': return '💰 Salario Diario'
      case 'weekly': return '💰 Salario Semanal'
      case 'biweekly': return '💰 Salario Quincenal'
      default: return '💰 Salario Bruto'
    }
  }

  return (
    <div
      style={{
        background: 'white',
        padding: '1.5rem',
        borderRadius: '12px',
        boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
        border: `3px solid ${borderColor}`,
        position: 'relative',
      }}
    >
      {showRemoveButton && onRemove && (
        <button
          type="button"
          onClick={onRemove}
          style={{
            position: 'absolute',
            top: '1rem',
            right: '1rem',
            background: '#ef4444',
            color: 'white',
            border: 'none',
            borderRadius: '50%',
            width: '32px',
            height: '32px',
            fontSize: '1.25rem',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
            transition: 'all 0.2s ease',
            zIndex: 10,
          }}
          title="Quitar comparación"
          onMouseOver={(e) => {
            e.currentTarget.style.background = '#dc2626'
            e.currentTarget.style.transform = 'scale(1.1)'
          }}
          onMouseOut={(e) => {
            e.currentTarget.style.background = '#ef4444'
            e.currentTarget.style.transform = 'scale(1)'
          }}
        >
          🗑️
        </button>
      )}

      <div style={{ marginBottom: '1.5rem', paddingBottom: '1rem', borderBottom: '2px solid #e2e8f0' }}>
        <input
          type="text"
          placeholder="Nombre del paquete"
          value={pkg.name}
          onChange={(e) => updateField('name', e.target.value)}
          style={{
            width: '100%',
            padding: '0.75rem',
            border: '2px solid #e2e8f0',
            borderRadius: '8px',
            fontSize: '1rem',
            fontWeight: 600,
            color: '#1e293b',
          }}
        />
      </div>

      {/* Regime */}
      <div style={{ marginBottom: '1rem' }}>
        <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
          📋 Régimen Fiscal
        </label>
        <select
          value={pkg.regime}
          onChange={(e) => toggleRegime(e.target.value)}
          style={{
            width: '100%',
            padding: '0.75rem',
            border: '2px solid #e2e8f0',
            borderRadius: '8px',
            fontSize: '0.875rem',
            background: 'white',
            cursor: 'pointer',
          }}
        >
          <option value="sueldos_salarios">Sueldos y Salarios</option>
          <option value="resico">RESICO</option>
        </select>
      </div>

      {/* Currency (RESICO only) */}
      {pkg.regime === 'resico' && (
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
            💵 Moneda
          </label>
          <select
            value={pkg.currency}
            onChange={(e) => updateField('currency', e.target.value)}
            style={{
              width: '100%',
              padding: '0.75rem',
              border: '2px solid #e2e8f0',
              borderRadius: '8px',
              fontSize: '0.875rem',
              background: 'white',
              cursor: 'pointer',
            }}
          >
            <option value="MXN">MXN (Pesos)</option>
            <option value="USD">USD (Dólares)</option>
          </select>
        </div>
      )}

      {/* Salary */}
      <div style={{ marginBottom: '1rem' }}>
        <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
          {getSalaryLabel()}
        </label>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr auto', gap: '0.5rem' }}>
          <input
            type="text"
            placeholder="Ej: 12,000"
            value={pkg.gross_monthly_salary ? formatNumber(pkg.gross_monthly_salary.toString()) : ''}
            onChange={(e) => {
              const value = e.target.value.replace(/,/g, '')
              updateField('gross_monthly_salary', parseFloat(value) || 0)
            }}
            style={{
              width: '100%',
              padding: '0.75rem',
              border: '2px solid #e2e8f0',
              borderRadius: '8px',
              fontSize: '1rem',
            }}
          />
          <select
            value={pkg.payment_frequency}
            onChange={(e) => updateField('payment_frequency', e.target.value)}
            style={{
              width: '120px',
              padding: '0.75rem',
              border: '2px solid #e2e8f0',
              borderRadius: '8px',
              fontSize: '0.75rem',
              background: 'white',
              cursor: 'pointer',
            }}
          >
            <option value="monthly">Mensual</option>
            <option value="biweekly">Quincenal</option>
            <option value="weekly">Semanal</option>
            {pkg.regime === 'resico' && <option value="daily">Diario</option>}
            {pkg.regime === 'resico' && <option value="hourly">Por Hora</option>}
          </select>
        </div>
      </div>

      {/* Hours per week (hourly only) */}
      {pkg.payment_frequency === 'hourly' && (
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
            ⏰ Horas por Semana
          </label>
          <input
            type="number"
            value={pkg.hours_per_week || 40}
            onChange={(e) => updateField('hours_per_week', parseInt(e.target.value) || 40)}
            min="1"
            max="168"
            style={{
              width: '100%',
              padding: '0.75rem',
              border: '2px solid #e2e8f0',
              borderRadius: '8px',
              fontSize: '0.875rem',
            }}
          />
        </div>
      )}

      {/* Unpaid Vacation (RESICO only) */}
      {pkg.regime === 'resico' && (
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
            📅 Días de Descanso (No pagados)
          </label>
          <input
            type="number"
            value={pkg.unpaid_vacation_days || 0}
            onChange={(e) => updateField('unpaid_vacation_days', parseInt(e.target.value) || 0)}
            min="0"
            max="365"
            style={{
              width: '100%',
              padding: '0.75rem',
              border: '2px solid #e2e8f0',
              borderRadius: '8px',
              fontSize: '0.875rem',
            }}
          />
          <p style={{ margin: '0.5rem 0 0 0', fontSize: '0.75rem', color: '#64748b', lineHeight: 1.4 }}>
            <em>Como RESICO, si no trabajas, no cobras. Esto ajustará tu ingreso anual real.</em>
          </p>
        </div>
      )}

      {/* Benefits (Sueldos y Salarios only) */}
      {pkg.regime === 'sueldos_salarios' && (
        <div style={{ background: '#f8fafc', padding: '1rem', borderRadius: '8px', marginBottom: '1rem' }}>
          <div style={{ fontSize: '0.875rem', fontWeight: 600, color: '#1e293b', marginBottom: '0.75rem' }}>
            🎁 Prestaciones
          </div>

          <label style={{ display: 'flex', alignItems: 'center', marginBottom: '0.5rem', cursor: 'pointer', fontSize: '0.875rem' }}>
            <input
              type="checkbox"
              checked={pkg.has_aguinaldo || false}
              onChange={(e) => updateField('has_aguinaldo', e.target.checked)}
              style={{ marginRight: '0.5rem' }}
            />
            Aguinaldo
            <input
              type="number"
              value={pkg.aguinaldo_days || 15}
              onChange={(e) => updateField('aguinaldo_days', parseInt(e.target.value) || 15)}
              min="15"
              max="30"
              style={{
                width: '60px',
                padding: '0.25rem',
                marginLeft: '0.5rem',
                border: '1px solid #e2e8f0',
                borderRadius: '4px',
                fontSize: '0.75rem',
              }}
            />
            días
          </label>

          <label style={{ display: 'flex', alignItems: 'center', marginBottom: '0.5rem', cursor: 'pointer', fontSize: '0.875rem' }}>
            <input
              type="checkbox"
              checked={pkg.has_vales_despensa || false}
              onChange={(e) => updateField('has_vales_despensa', e.target.checked)}
              style={{ marginRight: '0.5rem' }}
            />
            Vales de despensa $
            <input
              type="text"
              value={pkg.vales_despensa_amount ? formatNumber(pkg.vales_despensa_amount.toString()) : '3439'}
              onChange={(e) => {
                const value = e.target.value.replace(/,/g, '')
                updateField('vales_despensa_amount', parseFloat(value) || 3439)
              }}
              style={{
                width: '80px',
                padding: '0.25rem',
                marginLeft: '0.5rem',
                border: '1px solid #e2e8f0',
                borderRadius: '4px',
                fontSize: '0.75rem',
              }}
            />
            /mes
          </label>

          <label style={{ display: 'flex', alignItems: 'center', marginBottom: '0.5rem', cursor: 'pointer', fontSize: '0.875rem' }}>
            <input
              type="checkbox"
              checked={pkg.has_prima_vacacional || false}
              onChange={(e) => updateField('has_prima_vacacional', e.target.checked)}
              style={{ marginRight: '0.5rem' }}
            />
            Prima Vacacional:
            <input
              type="number"
              value={pkg.vacation_days || 12}
              onChange={(e) => updateField('vacation_days', parseInt(e.target.value) || 12)}
              min="12"
              style={{
                width: '50px',
                padding: '0.25rem',
                marginLeft: '0.5rem',
                border: '1px solid #e2e8f0',
                borderRadius: '4px',
                fontSize: '0.75rem',
              }}
            />
            días @
            <input
              type="number"
              value={pkg.prima_vacacional_percent || 25}
              onChange={(e) => updateField('prima_vacacional_percent', parseInt(e.target.value) || 25)}
              min="25"
              style={{
                width: '50px',
                padding: '0.25rem',
                marginLeft: '0.25rem',
                border: '1px solid #e2e8f0',
                borderRadius: '4px',
                fontSize: '0.75rem',
              }}
            />
            %
          </label>

          <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer', fontSize: '0.875rem' }}>
            <input
              type="checkbox"
              checked={pkg.has_fondo_ahorro || false}
              onChange={(e) => updateField('has_fondo_ahorro', e.target.checked)}
              style={{ marginRight: '0.5rem' }}
            />
            Fondo de Ahorro
            <input
              type="number"
              value={pkg.fondo_ahorro_percent || 13}
              onChange={(e) => updateField('fondo_ahorro_percent', parseInt(e.target.value) || 13)}
              min="1"
              max="13"
              style={{
                width: '50px',
                padding: '0.25rem',
                marginLeft: '0.5rem',
                border: '1px solid #e2e8f0',
                borderRadius: '4px',
                fontSize: '0.75rem',
              }}
            />
            %
          </label>
        </div>
      )}

      {/* Other Benefits */}
      <div style={{ background: '#fef3c7', padding: '1rem', borderRadius: '8px', marginBottom: '1rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
          <span style={{ fontSize: '0.875rem', fontWeight: 600, color: '#92400e' }}>✨ Otras Prestaciones</span>
          <button
            type="button"
            onClick={addBenefit}
            style={{
              background: '#059669',
              color: 'white',
              padding: '0.25rem 0.75rem',
              border: 'none',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.75rem',
            }}
          >
            + Agregar
          </button>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {(pkg.other_benefits || []).map((benefit, benefitIndex) => (
            <div
              key={benefitIndex}
              style={{
                display: 'flex',
                gap: '0.5rem',
                background: 'white',
                padding: '0.5rem',
                borderRadius: '6px',
                border: '1px solid #e2e8f0',
                flexWrap: 'wrap',
              }}
            >
              <input
                type="text"
                placeholder="Ej: Bono anual"
                value={benefit.name}
                onChange={(e) => updateBenefit(benefitIndex, 'name', e.target.value)}
                style={{
                  flex: 1,
                  minWidth: '120px',
                  padding: '0.5rem',
                  border: '1px solid #e2e8f0',
                  borderRadius: '4px',
                  fontSize: '0.75rem',
                }}
              />
              <input
                type="text"
                placeholder="$1,500"
                value={benefit.amount ? formatNumber(benefit.amount.toString()) : ''}
                onChange={(e) => {
                  const value = e.target.value.replace(/,/g, '')
                  updateBenefit(benefitIndex, 'amount', parseFloat(value) || 0)
                }}
                style={{
                  width: '90px',
                  padding: '0.5rem',
                  border: '1px solid #e2e8f0',
                  borderRadius: '4px',
                  fontSize: '0.75rem',
                }}
              />
              <select
                value={benefit.cadence}
                onChange={(e) => updateBenefit(benefitIndex, 'cadence', e.target.value)}
                style={{
                  width: '90px',
                  padding: '0.5rem',
                  border: '1px solid #e2e8f0',
                  borderRadius: '4px',
                  fontSize: '0.7rem',
                }}
              >
                <option value="monthly">Mensual</option>
                <option value="annual">Anual</option>
              </select>
              <label style={{ display: 'flex', alignItems: 'center', whiteSpace: 'nowrap', fontSize: '0.7rem', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={benefit.tax_free || false}
                  onChange={(e) => updateBenefit(benefitIndex, 'tax_free', e.target.checked)}
                  style={{ marginRight: '0.25rem' }}
                />
                Libre ISR
              </label>
              <button
                type="button"
                onClick={() => removeBenefit(benefitIndex)}
                style={{
                  background: '#ef4444',
                  color: 'white',
                  padding: '0.35rem 0.5rem',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                  fontSize: '0.7rem',
                }}
              >
                🗑️
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Equity Section */}
      <div style={{ marginTop: '1.5rem' }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer', marginBottom: '0.75rem' }}>
          <input
            type="checkbox"
            checked={showEquity}
            onChange={(e) => {
              setShowEquity(e.target.checked)
              updateField('has_equity', e.target.checked)
            }}
            style={{ width: '18px', height: '18px', cursor: 'pointer' }}
          />
          <span style={{ fontWeight: 600, color: '#1e293b', fontSize: '0.875rem' }}>📊 ¿Incluir Equity / RSUs?</span>
        </label>

        {showEquity && (
          <div
            style={{
              padding: '1.5rem',
              background: '#f8fafc',
              borderRadius: '8px',
              border: '2px solid #e2e8f0',
            }}
          >
            <h3 style={{ margin: '0 0 1rem 0', fontSize: '1rem', color: '#0f172a' }}>📊 Equity / RSUs</h3>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
                💰 Initial Equity Grant (USD)
              </label>
              <input
                type="text"
                placeholder="Ej: 30000"
                value={pkg.initial_equity_usd ? formatNumber(pkg.initial_equity_usd.toString()) : ''}
                onChange={(e) => {
                  const value = e.target.value.replace(/,/g, '')
                  updateField('initial_equity_usd', parseFloat(value) || 0)
                }}
                style={{
                  width: '100%',
                  padding: '0.75rem',
                  border: '2px solid #e2e8f0',
                  borderRadius: '8px',
                  fontSize: '0.875rem',
                }}
              />
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={showRefreshers}
                  onChange={(e) => {
                    setShowRefreshers(e.target.checked)
                    updateField('has_refreshers', e.target.checked)
                  }}
                  style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                />
                <span style={{ fontWeight: 600, color: '#1e293b', fontSize: '0.875rem' }}>✨ Annual Refreshers?</span>
              </label>
            </div>

            {showRefreshers && (
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
                    Min Refresher (USD/año)
                  </label>
                  <input
                    type="text"
                    placeholder="8000"
                    value={pkg.refresher_min_usd ? formatNumber(pkg.refresher_min_usd.toString()) : ''}
                    onChange={(e) => {
                      const value = e.target.value.replace(/,/g, '')
                      updateField('refresher_min_usd', parseFloat(value) || 0)
                    }}
                    style={{
                      width: '100%',
                      padding: '0.75rem',
                      border: '2px solid #e2e8f0',
                      borderRadius: '8px',
                      fontSize: '0.875rem',
                    }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontWeight: 600, marginBottom: '0.5rem', color: '#1e293b', fontSize: '0.875rem' }}>
                    Max Refresher (USD/año)
                  </label>
                  <input
                    type="text"
                    placeholder="12000"
                    value={pkg.refresher_max_usd ? formatNumber(pkg.refresher_max_usd.toString()) : ''}
                    onChange={(e) => {
                      const value = e.target.value.replace(/,/g, '')
                      updateField('refresher_max_usd', parseFloat(value) || 0)
                    }}
                    style={{
                      width: '100%',
                      padding: '0.75rem',
                      border: '2px solid #e2e8f0',
                      borderRadius: '8px',
                      fontSize: '0.875rem',
                    }}
                  />
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
