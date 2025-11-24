import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { authAPI } from '../api/auth'

export default function PasswordResetPage() {
  const { token } = useParams()
  const [newPassword, setNewPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token) return

    setLoading(true)
    setError('')

    try {
      await authAPI.resetPassword(token, newPassword)
      alert('Contraseña restablecida exitosamente')
      navigate('/login')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Error al restablecer la contraseña')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-[calc(100vh-16rem)] flex items-center justify-center px-4">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold text-gray-900">
            Nueva Contraseña
          </h2>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
              {error}
            </div>
          )}

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700">
              Nueva Contraseña (mínimo 8 caracteres)
            </label>
            <input
              id="password"
              type="password"
              required
              minLength={8}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary-600 hover:bg-primary-700 text-white py-2 px-4 rounded-md font-medium disabled:opacity-50"
          >
            {loading ? 'Guardando...' : 'Restablecer Contraseña'}
          </button>
        </form>
      </div>
    </div>
  )
}
