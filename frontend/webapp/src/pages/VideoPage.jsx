import React from 'react'
import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router'
import { $authToken } from "../auth/store/auth";
//nanostores
import { useStore } from '@nanostores/react'
import { fetchVideo } from '../stores/videos';
import {Video} from "../proto/videoservice"
import CommentSection from "../components/CommentSection";
import { Layout } from '../components/layout/Layout'

const CustomVideoPlayer = ({ videoUrl }) => {
    const videoRef = useRef(null)
    const [isPlaying, setIsPlaying] = useState(false)
    const [currentTime, setCurrentTime] = useState(0)
    const [duration, setDuration] = useState(0)
    const [volume, setVolume] = useState(1)
    const [videoSrc, setVideoSrc] = useState(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState(null)
    const authToken = useStore($authToken)

    useEffect(() => {
        // Create an authenticated video source using a blob URL
        const setupVideo = async () => {
            try {
                setIsLoading(true)
                setError(null)
                
                const response = await fetch(videoUrl, {
                    headers: {
                        'authorization': `${authToken}`,
                    }
                })
                
                if (!response.ok) {
                    throw new Error(`Failed to load video: ${response.status}`)
                }
                
                const videoBlob = await response.blob()
                const videoBlobUrl = URL.createObjectURL(videoBlob)
                setVideoSrc(videoBlobUrl)
            } catch (err) {
                console.error('Error loading video:', err)
                setError(err.message)
            } finally {
                setIsLoading(false)
            }
        }

        setupVideo()

        // Cleanup blob URL when component unmounts
        return () => {
            if (videoSrc) {
                URL.revokeObjectURL(videoSrc)
            }
        }
    }, [videoUrl, authToken])

    const togglePlay = async () => {
        try {
            if (videoRef.current.paused) {
                await videoRef.current.play()
                setIsPlaying(true)
            } else {
                videoRef.current.pause()
                setIsPlaying(false)
            }
        } catch (err) {
            console.error('Error playing video:', err)
            setError('Failed to play video')
        }
    }

    const handleTimeUpdate = () => {
        setCurrentTime(videoRef.current.currentTime)
    }

    const handleLoadedMetadata = () => {
        setDuration(videoRef.current.duration)
    }

    const handleSeek = (e) => {
        const time = e.target.value
        videoRef.current.currentTime = time
        setCurrentTime(time)
    }

    const handleVolumeChange = (e) => {
        const value = e.target.value
        setVolume(value)
        videoRef.current.volume = value
    }

    const formatTime = (time) => {
        if (!isFinite(time)) {
            return '0:00'
        }
        const minutes = Math.floor(time / 60)
        const seconds = Math.floor(time % 60)
        return `${minutes}:${seconds.toString().padStart(2, '0')}`
    }

    if (isLoading) {
        return (
            <div className="w-full max-w-4xl mx-auto bg-base-200 rounded-lg overflow-hidden">
                <div className="w-full aspect-video flex items-center justify-center">
                    <span className="loading loading-spinner loading-lg"></span>
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div className="w-full max-w-4xl mx-auto bg-base-200 rounded-lg overflow-hidden">
                <div className="w-full aspect-video flex items-center justify-center bg-error/10">
                    <div className="text-center">
                        <svg className="w-12 h-12 text-error mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <p className="text-error">Error loading video</p>
                        <p className="text-sm text-error/70">{error}</p>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="w-full max-w-4xl mx-auto bg-base-200 rounded-lg overflow-hidden">
            <video
                ref={videoRef}
                src={videoSrc}
                className="w-full aspect-video"
                onTimeUpdate={handleTimeUpdate}
                onLoadedMetadata={handleLoadedMetadata}
                onError={(e) => {
                    console.error('Video error:', e)
                    setError('Video playback error')
                }}
                controls={false} // We'll use custom controls
            />
            
            <div className="p-4 space-y-2">
                {/* Progress bar */}
                <input
                    type="range"
                    min="0"
                    max={duration}
                    value={currentTime}
                    onChange={handleSeek}
                    className="w-full"
                />
                
                <div className="flex items-center justify-between">
                    {/* Play/Pause button */}
                    <button
                        onClick={togglePlay}
                        className="btn btn-primary"
                        disabled={!videoSrc}
                    >
                        {isPlaying ? (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        ) : (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        )}
                    </button>

                    {/* Time display */}
                    <div className="text-sm">
                        {formatTime(currentTime)} / {formatTime(duration)}
                    </div>

                    {/* Volume control */}
                    <div className="flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                        </svg>
                        <input
                            type="range"
                            min="0"
                            max="1"
                            step="0.1"
                            value={volume}
                            onChange={handleVolumeChange}
                            className="w-24"
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}

export const VideoPage = () => {
    const { id } = useParams()
    const [video, setVideo] = useState(null)

    useEffect(() => {
        fetchVideo(id).then(video=>{
            setVideo(video)
        })
    }, [id])

   if (!video) return <div>Loading...</div>

    return (
            <Layout>
                <div className="container mx-auto px-4 py-8">
                    <h1 className="text-2xl font-bold mb-6">{video.title}</h1>
                    <CustomVideoPlayer videoUrl={`${import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, "")}/api/videoservice/video/${id}`} />
                    <div className="mt-6">
                        <p className="text-base-content/70">{video.description}</p>
                        <div className="mt-4 text-sm text-base-content/60">
                            Uploaded on {new Date(video.created_at?.seconds * 1000).toLocaleDateString()}
                        </div>
                    </div>
                    <CommentSection />
                </div>
            </Layout>
    )
} 