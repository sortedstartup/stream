import React, { useState } from 'react'
import { useStore } from '@nanostores/react'
import { Layout } from "../components/layout/Layout"
import { $spaces, createSpace } from '../stores/spaces'
import { useNavigate } from 'react-router'
import { $authToken } from "../auth/store/auth"

const CreateSpaceModal = ({ isOpen, onClose, onCreateSpace }) => {
    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [isLoading, setIsLoading] = useState(false)

    const handleSubmit = async (e) => {
        e.preventDefault()
        if (!name.trim()) return

        setIsLoading(true)
        try {
            await onCreateSpace(name.trim(), description.trim())
            setName('')
            setDescription('')
            onClose()
        } catch (error) {
            console.error('Failed to create space:', error)
        } finally {
            setIsLoading(false)
        }
    }

    if (!isOpen) return null

    return (
        <div className="modal modal-open">
            <div className="modal-box">
                <h3 className="font-bold text-lg mb-4">Create New Space</h3>
                <form onSubmit={handleSubmit}>
                    <div className="form-control mb-4">
                        <label className="label">
                            <span className="label-text">Space Name</span>
                        </label>
                        <input
                            type="text"
                            placeholder="Enter space name"
                            className="input input-bordered w-full"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            required
                        />
                    </div>
                    <div className="form-control mb-4">
                        <label className="label">
                            <span className="label-text">Description (optional)</span>
                        </label>
                        <textarea
                            placeholder="Enter space description"
                            className="textarea textarea-bordered w-full"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                        />
                    </div>
                    <div className="modal-action">
                        <button
                            type="button"
                            className="btn"
                            onClick={onClose}
                            disabled={isLoading}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="btn btn-primary"
                            disabled={isLoading || !name.trim()}
                        >
                            {isLoading ? 'Creating...' : 'Create Space'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}

const SpaceCard = ({ space, currentUserId }) => {
    const navigate = useNavigate()
    const isOwner = space.user_id === currentUserId
    const isShared = !isOwner

    return (
        <div 
            className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow duration-300 cursor-pointer" 
            onClick={() => navigate(`/spaces/${space.id}`)}
        >
            <div className="card-body">
                <div className="flex items-center gap-3 mb-2">
                    <div className={`w-12 h-12 ${isShared ? 'bg-secondary/20' : 'bg-primary/20'} rounded-lg flex items-center justify-center`}>
                        {isShared ? (
                            <svg className="w-6 h-6 text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
                            </svg>
                        ) : (
                            <svg className="w-6 h-6 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                            </svg>
                        )}
                    </div>
                    <div className="flex-1">
                        <div className="flex items-center gap-2">
                            <h3 className="card-title text-lg">{space.name}</h3>
                            {isShared && (
                                <span className="badge badge-secondary badge-sm">Shared</span>
                            )}
                        </div>
                        <p className="text-sm text-base-content/60">
                            {isShared ? 'Shared with you' : 'Created'} {new Date(space.created_at?.seconds * 1000).toLocaleDateString()}
                        </p>
                    </div>
                </div>
                
                {space.description && (
                    <p className="text-base-content/70 mb-4 line-clamp-2">
                        {space.description}
                    </p>
                )}
                
                <div className="card-actions justify-end">
                    <button className="btn btn-sm btn-primary">
                        View Space
                    </button>
                </div>
            </div>
        </div>
    )
}

export const SpacesPage = () => {
    const spaces = useStore($spaces)
    const authToken = useStore($authToken)
    const [isModalOpen, setIsModalOpen] = useState(false)

    // Extract user ID from auth token (Firebase UID)
    const currentUserId = authToken ? JSON.parse(atob(authToken.split('.')[1])).user_id : null

    const handleCreateSpace = async (name, description) => {
        await createSpace(name, description)
    }

    return (
        <Layout>
            <div className="space-y-8">
                <div className="flex justify-between items-center">
                    <div>
                        <h1 className="text-3xl font-bold mb-2">Spaces</h1>
                        <p className="text-base-content/70">Organize your videos into spaces</p>
                    </div>
                    <button 
                        className="btn btn-primary"
                        onClick={() => setIsModalOpen(true)}
                    >
                        <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                        Create Space
                    </button>
                </div>

                {spaces.length > 0 ? (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                        {spaces.map((space) => (
                            <SpaceCard key={space.id} space={space} currentUserId={currentUserId} />
                        ))}
                    </div>
                ) : (
                    <div className="text-center py-16">
                        <div className="w-24 h-24 bg-base-300 rounded-full flex items-center justify-center mx-auto mb-4">
                            <svg className="w-12 h-12 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                            </svg>
                        </div>
                        <h3 className="text-xl font-semibold mb-2">No spaces yet</h3>
                        <p className="text-base-content/70 mb-6">Create your first space to organize your videos</p>
                        <button 
                            className="btn btn-primary"
                            onClick={() => setIsModalOpen(true)}
                        >
                            Create Your First Space
                        </button>
                    </div>
                )}
            </div>

            <CreateSpaceModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onCreateSpace={handleCreateSpace}
            />
        </Layout>
    )
} 