import { ComparisonResponse } from '../api/calculator'

interface ResultsSectionProps {
  results: ComparisonResponse
  fiscalYear: any
}

export default function ResultsSection({ results }: ResultsSectionProps) {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('es-MX', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount)
  }

  const handleExportPDF = async () => {
    try {
      const response = await fetch('/api/export-pdf', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/pdf',
        },
      })
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'TotalComp_Comparacion_2025.pdf'
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (err) {
      console.error('Error exporting PDF:', err)
    }
  }

  return (
    <div style={{ background: 'linear-gradient(135deg, #f0fdf4 0%, #dbeafe 100%)', padding: '2rem', borderRadius: '12px', boxShadow: '0 4px 6px rgba(0,0,0,0.1)', marginBottom: '2rem' }}>
      <h2 style={{ textAlign: 'center', color: '#059669', marginBottom: '2rem', fontSize: '2rem' }}>
        📊 Comparación de Resultados
      </h2>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))', gap: '1.5rem' }}>
        {results.results.map((result, idx) => (
          <div key={idx} style={{ background: 'white', padding: '1.5rem', borderRadius: '10px', boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}>
            <h3 style={{ color: '#2563eb', marginBottom: '1rem', paddingBottom: '0.75rem', borderBottom: '2px solid #e2e8f0' }}>
              {result.package_name || `Paquete ${idx + 1}`}
            </h3>

            {/* Key Metrics */}
            <div style={{ background: '#eff6ff', padding: '1rem', borderRadius: '8px', marginBottom: '1rem', textAlign: 'center' }}>
              <div style={{ fontSize: '0.875rem', color: '#64748b', marginBottom: '0.25rem' }}>💰 Neto Mensual</div>
              <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#2563eb' }}>
                ${formatCurrency(result.net_salary)}
              </div>
            </div>

            <div style={{ background: '#f0fdf4', padding: '1rem', borderRadius: '8px', marginBottom: '1rem', textAlign: 'center' }}>
              <div style={{ fontSize: '0.875rem', color: '#64748b', marginBottom: '0.25rem' }}>📅 Neto Anual</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#059669' }}>
                ${formatCurrency(result.yearly_net)}
              </div>
            </div>

            <div style={{ background: '#fef3c7', padding: '1rem', borderRadius: '8px', textAlign: 'center' }}>
              <div style={{ fontSize: '0.875rem', color: '#64748b', marginBottom: '0.25rem' }}>✨ Neto Mensual Ajustado</div>
              <div style={{ fontSize: '1.25rem', fontWeight: 700, color: '#92400e' }}>
                ${formatCurrency(result.monthly_adjusted)}
              </div>
              <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.25rem' }}>(Incluye prestaciones anuales)</div>
            </div>

            {/* Detailed Breakdown */}
            <div style={{ marginTop: '1.5rem', paddingTop: '1rem', borderTop: '2px solid #e2e8f0' }}>
              <h4 style={{ fontSize: '0.9rem', fontWeight: 600, color: '#1e293b', marginBottom: '0.75rem' }}>💰 Desglose Mensual:</h4>

              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem' }}>
                <tbody>
                  <tr style={{ borderBottom: '2px solid #e2e8f0' }}>
                    <td style={{ padding: '0.5rem 0', color: '#64748b' }}>Salario Bruto</td>
                    <td style={{ padding: '0.5rem 0', textAlign: 'right', fontWeight: 600 }}>
                      ${formatCurrency(result.gross_salary)}
                    </td>
                  </tr>
                  <tr style={{ borderBottom: '1px solid #e2e8f0' }}>
                    <td style={{ padding: '0.5rem 0', color: '#64748b' }}>(-) ISR</td>
                    <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#ef4444', fontWeight: 600 }}>
                      -${formatCurrency(result.isr_tax)}
                    </td>
                  </tr>
                  {result.subsidio_empleo > 0 && (
                    <tr style={{ borderBottom: '1px solid #e2e8f0', background: '#f0fdf4' }}>
                      <td style={{ padding: '0.5rem 0', color: '#059669', fontWeight: 500 }}>(+) Subsidio al Empleo</td>
                      <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#059669', fontWeight: 600 }}>
                        +${formatCurrency(result.subsidio_empleo)}
                      </td>
                    </tr>
                  )}
                  {result.imss_worker > 0 && (
                    <tr style={{ borderBottom: '1px solid #e2e8f0' }}>
                      <td style={{ padding: '0.5rem 0', color: '#64748b' }}>(-) IMSS Trabajador</td>
                      <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#ef4444', fontWeight: 600 }}>
                        -${formatCurrency(result.imss_worker)}
                      </td>
                    </tr>
                  )}
                  {result.fondo_ahorro_employee > 0 && (
                    <tr style={{ borderBottom: '1px solid #e2e8f0' }}>
                      <td style={{ padding: '0.5rem 0', color: '#64748b' }}>(-) Fondo de Ahorro (empleado)</td>
                      <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#ef4444', fontWeight: 600 }}>
                        -${formatCurrency(result.fondo_ahorro_employee)}
                      </td>
                    </tr>
                  )}
                  {result.vales_despensa_monthly > 0 && (
                    <tr style={{ borderBottom: '1px solid #e2e8f0', background: '#f0fdf4' }}>
                      <td style={{ padding: '0.5rem 0', color: '#059669', fontWeight: 500 }}>(+) Vales de Despensa</td>
                      <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#059669', fontWeight: 600 }}>
                        +${formatCurrency(result.vales_despensa_monthly)}
                      </td>
                    </tr>
                  )}
                  {result.other_benefits && result.other_benefits.map((benefit, bidx) =>
                    benefit.cadence === 'monthly' ? (
                      <tr key={bidx} style={{ borderBottom: '1px solid #e2e8f0', background: '#f0fdf4' }}>
                        <td style={{ padding: '0.5rem 0', color: '#059669', fontWeight: 500 }}>
                          (+) {benefit.name} {benefit.tax_free && <span style={{ fontSize: '0.65rem', background: '#dcfce7', color: '#166534', padding: '0.125rem 0.25rem', borderRadius: '3px', marginLeft: '0.25rem' }}>Libre ISR</span>}
                        </td>
                        <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#059669', fontWeight: 600 }}>
                          +${formatCurrency(benefit.net)}
                        </td>
                      </tr>
                    ) : null
                  )}
                  <tr style={{ borderTop: '2px solid #2563eb', background: '#eff6ff' }}>
                    <td style={{ padding: '0.5rem 0', fontWeight: 700, color: '#2563eb' }}>Neto Mensual</td>
                    <td style={{ padding: '0.5rem 0', textAlign: 'right', fontWeight: 700, color: '#2563eb' }}>
                      ${formatCurrency(result.net_salary)}
                    </td>
                  </tr>
                </tbody>
              </table>

              {/* Annual Benefits */}
              {((result.aguinaldo_net || 0) > 0 || (result.prima_vacacional_net || 0) > 0 || (result.fondo_ahorro_yearly || 0) > 0) && (
                <>
                  <h4 style={{ fontSize: '0.9rem', fontWeight: 600, color: '#1e293b', marginTop: '1.5rem', marginBottom: '0.75rem' }}>
                    🎁 Prestaciones Anuales:
                  </h4>
                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem' }}>
                    <tbody>
                      {(result.aguinaldo_net || 0) > 0 && (
                        <tr style={{ borderBottom: '1px solid #e2e8f0' }}>
                          <td style={{ padding: '0.5rem 0', color: '#64748b' }}>🎄 Aguinaldo (neto)</td>
                          <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#059669', fontWeight: 600 }}>
                            ${formatCurrency(result.aguinaldo_net || 0)}
                          </td>
                        </tr>
                      )}
                      {(result.prima_vacacional_net || 0) > 0 && (
                        <tr style={{ borderBottom: '1px solid #e2e8f0' }}>
                          <td style={{ padding: '0.5rem 0', color: '#64748b' }}>🏖️ Prima Vacacional (neto)</td>
                          <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#059669', fontWeight: 600 }}>
                            ${formatCurrency(result.prima_vacacional_net || 0)}
                          </td>
                        </tr>
                      )}
                      {(result.fondo_ahorro_yearly || 0) > 0 && (
                        <tr style={{ borderBottom: '1px solid #e2e8f0' }}>
                          <td style={{ padding: '0.5rem 0', color: '#64748b' }}>💰 Fondo de Ahorro (retorno 2x)</td>
                          <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#059669', fontWeight: 600 }}>
                            ${formatCurrency(result.fondo_ahorro_yearly || 0)}
                          </td>
                        </tr>
                      )}
                      {result.other_benefits && result.other_benefits.map((benefit, bidx) =>
                        benefit.cadence === 'annual' ? (
                          <tr key={bidx} style={{ borderBottom: '1px solid #e2e8f0' }}>
                            <td style={{ padding: '0.5rem 0', color: '#64748b' }}>
                              ✨ {benefit.name} {benefit.tax_free && <span style={{ fontSize: '0.65rem', background: '#dcfce7', color: '#166534', padding: '0.125rem 0.25rem', borderRadius: '3px', marginLeft: '0.25rem' }}>Libre ISR</span>}
                            </td>
                            <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#059669', fontWeight: 600 }}>
                              ${formatCurrency(benefit.net)}
                            </td>
                          </tr>
                        ) : null
                      )}
                    </tbody>
                  </table>
                </>
              )}

              {/* Annual Totals */}
              <h4 style={{ fontSize: '0.9rem', fontWeight: 600, color: '#1e293b', marginTop: '1.5rem', marginBottom: '0.75rem' }}>
                📊 Totales Anuales:
              </h4>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem' }}>
                <tbody>
                  {result.yearly_gross_base > 0 && (
                    <tr style={{ borderBottom: '1px solid #e2e8f0', background: '#fef3c7' }}>
                      <td style={{ padding: '0.5rem 0', color: '#92400e', fontWeight: 600 }}>Bruto Anual (solo salario)</td>
                      <td style={{ padding: '0.5rem 0', textAlign: 'right', fontWeight: 700, color: '#92400e' }}>
                        ${formatCurrency(result.yearly_gross_base)}
                      </td>
                    </tr>
                  )}
                  <tr style={{ borderBottom: '2px solid #6366f1', background: '#eef2ff' }}>
                    <td style={{ padding: '0.5rem 0', color: '#4338ca', fontWeight: 700 }}>💰 Comp Total Anual</td>
                    <td style={{ padding: '0.5rem 0', textAlign: 'right', fontWeight: 700, color: '#4338ca' }}>
                      ${formatCurrency(result.yearly_gross)}
                    </td>
                  </tr>
                  {(result.unpaid_vacation_loss || 0) > 0 && (
                    <tr style={{ borderBottom: '1px solid #e2e8f0', background: '#fee2e2' }}>
                      <td style={{ padding: '0.5rem 0', color: '#991b1b', fontWeight: 500 }}>
                        (-) Ingreso no percibido por vacaciones
                        <span style={{ fontSize: '0.7rem', color: '#64748b' }}> ({result.unpaid_vacation_days} días)</span>
                      </td>
                      <td style={{ padding: '0.5rem 0', textAlign: 'right', color: '#ef4444', fontWeight: 600 }}>
                        -${formatCurrency(result.unpaid_vacation_loss || 0)}
                      </td>
                    </tr>
                  )}
                  <tr style={{ borderTop: '2px solid #7c3aed', background: '#faf5ff' }}>
                    <td style={{ padding: '0.5rem 0', fontWeight: 700, color: '#7c3aed' }}>Neto Anual Total</td>
                    <td style={{ padding: '0.5rem 0', textAlign: 'right', fontWeight: 700, color: '#7c3aed' }}>
                      ${formatCurrency(result.yearly_net)}
                    </td>
                  </tr>
                  <tr style={{ borderTop: '2px solid #059669', background: '#f0fdf4' }}>
                    <td style={{ padding: '0.5rem 0', fontWeight: 700, color: '#059669' }}>Neto Mensual Ajustado (÷ 12)</td>
                    <td style={{ padding: '0.5rem 0', textAlign: 'right', fontWeight: 700, color: '#059669', fontSize: '0.9rem' }}>
                      ${formatCurrency(result.monthly_adjusted)}
                    </td>
                  </tr>
                </tbody>
              </table>

              <div style={{ marginTop: '0.5rem', padding: '0.5rem', background: '#f0fdf4', borderRadius: '4px', fontSize: '0.7rem', color: '#059669' }}>
                💡 <strong>Neto Mensual Ajustado:</strong> Incluye prestaciones anuales prorrateadas mensualmente. Mejor métrica para comparar poder adquisitivo real.
              </div>

              {result.sbc > 0 && (
                <div style={{ marginTop: '1rem', padding: '0.75rem', background: '#f1f5f9', borderRadius: '6px', fontSize: '0.75rem', color: '#475569' }}>
                  <strong>📌 Info:</strong> SBC Diario: ${formatCurrency(result.sbc)}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Download PDF Button */}
      <div style={{ textAlign: 'center', marginTop: '2rem', paddingTop: '2rem', borderTop: '2px solid #e2e8f0' }}>
        <button
          onClick={handleExportPDF}
          style={{
            display: 'inline-block',
            background: 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)',
            color: 'white',
            padding: '0.875rem 2rem',
            border: 'none',
            borderRadius: '8px',
            fontWeight: 700,
            fontSize: '1rem',
            cursor: 'pointer',
            boxShadow: '0 6px 12px rgba(99, 102, 241, 0.3)',
            transition: 'transform 0.2s, box-shadow 0.2s',
          }}
          onMouseOver={(e) => {
            e.currentTarget.style.transform = 'translateY(-2px)'
            e.currentTarget.style.boxShadow = '0 8px 16px rgba(99, 102, 241, 0.4)'
          }}
          onMouseOut={(e) => {
            e.currentTarget.style.transform = 'translateY(0)'
            e.currentTarget.style.boxShadow = '0 6px 12px rgba(99, 102, 241, 0.3)'
          }}
        >
          📄 Descargar Reporte PDF
        </button>
        <div style={{ marginTop: '0.75rem', fontSize: '0.875rem', color: '#64748b' }}>
          {results.results.length > 1
            ? '💡 Descarga un reporte profesional lado a lado de ambos paquetes'
            : '💡 Descarga un reporte profesional de tu compensación'}
        </div>
      </div>
    </div>
  )
}
