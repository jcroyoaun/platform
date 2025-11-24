import { useState, useEffect } from 'react'
import { authAPI, User } from '../api/auth'
import { useNavigate } from 'react-router-dom'

export default function DeveloperPage() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    loadUser()
  }, [])

  const loadUser = async () => {
    try {
      const currentUser = await authAPI.getCurrentUser()
      if (!currentUser) {
        navigate('/login')
        return
      }
      setUser(currentUser)
    } catch (error) {
      navigate('/login')
    } finally {
      setLoading(false)
    }
  }

  const handleGenerateAPIKey = async () => {
    try {
      await authAPI.generateAPIKey()
      await loadUser()
    } catch (error) {
      console.error('Error generating API key:', error)
    }
  }

  const handleResendVerification = async () => {
    try {
      await authAPI.resendVerificationEmail()
      alert('Email de verificación enviado')
    } catch (error) {
      console.error('Error resending verification:', error)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[calc(100vh-16rem)]">
        <div className="text-gray-600">Cargando...</div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Panel de Desarrollador</h1>

      {user && !user.email_verified && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4 mb-6">
          <p className="text-yellow-800 mb-2">
            Tu email no está verificado. Por favor revisa tu bandeja de entrada.
          </p>
          <button
            onClick={handleResendVerification}
            className="text-yellow-900 underline hover:no-underline"
          >
            Reenviar email de verificación
          </button>
        </div>
      )}

      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Información de Cuenta</h2>
        <div className="space-y-3">
          <div>
            <span className="text-gray-600">Email:</span>
            <span className="ml-2 font-medium">{user?.email}</span>
          </div>
          <div>
            <span className="text-gray-600">Estado:</span>
            <span className={`ml-2 ${user?.email_verified ? 'text-green-600' : 'text-yellow-600'}`}>
              {user?.email_verified ? 'Verificado' : 'No verificado'}
            </span>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-4">API Key</h2>

        {user?.api_key ? (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-md font-mono text-sm break-all">
              {user.api_key}
            </div>
            <p className="text-sm text-gray-600">
              Usa esta clave en el header <code className="bg-gray-100 px-1 py-0.5 rounded">X-API-Key</code> para autenticar tus solicitudes.
            </p>
            <button
              onClick={handleGenerateAPIKey}
              className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm"
            >
              Regenerar API Key
            </button>
          </div>
        ) : (
          <div>
            <p className="text-gray-600 mb-4">
              Aún no tienes una API key. Genera una para comenzar a usar la API.
            </p>
            <button
              onClick={handleGenerateAPIKey}
              className="bg-primary-600 hover:bg-primary-700 text-white px-4 py-2 rounded-md"
            >
              Generar API Key
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
