import { useState } from 'react'
import { Link } from 'react-router-dom'
import { authAPI } from '../api/auth'

export default function ForgottenPasswordPage() {
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      await authAPI.forgottenPassword(email)
      setSubmitted(true)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Error al enviar el email')
    } finally {
      setLoading(false)
    }
  }

  if (submitted) {
    return (
      <div className="min-h-[calc(100vh-16rem)] flex items-center justify-center px-4">
        <div className="max-w-md w-full text-center">
          <h2 className="text-3xl font-bold text-gray-900 mb-4">
            Email Enviado
          </h2>
          <p className="text-gray-600 mb-6">
            Hemos enviado un enlace para restablecer tu contraseña a {email}.
            Por favor revisa tu bandeja de entrada.
          </p>
          <Link
            to="/login"
            className="text-primary-600 hover:text-primary-500 font-medium"
          >
            Volver a iniciar sesión
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-[calc(100vh-16rem)] flex items-center justify-center px-4">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold text-gray-900">
            Recuperar Contraseña
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Ingresa tu email y te enviaremos un enlace para restablecer tu contraseña
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
              {error}
            </div>
          )}

          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700">
              Email
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary-600 hover:bg-primary-700 text-white py-2 px-4 rounded-md font-medium disabled:opacity-50"
          >
            {loading ? 'Enviando...' : 'Enviar Enlace'}
          </button>

          <div className="text-center">
            <Link to="/login" className="text-sm text-primary-600 hover:text-primary-500">
              Volver a iniciar sesión
            </Link>
          </div>
        </form>
      </div>
    </div>
  )
}
