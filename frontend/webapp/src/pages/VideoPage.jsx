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
    const mediaSourceRef = useRef(null)
    const [isPlaying, setIsPlaying] = useState(false)
    const [currentTime, setCurrentTime] = useState(0)
    const [duration, setDuration] = useState(0)
    const [volume, setVolume] = useState(1)
    const [supportsMediaSource, setSupportsMediaSource] = useState(false)
    const authToken = useStore($authToken)

    useEffect(() => {
        // Check MediaSource support and determine if we need to use it
        setSupportsMediaSource('MediaSource' in window)
    }, [])

    const getVideoMimeType = (url) => {
        const extension = url.split('.').pop().toLowerCase()
        switch (extension) {
            case 'webm':
                return 'video/webm; codecs="vp8,opus"'
            case 'mp4':
                return 'video/mp4; codecs="avc1.42E01E,mp4a.40.2"'
            case 'mov':
                return 'video/mp4; codecs="avc1.42E01E,mp4a.40.2"'
            case 'avi':
                return 'video/mp4; codecs="avc1.42E01E,mp4a.40.2"'
            default:
                return 'video/mp4; codecs="avc1.42E01E,mp4a.40.2"'
        }
    }

    const setupMediaSource = () => {
        return new Promise((resolve, reject) => {
            // Create new MediaSource instance
            mediaSourceRef.current = new MediaSource()
            const mediaUrl = URL.createObjectURL(mediaSourceRef.current)
            videoRef.current.src = mediaUrl

            mediaSourceRef.current.addEventListener('sourceopen', () => {
                // Fetch video with authentication
                fetch(videoUrl, {
                    headers: {
                        'authorization': `${authToken}`,
                    }
                })
                .then(response => response.blob())
                .then(async videoBlob => {
                    const mimeType = getVideoMimeType(videoUrl)
                    
                    // Check if the MIME type is supported
                    if (!MediaSource.isTypeSupported(mimeType)) {
                        console.warn(`MIME type ${mimeType} not supported, falling back to direct video src`)
                        reject(new Error('MIME type not supported'))
                        return
                    }

                    const sourceBuffer = mediaSourceRef.current.addSourceBuffer(mimeType)
                    sourceBuffer.addEventListener('updateend', () => {
                         if (mediaSourceRef.current.readyState === 'open') {
                                try {
                                    const end = videoRef.current.buffered.end(videoRef.current.buffered.length - 1);
                                    mediaSourceRef.current.duration = end;
                                    setDuration(end);
                                } catch (err) {
                                    console.error('Error setting duration:', err);
                                }
                                mediaSourceRef.current.endOfStream();
                                resolve(); // Resolve the promise when setup is complete
                            }
                    })
                    sourceBuffer.appendBuffer(await videoBlob.arrayBuffer())
                })
                .catch(error => {
                    console.error('Error fetching video:', error)
                    reject(error)
                })
            })

            mediaSourceRef.current.addEventListener('error', (e) => {
                console.error('MediaSource error:', e)
                reject(e)
            })
        }) 
    }

    const setupDirectVideo = () => {
        return new Promise((resolve, reject) => {
            // Fetch video with authentication and create blob URL
            fetch(videoUrl, {
                headers: {
                    'authorization': `${authToken}`,
                }
            })
            .then(response => response.blob())
            .then(videoBlob => {
                const blobUrl = URL.createObjectURL(videoBlob)
                videoRef.current.src = blobUrl
                videoRef.current.addEventListener('loadedmetadata', () => {
                    setDuration(videoRef.current.duration)
                    resolve()
                }, { once: true })
                videoRef.current.load()
            })
            .catch(error => {
                console.error('Error fetching video:', error)
                reject(error)
            })
        })
    }

    const togglePlay = async () => {
        if (videoRef.current.paused) {
            // Use MediaSource for WebM files to maintain compatibility with screen recordings
            if (!mediaSourceRef.current && !videoRef.current.src && supportsMediaSource && videoUrl.includes('.webm')) {
                try {
                    await setupMediaSource()
                } catch (error) {
                    console.log('MediaSource setup failed, falling back to direct video:', error)
                    await setupDirectVideo()
                }
            } else if (!videoRef.current.src) {
                // For MP4 and other formats, use direct blob approach
                await setupDirectVideo()
            }
            
            await videoRef.current.play()
            setIsPlaying(true)
        } else {
            videoRef.current.pause()
            setIsPlaying(false)
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
            const formatTimeDuration = videoRef.current?.duration;
            if (isFinite(formatTimeDuration)) time = formatTimeDuration;
            else return '0:00';
        }
        const minutes = Math.floor(time / 60)
        const seconds = Math.floor(time % 60)
        return `${minutes}:${seconds.toString().padStart(2, '0')}`
    }

    return (
        <div className="w-full max-w-4xl mx-auto bg-base-200 rounded-lg overflow-hidden">
            <video
                ref={videoRef}
                className="w-full aspect-video"
                onTimeUpdate={handleTimeUpdate}
                onLoadedMetadata={handleLoadedMetadata}
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