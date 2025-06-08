import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router'
import { useStore } from '@nanostores/react'
import { Layout } from "../components/layout/Layout"
import { $spaceVideos, fetchSpace, fetchVideosInSpace } from '../stores/spaces'
import { VideoStatus, Visibility } from '../proto/videoservice'
import { AddVideosToSpaceModal } from '../components/AddVideosToSpaceModal'
import { RecordToSpaceModal } from '../components/RecordToSpaceModal'
import UploadToSpaceModal from '../components/UploadToSpaceModal'

const VideoCard = ({ video }) => {
    const navigate = useNavigate()

    const getStatusBadge = (status) => {
        const statusClasses = {
            [VideoStatus.STATUS_PROCESSING]: "badge badge-warning",
            [VideoStatus.STATUS_READY]: "badge badge-success",
            [VideoStatus.STATUS_FAILED]: "badge badge-error",
            [VideoStatus.STATUS_UNSPECIFIED]: "badge badge-ghost",
        }
        
        const statusText = VideoStatus[status].replace('STATUS_', '')
        return (
            <span className={statusClasses[status]}>
                {statusText}
            </span>
        )
    }

    return (
        <div 
            className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow duration-300 cursor-pointer" 
            onClick={() => navigate(`/video/${video.id}`)}
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
                    <div className="absolute inset-0 flex items-center justify-center">
                        <svg className="w-12 h-12 text-base-content opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                    </div>
                )}
                {/* Status Badge */}
                <div className="absolute top-2 right-2">
                    {getStatusBadge(video.status)}
                </div>
            </figure>

            {/* Content */}
            <div className="card-body p-4">
                <h3 className="card-title text-base-content truncate">{video.title}</h3>
                <p className="text-sm text-base-content/70 line-clamp-2">{video.description}</p>
                
                {/* Footer */}
                <div className="flex items-center justify-between mt-4">
                    <div className="flex items-center gap-2">
                        <span className="text-xs text-base-content/60">
                            {new Date(video.created_at?.seconds * 1000).toLocaleDateString()}
                        </span>
                        {video.visibility !== Visibility.VISIBILITY_PRIVATE && (
                            <span className="badge badge-primary">
                                {video.visibility === Visibility.VISIBILITY_SHARED ? 'Shared' : 'Public'}
                            </span>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}

export const SpaceDetailPage = () => {
    const { spaceId } = useParams()
    const navigate = useNavigate()
    const spaceVideos = useStore($spaceVideos)
    const [space, setSpace] = useState(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [isAddVideosModalOpen, setIsAddVideosModalOpen] = useState(false)
    const [isRecordModalOpen, setIsRecordModalOpen] = useState(false)
    const [isUploadModalOpen, setIsUploadModalOpen] = useState(false)
    const [showDropdown, setShowDropdown] = useState(false)

    const videos = spaceVideos[spaceId] || []

    useEffect(() => {
        const loadSpaceData = async () => {
            try {
                setLoading(true)
                const [spaceData] = await Promise.all([
                    fetchSpace(spaceId),
                    fetchVideosInSpace(spaceId)
                ])
                setSpace(spaceData)
            } catch (err) {
                console.error('Error loading space data:', err)
                setError('Failed to load space')
            } finally {
                setLoading(false)
            }
        }

        if (spaceId) {
            loadSpaceData()
        }
    }, [spaceId])

    const handleVideosAdded = (count) => {
        // Refresh the videos in the space
        fetchVideosInSpace(spaceId)
        console.log(`${count} video(s) added to space`)
    }

    const handleVideoRecorded = (result) => {
        // Refresh the videos in the space
        fetchVideosInSpace(spaceId)
        console.log('New video recorded and added to space:', result.videoId)
    }

    const handleVideoUploaded = (result) => {
        // Refresh the videos in the space
        fetchVideosInSpace(spaceId)
        console.log('New video uploaded and added to space:', result.videoId)
    }

    if (loading) {
        return (
            <Layout>
                <div className="flex justify-center items-center py-16">
                    <span className="loading loading-spinner loading-lg"></span>
                </div>
            </Layout>
        )
    }

    if (error || !space) {
        return (
            <Layout>
                <div className="text-center py-16">
                    <h2 className="text-2xl font-bold mb-4">Space not found</h2>
                    <p className="text-base-content/70 mb-6">The space you're looking for doesn't exist or you don't have access to it.</p>
                    <button 
                        className="btn btn-primary"
                        onClick={() => navigate('/spaces')}
                    >
                        Back to Spaces
                    </button>
                </div>
            </Layout>
        )
    }

    return (
        <Layout>
            <div className="space-y-8">
                {/* Header */}
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                    <div>
                        <div className="flex items-center gap-3 mb-2">
                            <button 
                                onClick={() => navigate('/spaces')}
                                className="btn btn-ghost btn-sm"
                            >
                                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                                </svg>
                                Back to Spaces
                            </button>
                        </div>
                        <h1 className="text-3xl font-bold mb-2">{space.name}</h1>
                        {space.description && (
                            <p className="text-base-content/70">{space.description}</p>
                        )}
                        <p className="text-sm text-base-content/60 mt-2">
                            Created {new Date(space.created_at?.seconds * 1000).toLocaleDateString()}
                        </p>
                    </div>
                    
                    <div className="flex gap-2 relative">
                        <div className="dropdown dropdown-end">
                            <button 
                                className="btn btn-primary"
                                onClick={() => setShowDropdown(!showDropdown)}
                            >
                                <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                                </svg>
                                Add Videos
                                <svg className="w-4 h-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                                </svg>
                            </button>
                            {showDropdown && (
                                <ul className="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52 mt-1">
                                    <li>
                                        <button 
                                            onClick={() => {
                                                setIsRecordModalOpen(true)
                                                setShowDropdown(false)
                                            }}
                                            className="flex items-center gap-2"
                                        >
                                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                            </svg>
                                            Record New Video
                                        </button>
                                    </li>
                                    <li>
                                        <button 
                                            onClick={() => {
                                                setIsUploadModalOpen(true)
                                                setShowDropdown(false)
                                            }}
                                            className="flex items-center gap-2"
                                        >
                                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                            </svg>
                                            Upload Video File
                                        </button>
                                    </li>
                                    <li>
                                        <button 
                                            onClick={() => {
                                                setIsAddVideosModalOpen(true)
                                                setShowDropdown(false)
                                            }}
                                            className="flex items-center gap-2"
                                        >
                                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                            </svg>
                                            Add Existing Videos
                                        </button>
                                    </li>
                                </ul>
                            )}
                        </div>
                    </div>
                </div>

                {/* Videos Grid */}
                {videos.length > 0 ? (
                    <div>
                        <h2 className="text-xl font-semibold mb-4">Videos ({videos.length})</h2>
                        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
                            {videos.map((video) => (
                                <div key={video.id} className="max-w-[225px] mx-auto w-full">
                                    <VideoCard video={video} />
                                </div>
                            ))}
                        </div>
                    </div>
                ) : (
                    <div className="text-center py-16">
                        <div className="w-24 h-24 bg-base-300 rounded-full flex items-center justify-center mx-auto mb-4">
                            <svg className="w-12 h-12 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                            </svg>
                        </div>
                        <h3 className="text-xl font-semibold mb-2">No videos in this space</h3>
                        <p className="text-base-content/70 mb-6">Add some videos to get started</p>
                        <div className="flex gap-2 justify-center">
                            <button 
                                className="btn btn-primary"
                                onClick={() => setIsRecordModalOpen(true)}
                            >
                                Record Video
                            </button>
                            <button 
                                className="btn btn-secondary"
                                onClick={() => setIsUploadModalOpen(true)}
                            >
                                Upload Video
                            </button>
                            <button 
                                className="btn btn-outline"
                                onClick={() => setIsAddVideosModalOpen(true)}
                            >
                                Add Existing Videos
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Modals */}
            <AddVideosToSpaceModal
                isOpen={isAddVideosModalOpen}
                onClose={() => setIsAddVideosModalOpen(false)}
                spaceId={spaceId}
                spaceName={space?.name}
                onVideosAdded={handleVideosAdded}
            />

            <RecordToSpaceModal
                isOpen={isRecordModalOpen}
                onClose={() => setIsRecordModalOpen(false)}
                spaceId={spaceId}
                spaceName={space?.name}
                onVideoRecorded={handleVideoRecorded}
            />

            <UploadToSpaceModal
                isOpen={isUploadModalOpen}
                onClose={() => setIsUploadModalOpen(false)}
                spaceId={spaceId}
                spaceName={space?.name}
                onUploadSuccess={handleVideoUploaded}
            />

            {/* Close dropdown when clicking outside */}
            {showDropdown && (
                <div 
                    className="fixed inset-0 z-0" 
                    onClick={() => setShowDropdown(false)}
                />
            )}
        </Layout>
    )
} 