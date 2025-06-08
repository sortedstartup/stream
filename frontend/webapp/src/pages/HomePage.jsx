import React from 'react'
import { Layout } from "../components/layout/Layout";
import { useNavigate } from 'react-router';

export const HomePage = () => {
    const navigate = useNavigate()
    return (
        <Layout>
            <div className="space-y-8">
                <div className="text-center">
                    <h1 className="text-3xl font-bold">Welcome to Stream</h1>
                    <p className="mt-2 text-base-content/70">Record, upload, and manage your videos easily</p>
                </div>

                <div className="flex justify-center gap-4">
                    <button onClick={() => navigate('/record')} className="btn btn-primary btn-lg gap-2">
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                            <circle cx="12" cy="12" r="10" strokeWidth="2"/>
                            <circle cx="12" cy="12" r="3" fill="currentColor"/>
                        </svg>
                        Start Recording
                    </button>
                    
                    <button onClick={() => navigate('/upload')} className="btn btn-secondary btn-lg gap-2">
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                        </svg>
                        Upload Video
                    </button>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-4xl mx-auto">
                    <div className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow">
                        <div className="card-body">
                            <div className="text-primary mb-4">
                                <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                </svg>
                            </div>
                            <h3 className="card-title">Quick Recording</h3>
                            <p className="text-base-content/70">Start recording your screen with just one click. Choose to record full screen or select a specific area.</p>
                        </div>
                    </div>

                    <div className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow">
                        <div className="card-body">
                            <div className="text-secondary mb-4">
                                <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                                </svg>
                            </div>
                            <h3 className="card-title">Upload Videos</h3>
                            <p className="text-base-content/70">Upload existing videos from your device. Supports multiple formats up to 500MB.</p>
                        </div>
                    </div>

                    <div className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow">
                        <div className="card-body">
                            <div className="text-primary mb-4">
                                <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                </svg>
                            </div>
                            <h3 className="card-title">Organize with Spaces</h3>
                            <p className="text-base-content/70">Create spaces to organize your videos by project, topic, or team. Easy collaboration and management.</p>
                        </div>
                    </div>

                    <div className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow">
                        <div className="card-body">
                            <div className="text-primary mb-4">
                                <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                                </svg>
                            </div>
                            <h3 className="card-title">Easy Sharing</h3>
                            <p className="text-base-content/70">Share your videos instantly with a generated link or organize them in spaces for team access.</p>
                        </div>
                    </div>
                </div>
            </div>
        </Layout>
    )
}