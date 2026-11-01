import { Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { RedirigirInicio } from './components/RedirigirInicio'
import { LoginPage } from './pages/LoginPage'
import { RegistroPage } from './pages/RegistroPage'
import { RecuperarPage } from './pages/RecuperarPage'
import { RestablecerPage } from './pages/RestablecerPage'
import { FormularioPage } from './pages/FormularioPage'
import { NoEncontradaPage } from './pages/NoEncontradaPage'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<RedirigirInicio />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/registro" element={<RegistroPage />} />
        <Route path="/recuperar" element={<RecuperarPage />} />
        <Route path="/restablecer" element={<RestablecerPage />} />
        <Route element={<ProtectedRoute />}>
          <Route path="/formulario" element={<FormularioPage />} />
        </Route>
        <Route path="*" element={<NoEncontradaPage />} />
      </Route>
    </Routes>
  )
}