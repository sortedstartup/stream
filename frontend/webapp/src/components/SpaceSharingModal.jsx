import React, { useState, useEffect } from 'react'
import { 
    listUsers, 
    listSpaceMembers, 
    addUserToSpace, 
    removeUserFromSpace, 
    updateUserSpaceAccess 
} from '../stores/videos'
import { AccessLevel } from '../proto/videoservice'

const SpaceSharingModal = ({ isOpen, onClose, spaceId, spaceName, spaceOwnerId }) => {
    const [users, setUsers] = useState([])
    const [members, setMembers] = useState([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')
    const [selectedUser, setSelectedUser] = useState('')
    const [selectedAccessLevel, setSelectedAccessLevel] = useState(AccessLevel.ACCESS_LEVEL_VIEW)

    const loadData = async () => {
        if (!isOpen || !spaceId) return
        
        try {
            setLoading(true)
            const [usersData, membersData] = await Promise.all([
                listUsers(),
                listSpaceMembers(spaceId)
            ])
            setUsers(usersData)
            setMembers(membersData)
        } catch (err) {
            console.error('Error loading sharing data:', err)
            setError('Failed to load users and members')
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        loadData()
    }, [isOpen, spaceId])

    const handleAddUser = async () => {
        if (!selectedUser) return

        try {
            setLoading(true)
            await addUserToSpace(spaceId, selectedUser, selectedAccessLevel)
            setSelectedUser('')
            setSelectedAccessLevel(AccessLevel.ACCESS_LEVEL_VIEW)
            await loadData() // Refresh data
        } catch (err) {
            console.error('Error adding user to space:', err)
            setError('Failed to add user to space')
        } finally {
            setLoading(false)
        }
    }

    const handleRemoveUser = async (userId) => {
        try {
            setLoading(true)
            await removeUserFromSpace(spaceId, userId)
            await loadData() // Refresh data
        } catch (err) {
            console.error('Error removing user from space:', err)
            setError('Failed to remove user from space')
        } finally {
            setLoading(false)
        }
    }

    const handleUpdateAccess = async (userId, newAccessLevel) => {
        try {
            setLoading(true)
            await updateUserSpaceAccess(spaceId, userId, newAccessLevel)
            await loadData() // Refresh data
        } catch (err) {
            console.error('Error updating user access:', err)
            setError('Failed to update user access')
        } finally {
            setLoading(false)
        }
    }

    const getAccessLevelName = (level) => {
        switch (level) {
            case AccessLevel.ACCESS_LEVEL_VIEW:
                return 'View'
            case AccessLevel.ACCESS_LEVEL_EDIT:
                return 'Edit'
            case AccessLevel.ACCESS_LEVEL_ADMIN:
                return 'Admin'
            default:
                return 'View'
        }
    }

    const getAccessLevelColor = (level) => {
        switch (level) {
            case AccessLevel.ACCESS_LEVEL_VIEW:
                return 'badge-info'
            case AccessLevel.ACCESS_LEVEL_EDIT:
                return 'badge-warning'
            case AccessLevel.ACCESS_LEVEL_ADMIN:
                return 'badge-error'
            default:
                return 'badge-info'
        }
    }

    // Filter out users who are already members
    const availableUsers = users.filter(user => 
        !members.some(member => member.user_id === user.id) && 
        user.id !== spaceOwnerId
    )

    if (!isOpen) return null

    return (
        <div className="modal modal-open">
            <div className="modal-box max-w-4xl">
                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                    <h3 className="font-bold text-lg">Share "{spaceName}"</h3>
                    <button
                        className="btn btn-sm btn-circle btn-ghost"
                        onClick={onClose}
                    >
                        ✕
                    </button>
                </div>

                {error && (
                    <div className="alert alert-error mb-4">
                        <svg className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span>{error}</span>
                    </div>
                )}

                {/* Add User Section */}
                <div className="bg-base-200 p-4 rounded-lg mb-6">
                    <h4 className="font-semibold mb-3">Add New Member</h4>
                    <div className="flex gap-3">
                        <select 
                            className="select select-bordered flex-1"
                            value={selectedUser}
                            onChange={(e) => setSelectedUser(e.target.value)}
                        >
                            <option value="">Select a user...</option>
                            {availableUsers.map(user => (
                                <option key={user.id} value={user.id}>
                                    {user.email}
                                </option>
                            ))}
                        </select>
                        
                        <select 
                            className="select select-bordered"
                            value={selectedAccessLevel}
                            onChange={(e) => setSelectedAccessLevel(parseInt(e.target.value))}
                        >
                            <option value={AccessLevel.ACCESS_LEVEL_VIEW}>View</option>
                            <option value={AccessLevel.ACCESS_LEVEL_EDIT}>Edit</option>
                            <option value={AccessLevel.ACCESS_LEVEL_ADMIN}>Admin</option>
                        </select>
                        
                        <button
                            className="btn btn-primary"
                            onClick={handleAddUser}
                            disabled={!selectedUser || loading}
                        >
                            {loading ? (
                                <span className="loading loading-spinner loading-sm"></span>
                            ) : (
                                'Add'
                            )}
                        </button>
                    </div>
                </div>

                {/* Current Members */}
                <div>
                    <h4 className="font-semibold mb-3">Current Members ({members.length})</h4>
                    
                    {loading && members.length === 0 ? (
                        <div className="flex justify-center py-8">
                            <span className="loading loading-spinner loading-lg"></span>
                        </div>
                    ) : members.length === 0 ? (
                        <div className="text-center py-8 text-base-content/60">
                            <p>No members have been added to this space yet.</p>
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {members.map(member => {
                                const user = users.find(u => u.id === member.user_id)
                                return (
                                    <div key={member.user_id} className="flex items-center justify-between p-3 bg-base-100 rounded-lg">
                                        <div className="flex items-center gap-3">
                                            <div className="avatar placeholder">
                                                <div className="bg-neutral-focus text-neutral-content rounded-full w-10">
                                                    <span className="text-sm">
                                                        {user?.email?.charAt(0).toUpperCase() || '?'}
                                                    </span>
                                                </div>
                                            </div>
                                            <div>
                                                <p className="font-medium">{user?.email || member.user_id}</p>
                                                <p className="text-sm text-base-content/60">
                                                    Added {new Date(member.created_at?.seconds * 1000).toLocaleDateString()}
                                                </p>
                                            </div>
                                        </div>
                                        
                                        <div className="flex items-center gap-2">
                                            <select
                                                className="select select-bordered select-sm"
                                                value={member.access_level}
                                                onChange={(e) => handleUpdateAccess(member.user_id, parseInt(e.target.value))}
                                                disabled={loading}
                                            >
                                                <option value={AccessLevel.ACCESS_LEVEL_VIEW}>View</option>
                                                <option value={AccessLevel.ACCESS_LEVEL_EDIT}>Edit</option>
                                                <option value={AccessLevel.ACCESS_LEVEL_ADMIN}>Admin</option>
                                            </select>
                                            
                                            <span className={`badge ${getAccessLevelColor(member.access_level)}`}>
                                                {getAccessLevelName(member.access_level)}
                                            </span>
                                            
                                            <button
                                                className="btn btn-ghost btn-sm text-error hover:bg-error hover:text-error-content"
                                                onClick={() => handleRemoveUser(member.user_id)}
                                                disabled={loading}
                                            >
                                                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                )
                            })}
                        </div>
                    )}
                </div>
            </div>
            
            {/* Background overlay */}
            <div className="modal-backdrop" onClick={onClose}></div>
        </div>
    )
}

export default SpaceSharingModal 