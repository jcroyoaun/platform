import { useState } from 'react'
import { calculatorAPI, PackageInput, ComparisonResponse } from '../api/calculator'

export default function HomePage() {
  const [packages, setPackages] = useState<PackageInput[]>([
    {
      name: 'Paquete 1',
      regime: 'sueldos_salarios',
      currency: 'MXN',
      payment_frequency: 'monthly',
      gross_monthly_salary: 0,
    },
    {
      name: 'Paquete 2',
      regime: 'sueldos_salarios',
      currency: 'MXN',
      payment_frequency: 'monthly',
      gross_monthly_salary: 0,
    },
  ])
  const [results, setResults] = useState<ComparisonResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      const response = await calculatorAPI.comparePackages({ packages })
      setResults(response)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Error al calcular las compensaciones')
    } finally {
      setLoading(false)
    }
  }

  const handleExportPDF = async () => {
    try {
      const blob = await calculatorAPI.exportPDF()
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

  const handleClear = async () => {
    try {
      await calculatorAPI.clearSession()
      setResults(null)
      setPackages([
        {
          name: 'Paquete 1',
          regime: 'sueldos_salarios',
          currency: 'MXN',
          payment_frequency: 'monthly',
          gross_monthly_salary: 0,
        },
        {
          name: 'Paquete 2',
          regime: 'sueldos_salarios',
          currency: 'MXN',
          payment_frequency: 'monthly',
          gross_monthly_salary: 0,
        },
      ])
    } catch (err) {
      console.error('Error clearing session:', err)
    }
  }

  const updatePackage = (index: number, field: keyof PackageInput, value: any) => {
    const newPackages = [...packages]
    newPackages[index] = { ...newPackages[index], [field]: value }
    setPackages(newPackages)
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('es-MX', {
      style: 'currency',
      currency: 'MXN',
    }).format(amount)
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Compara Paquetes de Compensación
        </h1>
        <p className="text-xl text-gray-600">
          Calcula y compara compensaciones totales incluyendo salario, prestaciones e impuestos
        </p>
      </div>

      {!results ? (
        <form onSubmit={handleSubmit} className="space-y-8">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {packages.map((pkg, index) => (
              <div key={index} className="bg-white rounded-lg shadow-md p-6">
                <h3 className="text-lg font-semibold mb-4">
                  <input
                    type="text"
                    value={pkg.name}
                    onChange={(e) => updatePackage(index, 'name', e.target.value)}
                    className="border-b border-gray-300 focus:border-primary-500 outline-none w-full"
                  />
                </h3>

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Régimen
                    </label>
                    <select
                      value={pkg.regime}
                      onChange={(e) => updatePackage(index, 'regime', e.target.value)}
                      className="w-full border border-gray-300 rounded-md px-3 py-2"
                    >
                      <option value="sueldos_salarios">Sueldos y Salarios</option>
                      <option value="resico">RESICO</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Salario Mensual Bruto
                    </label>
                    <input
                      type="number"
                      value={pkg.gross_monthly_salary || ''}
                      onChange={(e) => updatePackage(index, 'gross_monthly_salary', parseFloat(e.target.value) || 0)}
                      className="w-full border border-gray-300 rounded-md px-3 py-2"
                      required
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Moneda
                    </label>
                    <select
                      value={pkg.currency}
                      onChange={(e) => updatePackage(index, 'currency', e.target.value)}
                      className="w-full border border-gray-300 rounded-md px-3 py-2"
                    >
                      <option value="MXN">MXN</option>
                      <option value="USD">USD</option>
                    </select>
                  </div>

                  {pkg.regime === 'sueldos_salarios' && (
                    <>
                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          checked={pkg.has_aguinaldo || false}
                          onChange={(e) => updatePackage(index, 'has_aguinaldo', e.target.checked)}
                          className="mr-2"
                        />
                        <label className="text-sm text-gray-700">Aguinaldo</label>
                      </div>

                      {pkg.has_aguinaldo && (
                        <input
                          type="number"
                          placeholder="Días de aguinaldo"
                          value={pkg.aguinaldo_days || 15}
                          onChange={(e) => updatePackage(index, 'aguinaldo_days', parseInt(e.target.value))}
                          className="w-full border border-gray-300 rounded-md px-3 py-2"
                        />
                      )}

                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          checked={pkg.has_prima_vacacional || false}
                          onChange={(e) => updatePackage(index, 'has_prima_vacacional', e.target.checked)}
                          className="mr-2"
                        />
                        <label className="text-sm text-gray-700">Prima Vacacional</label>
                      </div>

                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          checked={pkg.has_vales_despensa || false}
                          onChange={(e) => updatePackage(index, 'has_vales_despensa', e.target.checked)}
                          className="mr-2"
                        />
                        <label className="text-sm text-gray-700">Vales de Despensa</label>
                      </div>

                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          checked={pkg.has_fondo_ahorro || false}
                          onChange={(e) => updatePackage(index, 'has_fondo_ahorro', e.target.checked)}
                          className="mr-2"
                        />
                        <label className="text-sm text-gray-700">Fondo de Ahorro</label>
                      </div>
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
              {error}
            </div>
          )}

          <div className="flex justify-center">
            <button
              type="submit"
              disabled={loading}
              className="bg-primary-600 hover:bg-primary-700 text-white px-8 py-3 rounded-md font-medium disabled:opacity-50"
            >
              {loading ? 'Calculando...' : 'Comparar Paquetes'}
            </button>
          </div>
        </form>
      ) : (
        <div className="space-y-8">
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-2xl font-bold mb-6">Resultados de Comparación</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {results.results.map((result, index) => (
                <div
                  key={index}
                  className={`border-2 rounded-lg p-6 ${
                    result.package_name === results.best_package.package_name
                      ? 'border-green-500 bg-green-50'
                      : 'border-gray-200'
                  }`}
                >
                  <h3 className="text-xl font-semibold mb-4">
                    {result.package_name}
                    {result.package_name === results.best_package.package_name && (
                      <span className="ml-2 text-sm text-green-600 font-normal">
                        Mejor Opción
                      </span>
                    )}
                  </h3>

                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Salario Bruto Mensual:</span>
                      <span className="font-semibold">
                        {formatCurrency(result.calculation.gross_salary)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Salario Neto Mensual:</span>
                      <span className="font-semibold">
                        {formatCurrency(result.calculation.net_salary)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">ISR:</span>
                      <span className="text-red-600">
                        -{formatCurrency(result.calculation.isr_tax)}
                      </span>
                    </div>
                    <div className="flex justify-between border-t pt-3">
                      <span className="text-gray-900 font-medium">Compensación Anual Neta:</span>
                      <span className="font-bold text-lg text-primary-600">
                        {formatCurrency(result.calculation.yearly_net)}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-center space-x-4">
            <button
              onClick={handleExportPDF}
              className="bg-primary-600 hover:bg-primary-700 text-white px-6 py-2 rounded-md font-medium"
            >
              Exportar PDF
            </button>
            <button
              onClick={handleClear}
              className="bg-gray-200 hover:bg-gray-300 text-gray-700 px-6 py-2 rounded-md font-medium"
            >
              Nueva Comparación
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
