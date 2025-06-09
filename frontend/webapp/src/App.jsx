import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router'
import { HomePage } from './pages/HomePage'
import { TeamPage } from './pages/TeamPage'
import { RecordPage } from './pages/RecordPage'
import { VideosPage } from './pages/VideosPage'
import { SpacesPage } from './pages/SpacesPage'
import { SpaceDetailPage } from './pages/SpaceDetailPage'
import { UploadPage } from './pages/UploadPage'
import { SettingsPage } from './pages/SettingsPage'
import { LoginPage } from './auth/pages/LoginPage'
import { ProfilePage } from './pages/ProfilePage'
import ProtectedRoute from './auth/components/ProtectedRoute'
import { VideoPage } from './pages/VideoPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<ProtectedRoute><HomePage /></ProtectedRoute>} />
        <Route path="/team" element={<ProtectedRoute><TeamPage /></ProtectedRoute>} />
        <Route path="/record" element={<ProtectedRoute><RecordPage /></ProtectedRoute>} />
        <Route path="/upload" element={<ProtectedRoute><UploadPage /></ProtectedRoute>} />
        <Route path="/videos" element={<ProtectedRoute><VideosPage /></ProtectedRoute>} />
        <Route path="/spaces" element={<ProtectedRoute><SpacesPage /></ProtectedRoute>} />
        <Route path="/spaces/:spaceId" element={<ProtectedRoute><SpaceDetailPage /></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
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
  )
}

export default App
