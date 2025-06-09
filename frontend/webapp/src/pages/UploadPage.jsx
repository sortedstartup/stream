import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import FileUploader from '../components/FileUploader'
import { fetchSpaces, $spaces } from '../stores/spaces'
import { useStore } from '@nanostores/react'
import { Layout } from '../components/layout/Layout'

const UploadPage = () => {
    const [showSuccessMessage, setShowSuccessMessage] = useState(false)
    const [uploadMessage, setUploadMessage] = useState('')
    const [selectedSpaceId, setSelectedSpaceId] = useState('')
    const [isLoadingSpaces, setIsLoadingSpaces] = useState(true)
    const navigate = useNavigate()
    const spaces = useStore($spaces)

    useEffect(() => {
        loadUserSpaces()
    }, [])

    const loadUserSpaces = async () => {
        try {
            setIsLoadingSpaces(true)
            await fetchSpaces()
        } catch (error) {
            console.error('Error loading spaces:', error)
        } finally {
            setIsLoadingSpaces(false)
        }
    }

    const handleUploadSuccess = (result) => {
        setUploadMessage(result.message)
        setShowSuccessMessage(true)
        
        // Hide success message after 5 seconds
        setTimeout(() => {
            setShowSuccessMessage(false)
        }, 5000)
    }

    const handleUploadError = (error) => {
        console.error('Upload error:', error)
        // You could add error toast here
        alert(`Upload failed: ${error.message}`)
    }

    return (
        <Layout>
            <div className="container mx-auto p-4 max-w-2xl">
                <div className="space-y-6">
                    {/* Header */}
                    <div className="text-center">
                        <h1 className="text-3xl font-bold">Upload Video</h1>
                        <p className="text-base-content/70 mt-2">
                            Upload videos from your device to your library
                        </p>
                    </div>

                    {/* Success Message */}
                    {showSuccessMessage && (
                        <div className="alert alert-success">
                            <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                            </svg>
                            <span>{uploadMessage}</span>
                        </div>
                    )}

                    {/* Space Selection */}
                    <div className="card bg-base-100 shadow">
                        <div className="card-body">
                            <h3 className="card-title text-lg">Upload Options</h3>
                            
                            <div className="form-control">
                                <label className="label">
                                    <span className="label-text">Assign to Space (Optional)</span>
                                </label>
                                
                                {isLoadingSpaces ? (
                                    <div className="flex items-center space-x-2">
                                        <span className="loading loading-spinner loading-sm"></span>
                                        <span>Loading your spaces...</span>
                                    </div>
                                ) : (
                                    <select 
                                        className="select select-bordered w-full"
                                        value={selectedSpaceId}
                                        onChange={(e) => setSelectedSpaceId(e.target.value)}
                                    >
                                        <option value="">No space (upload to library only)</option>
                                        {spaces.map(space => (
                                            <option key={space.id} value={space.id}>
                                                {space.name}
                                            </option>
                                        ))}
                                    </select>
                                )}
                                
                                <div className="label">
                                    <span className="label-text-alt">
                                        {selectedSpaceId ? 'Video will be added to the selected space' : 'Video will only be added to your library'}
                                    </span>
                                </div>
                            </div>

                            {spaces.length === 0 && !isLoadingSpaces && (
                                <div className="alert alert-info">
                                    <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                                    </svg>
                                    <div>
                                        <span>You don't have any spaces yet.</span>
                                        <button 
                                            className="btn btn-sm btn-primary ml-2"
                                            onClick={() => navigate('/spaces')}
                                        >
                                            Create a Space
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* File Uploader */}
                    <div className="card bg-base-100 shadow">
                        <div className="card-body">
                            <h3 className="card-title text-lg">Select Video File</h3>
                            <FileUploader
                                onUploadSuccess={handleUploadSuccess}
                                onUploadError={handleUploadError}
                                spaceId={selectedSpaceId || null}
                            />
                        </div>
                    </div>

                    {/* Alternative Options */}
                    <div className="card bg-base-100 shadow">
                        <div className="card-body">
                            <h3 className="card-title text-lg">Other Options</h3>
                            <div className="flex flex-col sm:flex-row gap-3">
                                <button 
                                    className="btn btn-outline"
                                    onClick={() => navigate('/record')}
                                >
                                    <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                    </svg>
                                    Record Screen Instead
                                </button>
                                
                                <button 
                                    className="btn btn-outline"
                                    onClick={() => navigate('/videos')}
                                >
                                    <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                    </svg>
                                    View My Videos
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </Layout>
    )
}

export { UploadPage } 