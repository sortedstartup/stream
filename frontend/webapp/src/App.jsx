import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router'
import { HomePage } from './pages/HomePage'
import { TeamPage } from './pages/TeamPage'
import { RecordPage } from './pages/RecordPage'
import { VideosPage } from './pages/VideosPage'
import { SettingsPage } from './pages/SettingsPage'
import { LoginPage } from './auth/pages/LoginPage'
import { ProfilePage } from './pages/ProfilePage'
import ProtectedRoute from './auth/components/ProtectedRoute'
import { VideoPage } from './pages/VideoPage'
import { UploadPage } from './pages/UploadPage';
import { RecordingProvider } from "./context/RecordingContext";

function App() {
  return (
    <RecordingProvider>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<ProtectedRoute><HomePage /></ProtectedRoute>} />
        <Route path="/team" element={<ProtectedRoute><TeamPage /></ProtectedRoute>} />
        <Route path="/record" element={<ProtectedRoute><RecordPage /></ProtectedRoute>} />
        <Route path="/videos" element={<ProtectedRoute><VideosPage /></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
        <Route path="/upload" element={<ProtectedRoute><UploadPage /></ProtectedRoute>} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/profile" element={
          <ProtectedRoute>
            <ProfilePage />
          </ProtectedRoute>
        } />
        <Route path="/video/:id" element={
          <ProtectedRoute>
            <VideoPage />
          </ProtectedRoute>
        } />
      </Routes>
    </BrowserRouter>
    </RecordingProvider>
  )
}

export default App
