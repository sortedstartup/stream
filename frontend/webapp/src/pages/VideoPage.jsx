import React from 'react'
import { useState, useEffect } from 'react'
import { useParams } from 'react-router'
import { useStore } from '@nanostores/react'
import { fetchVideo } from '../stores/videos';
import { $authToken } from '../auth/store/auth';
import { $currentTenant } from '../stores/tenants';
import CommentSection from "../components/CommentSection";
import { Layout } from '../components/layout/Layout'

const SimpleVideoPlayer = ({ videoId, tenantId }) => {
    const videoUrl = `${import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, "")}/api/videoservice/video/${videoId}?tenant=${encodeURIComponent(tenantId)}`;
    
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
    );
};

export const VideoPage = () => {
    const { id } = useParams()
    const [video, setVideo] = useState(null)
    const currentTenant = useStore($currentTenant);

    useEffect(() => {
        fetchVideo(id).then(video=>{
            setVideo(video)
        })
    }, [id])

    if (!video) return <div>Loading...</div>

    if (!currentTenant?.tenant?.id) {
        return (
            <Layout>
                <div className="container mx-auto px-4 py-8">
                    <div className="text-center">
                        <p className="text-error">No tenant selected. Please select a tenant first.</p>
                    </div>
                </div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="container mx-auto px-4 py-8">
                <h1 className="text-2xl font-bold mb-6">{video.title}</h1>
                <SimpleVideoPlayer 
                    videoId={id} 
                    tenantId={currentTenant.tenant.id} 
                />
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