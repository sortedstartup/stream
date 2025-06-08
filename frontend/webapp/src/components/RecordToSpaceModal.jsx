import React, { useState } from 'react'
import ScreenRecorder from './ScreenRecorder'

export const RecordToSpaceModal = ({ isOpen, onClose, spaceId, spaceName, onVideoRecorded }) => {
    const [uploadStatus, setUploadStatus] = useState({ type: null, message: null })

    const handleUploadSuccess = (result) => {
        setUploadStatus({
            type: 'success',
            message: `Video recorded and added to "${spaceName}" successfully!`
        })
        
        onVideoRecorded && onVideoRecorded(result)
        
        // Auto-close after success
        setTimeout(() => {
            onClose()
            setUploadStatus({ type: null, message: null })
        }, 2000)
    }

    const handleUploadError = (error) => {
        let errorMessage = 'Failed to record video.'
        
        if (error.message.includes('status: 413')) {
            errorMessage = 'Video file is too large. Please record a shorter video.'
        } else if (error.message.includes('status: 401')) {
            errorMessage = 'Session expired. Please login again.'
        } else if (!navigator.onLine) {
            errorMessage = 'No internet connection. Please check your network.'
        }

        setUploadStatus({
            type: 'error',
            message: errorMessage
        })
    }

    const handleClose = () => {
        setUploadStatus({ type: null, message: null })
        onClose()
    }

    if (!isOpen) return null

    return (
        <div className="modal modal-open">
            <div className="modal-box max-w-4xl w-full">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="font-bold text-lg">
                        Record Video to "{spaceName}"
                    </h3>
                    <button 
                        className="btn btn-sm btn-circle btn-ghost"
                        onClick={handleClose}
                    >
                        ✕
                    </button>
                </div>

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
                                    onClick={() => setUploadStatus({ type: null, message: null })}
                                >
                                    Dismiss
                                </button>
                            )}
                        </div>
                    </div>
                )}

                <div className="mt-4">
                    <ScreenRecorder 
                        spaceId={spaceId}
                        onUploadSuccess={handleUploadSuccess}
                        onUploadError={handleUploadError}
                    />
                </div>

                <div className="modal-action">
                    <button 
                        className="btn"
                        onClick={handleClose}
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    )
} 