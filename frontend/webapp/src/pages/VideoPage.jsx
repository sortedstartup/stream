import React from 'react'
import { useState, useEffect } from 'react'
import { useParams } from 'react-router'
import { fetchVideo } from '../stores/videos';
import CommentSection from "../components/CommentSection";
import { Layout } from '../components/layout/Layout'

const SimpleVideoPlayer = ({ videoUrl }) => {
    return (
        <div className="w-full max-w-4xl mx-auto bg-base-200 rounded-lg overflow-hidden">
            <video
                className="w-full aspect-video"
                controls
                src={videoUrl}
            >
                Your browser does not support the video tag.
            </video>
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
                    <SimpleVideoPlayer videoUrl={`${import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, "")}/api/videoservice/video/${id}`} />
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