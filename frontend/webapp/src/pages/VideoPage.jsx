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
import AuthVideo from '../components/ShakaAuthVideo'

export const VideoPage = () => {
    const { id } = useParams()
    const [video, setVideo] = useState(null)
    const authToken = useStore($authToken)

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
                    <AuthVideo 
                        src={`${import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, "")}/api/videoservice/video/${id}`}
                        token={authToken}
                        vjsOpts={{
                            responsive: true,
                            fluid: true,
                        }}
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