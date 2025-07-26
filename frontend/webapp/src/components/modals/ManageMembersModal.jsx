import React, { useState, useEffect } from 'react';
import { useStore } from '@nanostores/react';
import { $channelMembers, fetchChannelMembers, addChannelMember, removeChannelMember } from '../../stores/channels';
import { $currentTenant, getTenantUsers } from '../../stores/tenants';
import { $currentUser } from '../../auth/store/auth';

const ManageMembersModal = ({ isOpen, onClose, channel }) => {
  const members = useStore($channelMembers);
  const currentTenant = useStore($currentTenant);
  const currentUser = useStore($currentUser);
  
  const [tenantUsers, setTenantUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState('');
  const [showAddMember, setShowAddMember] = useState(false);
  const [selectedUser, setSelectedUser] = useState('');
  const [selectedRole, setSelectedRole] = useState('viewer');

  useEffect(() => {
    if (isOpen && channel) {
      loadChannelData();
    }
  }, [isOpen, channel]);

  const loadChannelData = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Load channel members and tenant users in parallel
      const [_, tenantUsersData] = await Promise.all([
        fetchChannelMembers(channel.id),
        getTenantUsers(currentTenant?.tenant?.id || '')
      ]);
      
      setTenantUsers(tenantUsersData || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleAddMember = async (e) => {
    e.preventDefault();
    
    if (!selectedUser) {
      setError('Please select a user to add');
      return;
    }

    setActionLoading(true);
    setError('');

    try {
      await addChannelMember(channel.id, selectedUser, selectedRole);
      
      // Refresh members list
      await fetchChannelMembers(channel.id);
      
      // Reset form
      setSelectedUser('');
      setSelectedRole('viewer');
      setShowAddMember(false);
    } catch (err) {
      setError(err.message);
    } finally {
      setActionLoading(false);
    }
  };

  const handleRemoveMember = async (userId) => {
    if (!confirm('Are you sure you want to remove this member from the channel?')) {
      return;
    }

    setActionLoading(true);
    setError('');

    try {
      await removeChannelMember(channel.id, userId);
      
      // Refresh members list
      await fetchChannelMembers(channel.id);
    } catch (err) {
      setError(err.message);
    } finally {
      setActionLoading(false);
    }
  };

  const getRoleColor = (role) => {
    switch (role) {
      case 'owner': return 'badge-primary';
      case 'uploader': return 'badge-secondary';
      case 'viewer': return 'badge-accent';
      default: return 'badge-ghost';
    }
  };

  const getRoleIcon = (role) => {
    switch (role) {
      case 'owner': 
        return (
          <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
          </svg>
        );
      case 'uploader': 
        return (
          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
        );
      case 'viewer': 
        return (
          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
        );
      default: 
        return (
          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        );
    }
  };

  // Filter available users (exclude current members)
  const availableUsers = tenantUsers.filter(
    tenantUser => !members.some(member => member.user?.id === tenantUser.user?.id)
  );

  // Helper function to determine if current user can remove a member
  const canRemoveMember = (member) => {
    const memberRole = member.role?.role || member.role;
    const currentUserId = currentUser?.uid;
    
    // Can't remove if no current user
    if (!currentUserId) return false;
    
    // Can't remove yourself (prevent self-removal)
    if (member.user?.id === currentUserId) return false;
    
    // Channel creator can remove anyone (including other owners)
    if (channel?.created_by === currentUserId) return true;
    
    // Non-owners can't remove anyone
    if (memberRole === 'owner') return false;
    
    // Can remove non-owners
    return true;
  };

  if (!isOpen) return null;

  return (
    <div className="modal modal-open">
      <div className="modal-box max-w-2xl">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-bold text-lg flex items-center gap-2">
            <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                    d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5 0a6 6 0 01-4.5 5.197" />
            </svg>
            Manage Members - {channel?.name}
          </h3>
          <button 
            className="btn btn-sm btn-circle btn-ghost"
            onClick={onClose}
            disabled={actionLoading}
          >
            ✕
          </button>
        </div>

        {loading ? (
          <div className="flex justify-center py-8">
            <div className="loading loading-spinner loading-lg"></div>
          </div>
        ) : (
          <>
            {/* Add Member Section */}
            <div className="mb-6">
              <div className="flex justify-between items-center mb-4">
                <h4 className="font-semibold">Channel Members ({members.length})</h4>
                {availableUsers.length > 0 ? (
                  <button
                    onClick={() => setShowAddMember(!showAddMember)}
                    className="btn btn-primary btn-sm gap-2"
                    disabled={actionLoading}
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                    Add Member
                  </button>
                ) : (
                  <div className="text-sm text-base-content/60">
                    No users available to add
                  </div>
                )}
              </div>

              {/* No Available Users Message */}
              {availableUsers.length === 0 && tenantUsers.length === 0 && (
                <div className="alert alert-info mb-4">
                  <svg className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                  </svg>
                  <div>
                    <h3 className="font-bold">No users in this workspace</h3>
                    <div className="text-sm mb-2">To add members to this channel, first add users to your workspace from the Team Management page.</div>
                    <a href="/team" className="btn btn-sm btn-outline">
                      <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                      </svg>
                      Go to Team Management
                    </a>
                  </div>
                </div>
              )}

              {/* All Users Already Added Message */}
              {availableUsers.length === 0 && tenantUsers.length > 0 && (
                <div className="alert alert-warning mb-4">
                  <svg className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16c-.77.833.192 2.5 1.732 2.5z"></path>
                  </svg>
                  <div>
                    <h3 className="font-bold">All workspace users are already channel members</h3>
                    <div className="text-sm">Every user in this workspace has already been added to this channel.</div>
                  </div>
                </div>
              )}

              {/* Add Member Form */}
              {showAddMember && (
                <div className="bg-base-200 rounded-lg p-4 mb-4">
                  <form onSubmit={handleAddMember} className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="form-control">
                        <label className="label">
                          <span className="label-text font-medium">Select User</span>
                        </label>
                        <select
                          value={selectedUser}
                          onChange={(e) => setSelectedUser(e.target.value)}
                          className="select select-bordered w-full"
                          disabled={actionLoading}
                        >
                          <option value="">Choose a user...</option>
                          {availableUsers.map((tenantUser) => (
                            <option key={tenantUser.user?.id || `user-${Math.random()}`} value={tenantUser.user?.id}>
                              {tenantUser.user?.email} ({tenantUser.user?.username})
                            </option>
                          ))}
                        </select>
                      </div>

                      <div className="form-control">
                        <label className="label">
                          <span className="label-text font-medium">Role</span>
                        </label>
                        <select
                          value={selectedRole}
                          onChange={(e) => setSelectedRole(e.target.value)}
                          className="select select-bordered w-full"
                          disabled={actionLoading}
                        >
                          <option value="viewer">Viewer</option>
                          <option value="uploader">Uploader</option>
                          <option value="owner">Owner</option>
                        </select>
                      </div>
                    </div>

                    <div className="flex justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => setShowAddMember(false)}
                        className="btn btn-ghost btn-sm"
                        disabled={actionLoading}
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        className={`btn btn-primary btn-sm ${actionLoading ? 'loading' : ''}`}
                        disabled={actionLoading || !selectedUser}
                      >
                        Add Member
                      </button>
                    </div>
                  </form>
                </div>
              )}
            </div>

            {/* Error Message */}
            {error && (
              <div className="alert alert-error mb-4">
                <svg className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" 
                        d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                <span>{error}</span>
              </div>
            )}

            {/* Members List */}
            <div className="space-y-3">
              {members.length === 0 ? (
                <div className="text-center py-8 text-base-content/60">
                  <div className="flex justify-center mb-2">
                    <svg className="w-12 h-12 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5 0a6 6 0 01-4.5 5.197" />
                    </svg>
                  </div>
                  <p>No members in this channel yet</p>
                </div>
              ) : (
                members.map((member, index) => (
                  <div key={member.user?.id || `member-${index}`} className="flex items-center justify-between p-3 bg-base-200 rounded-lg">
                    <div className="flex items-center gap-3">
                      <div className="avatar placeholder">
                        <div className="bg-neutral-focus text-neutral-content rounded-full w-10">
                          <span className="text-sm">
                            {member.user?.username?.charAt(0)?.toUpperCase() || member.user?.email?.charAt(0)?.toUpperCase() || '?'}
                          </span>
                        </div>
                      </div>
                      <div>
                        <p className="font-medium">{member.user?.username || member.user?.email || 'Unknown User'}</p>
                        <p className="text-sm text-base-content/60">{member.user?.email || 'No email'}</p>
                      </div>
                    </div>

                    <div className="flex items-center gap-3">
                      <div className="flex items-center gap-2">
                      <div className={`badge badge-sm ${getRoleColor(member.role?.role || member.role)}`}>
                        <span className="flex items-center gap-1">
                          {getRoleIcon(member.role?.role || member.role)} {member.role?.role || member.role}
                        </span>
                        </div>
                        {member.user?.id === channel?.created_by && (
                          <div className="badge badge-xs badge-info">
                            <span className="flex items-center gap-1">
                              <svg className="w-2 h-2" fill="currentColor" viewBox="0 0 24 24">
                                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                              </svg>
                              Creator
                            </span>
                          </div>
                        )}
                      </div>
                      
                      {canRemoveMember(member) && (
                        <button
                          onClick={() => handleRemoveMember(member.user?.id)}
                          className="btn btn-ghost btn-xs text-error"
                          disabled={actionLoading}
                        >
                          Remove
                        </button>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          </>
        )}

        {/* Close Button */}
        <div className="modal-action">
          <button
            onClick={onClose}
            className="btn"
            disabled={actionLoading}
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default ManageMembersModal; 