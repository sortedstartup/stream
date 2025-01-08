import { useState } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router'
import Home from './pages/Home'
import LoginPage from './auth/pages/LoginPage'


function App() {
  const [count, setCount] = useState(0)

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        {/* login page */ }
        <Route path="/login" element={<LoginPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
