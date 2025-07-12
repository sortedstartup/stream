import React, { useState, useEffect } from 'react';
import { useStore } from '@nanostores/react';
import { $currentTenant, $currentUserRole } from '../stores/tenants';
import { $channels, $isLoadingChannels, $channelError, fetchChannels } from '../stores/channels';
import ChannelCard from './ChannelCard';
import CreateChannelModal from './modals/CreateChannelModal';
import ManageMembersModal from './modals/ManageMembersModal';
import ChannelSettingsModal from './modals/ChannelSettingsModal';

const ChannelDashboard = () => {
  const currentTenant = useStore($currentTenant);
  const currentUserRole = useStore($currentUserRole);
  const channels = useStore($channels);
  const loading = useStore($isLoadingChannels);
  const error = useStore($channelError);
  
  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showMembersModal, setShowMembersModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [selectedChannel, setSelectedChannel] = useState(null);

  // Load channels on component mount
  useEffect(() => {
    if (currentTenant?.tenant?.id) {
      fetchChannels();
    }
  }, [currentTenant?.tenant?.id]);

  const handleCreateChannel = () => {
    setShowCreateModal(true);
  };

  const handleManageMembers = (channel) => {
    setSelectedChannel(channel);
    setShowMembersModal(true);
  };

  const handleSettings = (channel) => {
    setSelectedChannel(channel);
    setShowSettingsModal(true);
  };

  const handleChannelCreated = () => {
    setShowCreateModal(false);
    // fetchChannels is automatically called by the store
  };

  const handleChannelUpdated = () => {
    setShowSettingsModal(false);
    setSelectedChannel(null);
    // Store is automatically updated
  };

  // Determine if user can create channels
  const isPersonalTenant = currentTenant?.tenant?.is_personal || false;
  const canCreateChannels = isPersonalTenant || currentUserRole === 'super_admin';

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="loading loading-spinner loading-lg"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-6">
        <div className="alert alert-error">
          <svg className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" 
                  d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>Error: {error}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-base-content flex items-center gap-3">
            {isPersonalTenant ? (
              <svg className="w-8 h-8 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
            ) : (
              <svg className="w-8 h-8 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            )}
            {currentTenant?.tenant?.name || 'Channels'}
          </h1>
          <p className="text-base-content/70 mt-1">
            {isPersonalTenant ? 'Personal workspace' : 'Team workspace'} • {channels.length} channels
          </p>
        </div>
        
        {/* Create Channel Button - Only for authorized users */}
        {canCreateChannels && (
          <button 
            onClick={handleCreateChannel}
            className="btn btn-primary gap-2"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create Channel
          </button>
        )}
      </div>

      {/* Channels Grid */}
      {channels.length === 0 ? (
        <div className="text-center py-12">
          <div className="flex justify-center mb-4">
            <svg className="w-16 h-16 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold mb-2">No channels yet</h3>
          <p className="text-base-content/70 mb-6">
            {canCreateChannels 
              ? "Create your first channel to organize your videos" 
              : "No channels available in this workspace"
            }
          </p>
          {canCreateChannels && (
            <button 
              onClick={handleCreateChannel}
              className="btn btn-primary gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Create First Channel
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {channels.map((channel) => (
            <ChannelCard
              key={channel.id}
              channel={channel}
              userRole={channel.user_role || 'viewer'} // Use user_role field from backend
              onManageMembers={handleManageMembers}
              onSettings={handleSettings}
            />
          ))}
        </div>
      )}

      {/* Modals */}
      {showCreateModal && (
        <CreateChannelModal
          isOpen={showCreateModal}
          onClose={() => setShowCreateModal(false)}
          onChannelCreated={handleChannelCreated}
        />
      )}

      {showMembersModal && selectedChannel && (
        <ManageMembersModal
          isOpen={showMembersModal}
          onClose={() => {
            setShowMembersModal(false);
            setSelectedChannel(null);
          }}
          channel={selectedChannel}
        />
      )}

      {showSettingsModal && selectedChannel && (
        <ChannelSettingsModal
          isOpen={showSettingsModal}
          onClose={() => {
            setShowSettingsModal(false);
            setSelectedChannel(null);
          }}
          channel={selectedChannel}
          onChannelUpdated={handleChannelUpdated}
        />
      )}
    </div>
  );
};

export default ChannelDashboard; 