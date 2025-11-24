export default function DevelopersPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h1 className="text-4xl font-bold text-gray-900 mb-4">
        API para Desarrolladores
      </h1>
      <p className="text-xl text-gray-600 mb-12">
        Integra cálculos de compensación total en tus aplicaciones
      </p>

      <div className="bg-white rounded-lg shadow-md p-8 mb-8">
        <h2 className="text-2xl font-semibold mb-4">Comenzar</h2>
        <ol className="list-decimal list-inside space-y-3 text-gray-700">
          <li>Crea una cuenta y verifica tu email</li>
          <li>Genera tu API key desde tu panel de desarrollador</li>
          <li>Realiza solicitudes a la API usando tu key</li>
        </ol>
      </div>

      <div className="bg-white rounded-lg shadow-md p-8 mb-8">
        <h2 className="text-2xl font-semibold mb-4">Autenticación</h2>
        <p className="text-gray-700 mb-4">
          Todas las solicitudes requieren una API key en el header:
        </p>
        <div className="bg-gray-50 p-4 rounded-md font-mono text-sm">
          X-API-Key: tu_api_key_aqui
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-md p-8">
        <h2 className="text-2xl font-semibold mb-4">Endpoint: Calcular Compensación</h2>
        <p className="text-gray-700 mb-4">
          <code className="bg-gray-100 px-2 py-1 rounded">POST /api/v1/calculate</code>
        </p>

        <h3 className="font-semibold mt-6 mb-2">Request Body:</h3>
        <pre className="bg-gray-50 p-4 rounded-md overflow-x-auto text-sm">
{`{
  "salary": 50000,
  "regime": "sueldos",
  "has_aguinaldo": true,
  "aguinaldo_days": 15,
  "has_prima_vacacional": true,
  "vacation_days": 12,
  "prima_vacacional_percent": 25
}`}
        </pre>

        <h3 className="font-semibold mt-6 mb-2">Response:</h3>
        <pre className="bg-gray-50 p-4 rounded-md overflow-x-auto text-sm">
{`{
  "success": true,
  "data": {
    "gross_salary": 50000,
    "net_salary": 42350,
    "isr_tax": 6500,
    "yearly_net": 520000,
    "breakdown": {
      "aguinaldo_gross": 24630,
      "aguinaldo_net": 22450,
      ...
    }
  }
}`}
        </pre>
      </div>
    </div>
  )
}
