import React from 'react'
import { Layout } from "../components/layout/Layout";
import ListOfVideos from "../components/ListOfVideos";

export const VideosPage = () => {
  return (
    <Layout>
        <div className="space-y-8">
            <div className="text-center">
                <h1 className="text-3xl font-bold mb-4">Videos</h1>
                <p className="mb-4">View and manage your recorded videos</p>
            </div>
            <div className="flex justify-center">
                <ListOfVideos />
            </div>
        </div>
    </Layout>
    )
} 