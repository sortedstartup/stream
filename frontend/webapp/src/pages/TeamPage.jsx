import React, { useState, useEffect } from 'react'
import { Layout } from "../components/layout/Layout"
import { useStore } from '@nanostores/react'
import { $spaces } from '../stores/spaces'
import { $authToken } from "../auth/store/auth"
import { useNavigate } from 'react-router'

export const TeamPage = () => {
    const spaces = useStore($spaces)
    const authToken = useStore($authToken)
    const navigate = useNavigate()
    
    // Extract current user ID from auth token
    const currentUserId = authToken ? JSON.parse(atob(authToken.split('.')[1])).user_id : null
    
    // Separate owned and shared spaces
    const ownedSpaces = spaces.filter(space => space.user_id === currentUserId)
    const sharedSpaces = spaces.filter(space => space.user_id !== currentUserId)

    return (
        <Layout>
            <div className="space-y-8">
                <div>
                    <h1 className="text-3xl font-bold mb-2">Team & Collaboration</h1>
                    <p className="text-base-content/70">Manage shared spaces and collaborate with team members</p>
                </div>

                {/* Shared Spaces Section */}
                <div>
                    <h2 className="text-2xl font-semibold mb-4">Spaces Shared With You ({sharedSpaces.length})</h2>
                    
                    {sharedSpaces.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {sharedSpaces.map((space) => (
                                <div 
                                    key={space.id} 
                                    className="card bg-base-100 shadow-md hover:shadow-lg transition-shadow cursor-pointer"
                                    onClick={() => navigate(`/spaces/${space.id}`)}
                                >
                                    <div className="card-body p-4">
                                        <div className="flex items-center gap-3">
                                            <div className="w-10 h-10 bg-secondary/20 rounded-lg flex items-center justify-center">
                                                <svg className="w-5 h-5 text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
                                                </svg>
                                            </div>
                                            <div className="flex-1">
                                                <h3 className="font-semibold">{space.name}</h3>
                                                <p className="text-sm text-base-content/60">
                                                    Shared by owner
                                                </p>
                                            </div>
                                            <span className="badge badge-secondary badge-sm">Shared</span>
                                        </div>
                                        {space.description && (
                                            <p className="text-sm text-base-content/70 mt-2 line-clamp-2">
                                                {space.description}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center py-12 bg-base-200 rounded-lg">
                            <div className="w-16 h-16 bg-base-300 rounded-full flex items-center justify-center mx-auto mb-4">
                                <svg className="w-8 h-8 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                </svg>
                            </div>
                            <h3 className="text-lg font-medium mb-2">No shared spaces yet</h3>
                            <p className="text-base-content/60">Spaces shared with you will appear here</p>
                        </div>
                    )}
                </div>

                {/* Your Spaces for Sharing */}
                <div>
                    <h2 className="text-2xl font-semibold mb-4">Your Spaces ({ownedSpaces.length})</h2>
                    <p className="text-base-content/70 mb-4">
                        These are spaces you own. Click on any space to manage sharing and collaboration.
                    </p>
                    
                    {ownedSpaces.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {ownedSpaces.map((space) => (
                                <div 
                                    key={space.id} 
                                    className="card bg-base-100 shadow-md hover:shadow-lg transition-shadow cursor-pointer"
                                    onClick={() => navigate(`/spaces/${space.id}`)}
                                >
                                    <div className="card-body p-4">
                                        <div className="flex items-center gap-3">
                                            <div className="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center">
                                                <svg className="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                                </svg>
                                            </div>
                                            <div className="flex-1">
                                                <h3 className="font-semibold">{space.name}</h3>
                                                <p className="text-sm text-base-content/60">
                                                    Created {new Date(space.created_at?.seconds * 1000).toLocaleDateString()}
                                                </p>
                                            </div>
                                            <span className="badge badge-primary badge-sm">Owner</span>
                                        </div>
                                        {space.description && (
                                            <p className="text-sm text-base-content/70 mt-2 line-clamp-2">
                                                {space.description}
                                            </p>
                                        )}
                                        <div className="flex justify-between items-center mt-3">
                                            <span className="text-xs text-base-content/60">
                                                Click to manage sharing
                                            </span>
                                            <svg className="w-4 h-4 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                                            </svg>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center py-12 bg-base-200 rounded-lg">
                            <div className="w-16 h-16 bg-base-300 rounded-full flex items-center justify-center mx-auto mb-4">
                                <svg className="w-8 h-8 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                </svg>
                            </div>
                            <h3 className="text-lg font-medium mb-2">No spaces created yet</h3>
                            <p className="text-base-content/60 mb-4">Create a space to start collaborating with your team</p>
                            <button 
                                className="btn btn-primary"
                                onClick={() => navigate('/spaces')}
                            >
                                Create Your First Space
                            </button>
                        </div>
                    )}
                </div>

                {/* Help Section */}
                <div className="bg-info/10 border border-info/20 rounded-lg p-6">
                    <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
                        <svg className="w-5 h-5 text-info" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        How Space Sharing Works
                    </h3>
                    <div className="space-y-2 text-sm text-base-content/80">
                        <p>• <strong>Space Owners</strong> can share their spaces with other users</p>
                        <p>• <strong>Shared Users</strong> can view or edit space content based on their access level</p>
                        <p>• <strong>Access Levels:</strong></p>
                        <ul className="ml-4 space-y-1">
                            <li>- <span className="badge badge-info badge-sm mr-2">View</span> Can view videos and space content</li>
                            <li>- <span className="badge badge-warning badge-sm mr-2">Edit</span> Can add, remove, and modify videos</li>
                            <li>- <span className="badge badge-error badge-sm mr-2">Admin</span> Can manage space settings and members</li>
                        </ul>
                    </div>
                </div>
            </div>
        </Layout>
    )
} 