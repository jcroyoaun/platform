export default function TermsPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h1 className="text-4xl font-bold text-gray-900 mb-8">Términos y Condiciones</h1>

      <div className="prose prose-lg max-w-none">
        <p className="text-gray-600 mb-6">Última actualización: Noviembre 2025</p>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">1. Aceptación de los Términos</h2>
          <p className="text-gray-700">
            Al acceder y usar TotalCompMX, usted acepta estar sujeto a estos términos y condiciones.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">2. Uso del Servicio</h2>
          <p className="text-gray-700">
            TotalCompMX proporciona herramientas de cálculo de compensación total. Los resultados
            son estimaciones y no constituyen asesoría legal, fiscal o financiera.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">3. Responsabilidades del Usuario</h2>
          <ul className="list-disc list-inside text-gray-700 space-y-1">
            <li>Mantener la confidencialidad de su cuenta</li>
            <li>Proporcionar información precisa</li>
            <li>No usar el servicio para fines ilegales</li>
            <li>No intentar acceder a datos de otros usuarios</li>
          </ul>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">4. Limitación de Responsabilidad</h2>
          <p className="text-gray-700">
            TotalCompMX no se hace responsable de decisiones tomadas basándose únicamente en
            los cálculos proporcionados. Recomendamos consultar con profesionales calificados.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">5. Modificaciones</h2>
          <p className="text-gray-700">
            Nos reservamos el derecho de modificar estos términos en cualquier momento. Las
            modificaciones entrarán en vigor inmediatamente después de su publicación.
          </p>
        </section>
      </div>
    </div>
  )
}
