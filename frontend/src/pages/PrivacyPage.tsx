export default function PrivacyPage() {
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h1 className="text-4xl font-bold text-gray-900 mb-8">Aviso de Privacidad</h1>

      <div className="prose prose-lg max-w-none">
        <p className="text-gray-600 mb-6">Última actualización: Noviembre 2025</p>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">1. Responsable del Tratamiento</h2>
          <p className="text-gray-700">
            TotalCompMX es responsable del tratamiento de sus datos personales.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">2. Datos Personales Recabados</h2>
          <p className="text-gray-700 mb-2">Recabamos los siguientes datos personales:</p>
          <ul className="list-disc list-inside text-gray-700 space-y-1">
            <li>Correo electrónico</li>
            <li>Contraseña (encriptada)</li>
            <li>Datos de compensación ingresados voluntariamente</li>
          </ul>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">3. Finalidades del Tratamiento</h2>
          <p className="text-gray-700 mb-2">Sus datos serán utilizados para:</p>
          <ul className="list-disc list-inside text-gray-700 space-y-1">
            <li>Proveer acceso a la plataforma</li>
            <li>Realizar cálculos de compensación</li>
            <li>Generar reportes y análisis</li>
            <li>Mejorar nuestros servicios</li>
          </ul>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">4. Derechos ARCO</h2>
          <p className="text-gray-700">
            Usted tiene derecho a Acceder, Rectificar, Cancelar u Oponerse al tratamiento de sus datos
            personales, así como a revocar su consentimiento para el tratamiento de los mismos.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">5. Seguridad de los Datos</h2>
          <p className="text-gray-700">
            Implementamos medidas de seguridad técnicas, administrativas y físicas para proteger
            sus datos personales contra daño, pérdida, alteración, destrucción o uso no autorizado.
          </p>
        </section>
      </div>
    </div>
  )
}
