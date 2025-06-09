import React, { useState } from 'react'
import FileUploader from './FileUploader'

const UploadToSpaceModal = ({ isOpen, onClose, spaceName, spaceId, onUploadSuccess }) => {
    const [uploadInProgress, setUploadInProgress] = useState(false)

    const handleUploadSuccess = (result) => {
        onUploadSuccess && onUploadSuccess(result)
        onClose()
    }

    const handleUploadError = (error) => {
        console.error('Upload error:', error)
        alert(`Upload failed: ${error.message}`)
        setUploadInProgress(false)
    }

    if (!isOpen) return null

    return (
        <div className="modal modal-open">
            <div className="modal-box max-w-2xl">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="font-bold text-lg">Upload to {spaceName}</h3>
                    <button 
                        className="btn btn-sm btn-circle btn-ghost"
                        onClick={onClose}
                        disabled={uploadInProgress}
                    >
                        ✕
                    </button>
                </div>
                
                <div className="space-y-4">
                    <div className="alert alert-info">
                        <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                        </svg>
                        <span>Videos will be automatically added to <strong>{spaceName}</strong> when uploaded.</span>
                    </div>

                    <FileUploader
                        onUploadSuccess={handleUploadSuccess}
                        onUploadError={handleUploadError}
                        spaceId={spaceId}
                    />
                </div>
            </div>
            <div className="modal-backdrop" onClick={onClose} />
        </div>
    )
}

export default UploadToSpaceModal 