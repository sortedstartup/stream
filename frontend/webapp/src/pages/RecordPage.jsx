import React from 'react'
import { Layout } from "../components/layout/Layout";
import ScreenRecorder from "../components/ScreenRecorder";
import { useState } from "react";

export const RecordPage = () => {
  const [uploadStatus, setUploadStatus] = useState({ type: null, message: null });

  const handleUploadSuccess = () => {
    setUploadStatus({
      type: 'success',
      message: 'Video uploaded successfully!'
    });
  };

  const handleUploadError = (error) => {
    let errorMessage = 'Failed to upload video.';
    
    // Provide more specific error messages based on the error
    if (error.message.includes('status: 413')) {
      errorMessage = 'Video file is too large. Please record a shorter video.';
    } else if (error.message.includes('status: 401')) {
      errorMessage = 'Session expired. Please login again.';
    } else if (!navigator.onLine) {
      errorMessage = 'No internet connection. Please check your network.';
    }

    setUploadStatus({
      type: 'error',
      message: errorMessage
    });
  };

  const clearStatus = () => {
    setUploadStatus({ type: null, message: null });
  };

  return (
    <Layout>
      <div className="space-y-8">
        <div className="text-center">
          <h1 className="text-3xl font-bold mb-4">Record</h1>
          <p className="mb-4">Start recording your screen or camera</p>
          
          {/* Status Messages */}
          {uploadStatus.message && (
            <div className={`alert ${uploadStatus.type === 'success' ? 'alert-success' : 'alert-error'} mb-4`}>
              <div className="flex justify-between items-center w-full">
                <div className="flex items-center">
                  {uploadStatus.type === 'error' ? (
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  ) : (
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  )}
                  <span>{uploadStatus.message}</span>
                </div>
                {uploadStatus.type === 'error' && (
                  <button 
                    className="btn btn-sm btn-outline"
                    onClick={clearStatus}
                  >
                    Dismiss
                  </button>
                )}
              </div>
            </div>
          )}

          <div className="flex justify-center">   
            <ScreenRecorder 
              onUploadSuccess={handleUploadSuccess}
              onUploadError={handleUploadError}
            />
          </div>
        </div>
      </div>
    </Layout>
  );
} 