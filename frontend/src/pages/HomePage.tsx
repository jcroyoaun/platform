import { useState, useEffect } from 'react'
import PackageForm from '../components/PackageForm'
import ResultsSection from '../components/ResultsSection'
import { calculatorAPI, PackageInput, ComparisonResponse } from '../api/calculator'

export default function HomePage() {
  const [packages, setPackages] = useState<PackageInput[]>([
    {
      name: 'Paquete 1',
      regime: 'sueldos_salarios',
      currency: 'MXN',
      payment_frequency: 'monthly',
      gross_monthly_salary: 0,
      hours_per_week: 40,
      has_aguinaldo: true,
      aguinaldo_days: 15,
      has_vales_despensa: true,
      vales_despensa_amount: 3439,
      has_prima_vacacional: true,
      vacation_days: 12,
      prima_vacacional_percent: 25,
      has_fondo_ahorro: true,
      fondo_ahorro_percent: 13,
      unpaid_vacation_days: 0,
      has_equity: false,
      initial_equity_usd: 0,
      has_refreshers: false,
      refresher_min_usd: 0,
      refresher_max_usd: 0,
      other_benefits: [],
    },
  ])

  const [comparisonMode, setComparisonMode] = useState(false)
  const [results, setResults] = useState<ComparisonResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fiscalYear, setFiscalYear] = useState<any>(null)

  useEffect(() => {
    // Fetch fiscal year data on mount
    const fetchFiscalYear = async () => {
      try {
        // This would ideally come from an API, but for now we'll use a default
        setFiscalYear({
          year: 2025,
          usd_mxn_rate: 20.0,
          uma: 3439.46,
          smg: 278.80,
        })
      } catch (err) {
        console.error('Error fetching fiscal year:', err)
      }
    }
    fetchFiscalYear()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      const packagesToSubmit = comparisonMode ? packages : [packages[0]]
      const response = await calculatorAPI.comparePackages({ packages: packagesToSubmit })
      setResults(response)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Error al calcular las compensaciones')
    } finally {
      setLoading(false)
    }
  }

  const handleClear = () => {
    if (!window.confirm('¿Estás seguro de que quieres limpiar todos los campos?')) {
      return
    }
    setResults(null)
    setPackages([
      {
        name: 'Paquete 1',
        regime: 'sueldos_salarios',
        currency: 'MXN',
        payment_frequency: 'monthly',
        gross_monthly_salary: 0,
        hours_per_week: 40,
        has_aguinaldo: true,
        aguinaldo_days: 15,
        has_vales_despensa: true,
        vales_despensa_amount: 3439,
        has_prima_vacacional: true,
        vacation_days: 12,
        prima_vacacional_percent: 25,
        has_fondo_ahorro: true,
        fondo_ahorro_percent: 13,
        unpaid_vacation_days: 0,
        has_equity: false,
        initial_equity_usd: 0,
        has_refreshers: false,
        refresher_min_usd: 0,
        refresher_max_usd: 0,
        other_benefits: [],
      },
    ])
    setComparisonMode(false)
  }

  const addComparisonPackage = () => {
    setComparisonMode(true)
    if (packages.length === 1) {
      setPackages([
        ...packages,
        {
          name: 'Paquete 2',
          regime: 'sueldos_salarios',
          currency: 'MXN',
          payment_frequency: 'monthly',
          gross_monthly_salary: 0,
          hours_per_week: 40,
          has_aguinaldo: true,
          aguinaldo_days: 15,
          has_vales_despensa: true,
          vales_despensa_amount: 3439,
          has_prima_vacacional: true,
          vacation_days: 12,
          prima_vacacional_percent: 25,
          has_fondo_ahorro: true,
          fondo_ahorro_percent: 13,
          unpaid_vacation_days: 0,
          has_equity: false,
          initial_equity_usd: 0,
          has_refreshers: false,
          refresher_min_usd: 0,
          refresher_max_usd: 0,
          other_benefits: [],
        },
      ])
    }
  }

  const removeComparison = () => {
    setComparisonMode(false)
    setPackages([packages[0]])
  }

  const updatePackage = (index: number, updatedPackage: PackageInput) => {
    const newPackages = [...packages]
    newPackages[index] = updatedPackage
    setPackages(newPackages)
  }

  return (
    <div style={{ width: '100%', padding: '1.5rem', boxSizing: 'border-box' }}>
      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: '2rem', maxWidth: '780px', marginLeft: 'auto', marginRight: 'auto' }}>
        <h1 style={{ fontSize: '2.5rem', color: '#0f172a', marginBottom: '0.75rem', marginTop: 0, fontWeight: 700 }}>
          💰 Calculadora de Sueldo & Comparador
        </h1>
        <p style={{ fontSize: '1.25rem', color: '#10b981', margin: '0 0 0.5rem 0', fontWeight: 600 }}>
          Tu salario real, sin letras chiquitas
        </p>
        <p style={{ fontSize: '0.875rem', color: '#6366f1', margin: '0 0 0.5rem 0', fontWeight: 500, fontStyle: 'italic' }}>
          Ideal para Ingenieros, Freelancers y Contractors.
        </p>
        <p style={{ fontSize: '0.9rem', color: '#64748b', margin: 0 }}>
          Compara ofertas laborales lado a lado • Sueldos y Salarios vs RESICO • ISR 2025 • IMSS • UMA • Prestaciones
        </p>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} noValidate>
        {/* Packages Wrapper */}
        <div style={{ marginBottom: '2rem', position: 'relative' }}>
          <div
            style={{
              display: comparisonMode ? 'grid' : 'block',
              gridTemplateColumns: comparisonMode ? 'repeat(auto-fit, minmax(400px, 1fr))' : '1fr',
              gap: '1.5rem',
              width: comparisonMode ? 'auto' : '780px',
              maxWidth: comparisonMode ? '100%' : '780px',
              margin: '0 auto',
            }}
          >
            <PackageForm
              package={packages[0]}
              index={0}
              borderColor="#3b82f6"
              showRemoveButton={false}
              fiscalYear={fiscalYear}
              onChange={(updatedPackage) => updatePackage(0, updatedPackage)}
            />

            {comparisonMode && packages[1] && (
              <PackageForm
                package={packages[1]}
                index={1}
                borderColor="#8b5cf6"
                showRemoveButton={true}
                fiscalYear={fiscalYear}
                onChange={(updatedPackage) => updatePackage(1, updatedPackage)}
                onRemove={removeComparison}
              />
            )}
          </div>

          {/* Add Comparison Button */}
          {!comparisonMode && (
            <div
              style={{
                display: 'flex',
                position: 'absolute',
                left: 'calc(50% + 390px + 1rem)',
                top: 0,
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <button
                type="button"
                onClick={addComparisonPackage}
                style={{
                  writingMode: 'vertical-rl',
                  textOrientation: 'mixed',
                  background: 'linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%)',
                  color: '#3b82f6',
                  padding: '1.5rem 0.75rem',
                  border: '2px dashed #3b82f6',
                  borderRadius: '8px',
                  fontSize: '0.7rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  whiteSpace: 'nowrap',
                  maxHeight: '420px',
                }}
                onMouseOver={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%)'
                  e.currentTarget.style.borderColor = '#2563eb'
                  e.currentTarget.style.color = '#2563eb'
                  e.currentTarget.style.borderStyle = 'solid'
                }}
                onMouseOut={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%)'
                  e.currentTarget.style.borderColor = '#3b82f6'
                  e.currentTarget.style.color = '#3b82f6'
                  e.currentTarget.style.borderStyle = 'dashed'
                }}
              >
                <span style={{ fontSize: '1.1rem' }}>➕</span>
                <span>Comparar con otro paquete de compensación</span>
              </button>
            </div>
          )}
        </div>

        {/* Validation Errors */}
        {error && (
          <div style={{ background: '#fee2e2', borderLeft: '4px solid #ef4444', padding: '1rem', borderRadius: '8px', marginBottom: '1.5rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{ fontSize: '1.5rem' }}>⚠️</span>
              <div>
                <div style={{ fontWeight: 700, color: '#991b1b', marginBottom: '0.25rem' }}>Error de validación</div>
                <div style={{ color: '#991b1b' }}>{error}</div>
              </div>
            </div>
          </div>
        )}

        {/* Submit Button */}
        <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
          <button
            type="submit"
            disabled={loading}
            style={{
              background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
              color: 'white',
              padding: '1rem 3rem',
              border: 'none',
              borderRadius: '8px',
              fontSize: '1.25rem',
              fontWeight: 700,
              cursor: loading ? 'not-allowed' : 'pointer',
              boxShadow: '0 6px 12px rgba(16, 185, 129, 0.3)',
              transition: 'transform 0.2s, box-shadow 0.2s',
              marginRight: '1rem',
              opacity: loading ? 0.7 : 1,
            }}
            onMouseOver={(e) => {
              if (!loading) {
                e.currentTarget.style.transform = 'translateY(-2px)'
                e.currentTarget.style.boxShadow = '0 8px 16px rgba(16, 185, 129, 0.4)'
              }
            }}
            onMouseOut={(e) => {
              e.currentTarget.style.transform = 'translateY(0)'
              e.currentTarget.style.boxShadow = '0 6px 12px rgba(16, 185, 129, 0.3)'
            }}
          >
            {loading ? 'Calculando...' : '💰 Calcular Compensación'}
          </button>
          <button
            type="button"
            onClick={handleClear}
            style={{
              background: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
              color: 'white',
              padding: '1rem 2rem',
              border: 'none',
              borderRadius: '8px',
              fontSize: '1rem',
              fontWeight: 600,
              cursor: 'pointer',
              boxShadow: '0 4px 8px rgba(239, 68, 68, 0.3)',
              transition: 'transform 0.2s, box-shadow 0.2s',
            }}
            onMouseOver={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)'
              e.currentTarget.style.boxShadow = '0 6px 12px rgba(239, 68, 68, 0.4)'
            }}
            onMouseOut={(e) => {
              e.currentTarget.style.transform = 'translateY(0)'
              e.currentTarget.style.boxShadow = '0 4px 8px rgba(239, 68, 68, 0.3)'
            }}
          >
            🗑️ Limpiar Todo
          </button>
        </div>
      </form>

      {/* Results Section */}
      {results && <ResultsSection results={results} fiscalYear={fiscalYear} />}

      {/* Info Box */}
      {results && (
        <div style={{ background: '#eff6ff', borderLeft: '4px solid #2563eb', padding: '1.25rem', borderRadius: '8px', fontSize: '0.875rem', color: '#1e40af' }}>
          <strong>💡 Tip:</strong> El "Neto Mensual Ajustado" incluye las prestaciones anuales divididas entre 12 meses. Es una mejor métrica para comparar tu poder adquisitivo real.
        </div>
      )}
    </div>
  )
}
