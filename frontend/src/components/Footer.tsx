import { Link } from 'react-router-dom'

export default function Footer() {
  return (
    <footer className="bg-gray-50 border-t mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              TotalComp<span className="text-primary-600">MX</span>
            </h3>
            <p className="text-gray-600 text-sm">
              Compara compensaciones totales en México de forma transparente y precisa.
            </p>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-gray-900 mb-4">Legal</h4>
            <ul className="space-y-2">
              <li>
                <Link to="/privacy" className="text-sm text-gray-600 hover:text-gray-900">
                  Aviso de Privacidad
                </Link>
              </li>
              <li>
                <Link to="/terms" className="text-sm text-gray-600 hover:text-gray-900">
                  Términos y Condiciones
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-gray-900 mb-4">Recursos</h4>
            <ul className="space-y-2">
              <li>
                <Link to="/developers" className="text-sm text-gray-600 hover:text-gray-900">
                  API para Desarrolladores
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 pt-8 border-t border-gray-200">
          <p className="text-center text-sm text-gray-500">
            © {new Date().getFullYear()} TotalCompMX. Todos los derechos reservados.
          </p>
        </div>
      </div>
    </footer>
  )
}
