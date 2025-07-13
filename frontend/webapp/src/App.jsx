import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router'
import { HomePage } from './pages/HomePage'
import { TeamPage } from './pages/TeamPage'
import { RecordPage } from './pages/RecordPage'
import { SettingsPage } from './pages/SettingsPage'
import { LoginPage } from './auth/pages/LoginPage'
import { ProfilePage } from './pages/ProfilePage'
import ProtectedRoute from './auth/components/ProtectedRoute'
import { VideoPage } from './pages/VideoPage'
import { UploadPage } from './pages/UploadPage';
import { ChannelDashboardPage } from './pages/ChannelDashboard';
import ChannelPage from './pages/ChannelPage';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={
          <ProtectedRoute>
            <HomePage />
          </ProtectedRoute>
        } />
        <Route path="/team" element={
          <ProtectedRoute>
            <TeamPage />
          </ProtectedRoute>
        } />
        <Route path="/record" element={
          <ProtectedRoute>
            <RecordPage />
          </ProtectedRoute>
        } />
        <Route path="/upload" element={
          <ProtectedRoute>
            <UploadPage />
          </ProtectedRoute>
        } />
        <Route path="/channels" element={
          <ProtectedRoute>
            <ChannelDashboardPage />
          </ProtectedRoute>
        } />
        <Route path="/channel/:id" element={
          <ProtectedRoute>
            <ChannelPage />
          </ProtectedRoute>
        } />
        <Route path="/video/:id" element={
          <ProtectedRoute>
            <VideoPage />
          </ProtectedRoute>
        } />
        <Route path="/settings" element={
          <ProtectedRoute>
            <SettingsPage />
          </ProtectedRoute>
        } />
        <Route path="/profile" element={
          <ProtectedRoute>
            <ProfilePage />
          </ProtectedRoute>
        } />
      </Routes>
    </BrowserRouter>
  )
}

export default App
