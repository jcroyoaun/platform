import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { authAPI } from '../api/auth'

export default function EmailVerificationPage() {
  const { token } = useParams()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [message, setMessage] = useState('')

  useEffect(() => {
    if (token) {
      verifyEmail(token)
    }
  }, [token])

  const verifyEmail = async (token: string) => {
    try {
      await authAPI.verifyEmail(token)
      setStatus('success')
      setMessage('Tu email ha sido verificado exitosamente')
    } catch (err: any) {
      setStatus('error')
      setMessage(err.response?.data?.error || 'El enlace de verificación es inválido o ha expirado')
    }
  }

  return (
    <div className="min-h-[calc(100vh-16rem)] flex items-center justify-center px-4">
      <div className="max-w-md w-full text-center">
        {status === 'loading' && (
          <div>
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Verificando...
            </h2>
          </div>
        )}

        {status === 'success' && (
          <div>
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
              <svg className="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7"></path>
              </svg>
            </div>
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Email Verificado
            </h2>
            <p className="text-gray-600 mb-6">{message}</p>
            <Link
              to="/account/developer"
              className="inline-block bg-primary-600 hover:bg-primary-700 text-white px-6 py-2 rounded-md font-medium"
            >
              Ir a mi cuenta
            </Link>
          </div>
        )}

        {status === 'error' && (
          <div>
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
              <svg className="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </div>
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Error de Verificación
            </h2>
            <p className="text-gray-600 mb-6">{message}</p>
            <Link
              to="/"
              className="text-primary-600 hover:text-primary-500 font-medium"
            >
              Volver al inicio
            </Link>
          </div>
        )}
      </div>
    </div>
  )
}
