import React, { useState, useRef } from 'react'
import { useStore } from '@nanostores/react'
import { $authToken } from '../auth/store/auth'

const FileUploader = ({ onUploadSuccess, onUploadError, spaceId = null }) => {
    const [selectedFile, setSelectedFile] = useState(null)
    const [title, setTitle] = useState('')
    const [description, setDescription] = useState('')
    const [isUploading, setIsUploading] = useState(false)
    const [uploadProgress, setUploadProgress] = useState(0)
    const [isDragOver, setIsDragOver] = useState(false)
    const [previewUrl, setPreviewUrl] = useState(null)
    const fileInputRef = useRef(null)
    const authToken = useStore($authToken)

    const supportedFormats = ['.mp4', '.mov', '.avi', '.webm', '.mkv', '.flv', '.wmv', '.m4v', '.3gp', '.ogv']
    const maxFileSize = 500 * 1024 * 1024 // 500MB in bytes

    const validateFile = (file) => {
        if (!file) return { valid: false, error: 'No file selected' }
        
        const fileName = file.name.toLowerCase()
        const fileExtension = '.' + fileName.split('.').pop()
        
        if (!supportedFormats.includes(fileExtension)) {
            return { 
                valid: false, 
                error: `Unsupported format. Supported: ${supportedFormats.join(', ')}` 
            }
        }
        
        if (file.size > maxFileSize) {
            return { 
                valid: false, 
                error: `File too large. Maximum size: 500MB (current: ${(file.size / 1024 / 1024).toFixed(1)}MB)` 
            }
        }
        
        return { valid: true }
    }

    const handleFileSelect = (file) => {
        const validation = validateFile(file)
        if (!validation.valid) {
            onUploadError && onUploadError(new Error(validation.error))
            return
        }

        setSelectedFile(file)
        setTitle(file.name.replace(/\.[^/.]+$/, "")) // Remove extension for default title
        
        // Create preview URL for video
        const url = URL.createObjectURL(file)
        setPreviewUrl(url)
    }

    const handleDrop = (e) => {
        e.preventDefault()
        setIsDragOver(false)
        
        const files = Array.from(e.dataTransfer.files)
        if (files.length > 0) {
            handleFileSelect(files[0])
        }
    }

    const handleDragOver = (e) => {
        e.preventDefault()
        setIsDragOver(true)
    }

    const handleDragLeave = (e) => {
        e.preventDefault()
        setIsDragOver(false)
    }

    const handleFileInputChange = (e) => {
        if (e.target.files.length > 0) {
            handleFileSelect(e.target.files[0])
        }
    }

    const uploadFile = async () => {
        if (!selectedFile || !title.trim()) {
            onUploadError && onUploadError(new Error('Please select a file and enter a title'))
            return
        }

        setIsUploading(true)
        setUploadProgress(0)

        const formData = new FormData()
        formData.append('video', selectedFile)
        formData.append('title', title.trim())
        formData.append('description', description.trim())
        
        if (spaceId) {
            formData.append('space_id', spaceId)
        }

        try {
            const xhr = new XMLHttpRequest()
            
            // Track upload progress
            xhr.upload.onprogress = (event) => {
                if (event.lengthComputable) {
                    const progress = (event.loaded / event.total) * 100
                    setUploadProgress(Math.round(progress))
                }
            }

            // Handle completion
            xhr.onload = () => {
                if (xhr.status === 200) {
                    const response = JSON.parse(xhr.responseText)
                    onUploadSuccess && onUploadSuccess({
                        message: response.message,
                        videoId: response.video_id,
                        assignedToSpace: !!spaceId
                    })
                    
                    // Reset form
                    setSelectedFile(null)
                    setTitle('')
                    setDescription('')
                    setPreviewUrl(null)
                    setUploadProgress(0)
                    if (fileInputRef.current) {
                        fileInputRef.current.value = ''
                    }
                } else {
                    throw new Error(`Upload failed with status: ${xhr.status}`)
                }
            }

            xhr.onerror = () => {
                throw new Error('Upload failed due to network error')
            }

            xhr.open('POST', `${import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, "")}/api/videoservice/upload`)
            xhr.setRequestHeader('authorization', authToken)
            xhr.send(formData)

        } catch (error) {
            console.error('Error uploading file:', error)
            onUploadError && onUploadError(error)
            setUploadProgress(0)
        } finally {
            setIsUploading(false)
        }
    }

    const clearFile = () => {
        setSelectedFile(null)
        setTitle('')
        setDescription('')
        if (previewUrl) {
            URL.revokeObjectURL(previewUrl)
            setPreviewUrl(null)
        }
        if (fileInputRef.current) {
            fileInputRef.current.value = ''
        }
    }

    const formatFileSize = (bytes) => {
        if (bytes === 0) return '0 B'
        const k = 1024
        const sizes = ['B', 'KB', 'MB', 'GB']
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
    }

    return (
        <div className="space-y-4">
            {/* {spaceId && (
                <div className="alert alert-info">
                    <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                    </svg>
                    <span>Video will be automatically added to this space when uploaded.</span>
                </div>
            )} */}

            {!selectedFile ? (
                <div
                    className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                        isDragOver ? 'border-primary bg-primary/10' : 'border-base-300 hover:border-primary/50'
                    }`}
                    onDrop={handleDrop}
                    onDragOver={handleDragOver}
                    onDragLeave={handleDragLeave}
                >
                    <div className="flex flex-col items-center space-y-4">
                        <svg className="w-12 h-12 text-base-content/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                        </svg>
                        <div>
                            <p className="text-lg font-semibold">Drop your video here</p>
                            <p className="text-sm text-base-content/70">or click to browse</p>
                        </div>
                        <div className="text-xs text-base-content/60">
                            <p>Supported formats: {supportedFormats.join(', ')}</p>
                            <p>Maximum size: 500MB</p>
                        </div>
                        <button 
                            className="btn btn-primary"
                            onClick={() => fileInputRef.current?.click()}
                            disabled={isUploading}
                        >
                            Select Video File
                        </button>
                    </div>
                    
                    <input
                        ref={fileInputRef}
                        type="file"
                        accept={supportedFormats.join(',')}
                        onChange={handleFileInputChange}
                        className="hidden"
                    />
                </div>
            ) : (
                <div className="space-y-4">
                    {/* File Info */}
                    <div className="card bg-base-200 shadow">
                        <div className="card-body p-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-3">
                                    <svg className="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                    </svg>
                                    <div>
                                        <p className="font-semibold">{selectedFile.name}</p>
                                        <p className="text-sm text-base-content/70">{formatFileSize(selectedFile.size)}</p>
                                    </div>
                                </div>
                                <button 
                                    className="btn btn-ghost btn-sm"
                                    onClick={clearFile}
                                    disabled={isUploading}
                                >
                                    ✕
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Video Preview */}
                    {previewUrl && (
                        <div className="flex justify-center">
                            <video 
                                src={previewUrl} 
                                controls 
                                className="max-w-md w-full rounded-lg shadow-lg"
                                style={{ maxHeight: '300px' }}
                            />
                        </div>
                    )}

                    {/* Upload Form */}
                    <div className="space-y-3">
                        <div className="form-control">
                            <label className="label">
                                <span className="label-text">Title *</span>
                            </label>
                            <input
                                type="text"
                                value={title}
                                onChange={(e) => setTitle(e.target.value)}
                                placeholder="Enter video title"
                                className="input input-bordered w-full"
                                disabled={isUploading}
                            />
                        </div>

                        <div className="form-control">
                            <label className="label">
                                <span className="label-text">Description</span>
                            </label>
                            <textarea
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="Enter video description (optional)"
                                className="textarea textarea-bordered w-full"
                                disabled={isUploading}
                            />
                        </div>

                        {/* Upload Progress */}
                        {isUploading && (
                            <div className="space-y-2">
                                <div className="flex justify-between text-sm">
                                    <span>Uploading...</span>
                                    <span>{uploadProgress}%</span>
                                </div>
                                <progress className="progress progress-primary w-full" value={uploadProgress} max="100"></progress>
                            </div>
                        )}

                        {/* Upload Button */}
                        <button
                            className="btn btn-primary w-full"
                            onClick={uploadFile}
                            disabled={isUploading || !title.trim()}
                        >
                            {isUploading ? (
                                <>
                                    <span className="loading loading-spinner loading-sm"></span>
                                    Uploading... {uploadProgress}%
                                </>
                            ) : (
                                spaceId ? "Upload to Space" : "Upload Video"
                            )}
                        </button>
                    </div>
                </div>
            )}
        </div>
    )
}

export default FileUploader 