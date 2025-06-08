import React, { useState, useEffect } from 'react'
import { useStore } from '@nanostores/react'
import { $videos, fetchVideos } from '../stores/videos'
import { addVideoToSpace } from '../stores/spaces'

const VideoSelectCard = ({ video, isSelected, onToggle }) => {
    return (
        <div 
            className={`card bg-base-100 shadow-md cursor-pointer transition-all duration-200 ${
                isSelected ? 'ring-2 ring-primary shadow-lg' : 'hover:shadow-lg'
            }`}
            onClick={() => onToggle(video.id)}
        >
            {/* Thumbnail */}
            <figure className="relative aspect-video">
                {video.thumbnail_url ? (
                    <img 
                        src={video.thumbnail_url} 
                        alt={video.title}
                        className="w-full h-full object-cover"
                    />
                ) : (
                    <div className="absolute inset-0 flex items-center justify-center bg-base-200">
                        <svg className="w-8 h-8 text-base-content opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                    </div>
                )}
                
                {/* Selection indicator */}
                <div className="absolute top-2 right-2">
                    <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => onToggle(video.id)}
                        className="checkbox checkbox-primary"
                        onClick={(e) => e.stopPropagation()}
                    />
                </div>
            </figure>

            {/* Content */}
            <div className="card-body p-3">
                <h3 className="text-sm font-semibold text-base-content truncate">{video.title}</h3>
                <p className="text-xs text-base-content/70 line-clamp-2">{video.description}</p>
                <div className="text-xs text-base-content/60 mt-1">
                    {new Date(video.created_at?.seconds * 1000).toLocaleDateString()}
                </div>
            </div>
        </div>
    )
}

export const AddVideosToSpaceModal = ({ isOpen, onClose, spaceId, spaceName, onVideosAdded }) => {
    const videos = useStore($videos)
    const [selectedVideoIds, setSelectedVideoIds] = useState(new Set())
    const [isLoading, setIsLoading] = useState(false)
    const [isAddingVideos, setIsAddingVideos] = useState(false)

    useEffect(() => {
        if (isOpen) {
            // Refresh videos when modal opens
            fetchVideos()
            setSelectedVideoIds(new Set())
        }
    }, [isOpen])

    const handleToggleVideo = (videoId) => {
        const newSelected = new Set(selectedVideoIds)
        if (newSelected.has(videoId)) {
            newSelected.delete(videoId)
        } else {
            newSelected.add(videoId)
        }
        setSelectedVideoIds(newSelected)
    }

    const handleSelectAll = () => {
        if (selectedVideoIds.size === videos.length) {
            setSelectedVideoIds(new Set())
        } else {
            setSelectedVideoIds(new Set(videos.map(v => v.id)))
        }
    }

    const handleAddVideos = async () => {
        if (selectedVideoIds.size === 0) return

        setIsAddingVideos(true)
        try {
            const promises = Array.from(selectedVideoIds).map(videoId => 
                addVideoToSpace(videoId, spaceId)
            )
            await Promise.all(promises)
            
            onVideosAdded && onVideosAdded(selectedVideoIds.size)
            onClose()
        } catch (error) {
            console.error('Failed to add videos to space:', error)
            // TODO: Show error message to user
        } finally {
            setIsAddingVideos(false)
        }
    }

    if (!isOpen) return null

    return (
        <div className="modal modal-open">
            <div className="modal-box max-w-4xl w-full">
                <h3 className="font-bold text-lg mb-4">
                    Add Videos to "{spaceName}"
                </h3>
                
                {videos.length > 0 ? (
                    <>
                        {/* Controls */}
                        <div className="flex justify-between items-center mb-4">
                            <div className="flex items-center gap-2">
                                <button
                                    type="button"
                                    className="btn btn-sm btn-outline"
                                    onClick={handleSelectAll}
                                >
                                    {selectedVideoIds.size === videos.length ? 'Deselect All' : 'Select All'}
                                </button>
                                <span className="text-sm text-base-content/70">
                                    {selectedVideoIds.size} of {videos.length} selected
                                </span>
                            </div>
                        </div>

                        {/* Videos Grid */}
                        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3 mb-6 max-h-96 overflow-y-auto">
                            {videos.map((video) => (
                                <VideoSelectCard
                                    key={video.id}
                                    video={video}
                                    isSelected={selectedVideoIds.has(video.id)}
                                    onToggle={handleToggleVideo}
                                />
                            ))}
                        </div>
                    </>
                ) : (
                    <div className="text-center py-8">
                        <div className="w-16 h-16 bg-base-300 rounded-full flex items-center justify-center mx-auto mb-4">
                            <svg className="w-8 h-8 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                            </svg>
                        </div>
                        <h3 className="text-lg font-semibold mb-2">No videos available</h3>
                        <p className="text-base-content/70">Record or upload some videos first</p>
                    </div>
                )}

                <div className="modal-action">
                    <button
                        type="button"
                        className="btn"
                        onClick={onClose}
                        disabled={isAddingVideos}
                    >
                        Cancel
                    </button>
                    {videos.length > 0 && (
                        <button
                            type="button"
                            className="btn btn-primary"
                            onClick={handleAddVideos}
                            disabled={isAddingVideos || selectedVideoIds.size === 0}
                        >
                            {isAddingVideos ? (
                                <>
                                    <span className="loading loading-spinner loading-sm"></span>
                                    Adding...
                                </>
                            ) : (
                                `Add ${selectedVideoIds.size} Video${selectedVideoIds.size !== 1 ? 's' : ''}`
                            )}
                        </button>
                    )}
                </div>
            </div>
            <div className="modal-backdrop" onClick={onClose} />
        </div>
    )
} 