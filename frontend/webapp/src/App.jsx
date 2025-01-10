import { BrowserRouter, Routes, Route } from 'react-router'
import Home from './pages/Home'
import LoginPage from './auth/pages/LoginPage'
import ProtectedRoute from './auth/components/ProtectedRoute'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={
          <ProtectedRoute>
            <Home />
          </ProtectedRoute>
        } />
        <Route path="/login" element={<LoginPage />} />
      </Routes>
    </BrowserRouter>
  )
}


export default App
