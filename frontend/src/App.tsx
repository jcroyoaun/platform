import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import HomePage from './pages/HomePage'
import LoginPage from './pages/LoginPage'
import SignupPage from './pages/SignupPage'
import DeveloperPage from './pages/DeveloperPage'
import DevelopersPage from './pages/DevelopersPage'
import PrivacyPage from './pages/PrivacyPage'
import TermsPage from './pages/TermsPage'
import ForgottenPasswordPage from './pages/ForgottenPasswordPage'
import PasswordResetPage from './pages/PasswordResetPage'
import EmailVerificationPage from './pages/EmailVerificationPage'
import NotFoundPage from './pages/NotFoundPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<HomePage />} />
          <Route path="login" element={<LoginPage />} />
          <Route path="signup" element={<SignupPage />} />
          <Route path="account/developer" element={<DeveloperPage />} />
          <Route path="developers" element={<DevelopersPage />} />
          <Route path="privacy" element={<PrivacyPage />} />
          <Route path="terms" element={<TermsPage />} />
          <Route path="forgotten-password" element={<ForgottenPasswordPage />} />
          <Route path="password-reset/:token" element={<PasswordResetPage />} />
          <Route path="verify-email/:token" element={<EmailVerificationPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
