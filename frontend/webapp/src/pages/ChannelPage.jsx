import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router';
import { useStore } from '@nanostores/react';
import { Layout } from "../components/layout/Layout";
import { $currentChannel, $channels, getChannelById, fetchChannels } from '../stores/channels';
import { $videos, fetchVideos } from '../stores/videos';
import { $currentTenant } from '../stores/tenants';
import ManageMembersModal from '../components/modals/ManageMembersModal';
import ChannelSettingsModal from '../components/modals/ChannelSettingsModal';

const ChannelPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const channels = useStore($channels);
  const videos = useStore($videos);
  const currentTenant = useStore($currentTenant);
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [userRole, setUserRole] = useState(null);
  const [currentChannel, setCurrentChannel] = useState(null);
  
  // Modal states
  const [showMembersModal, setShowMembersModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);

  useEffect(() => {
    if (id) {
      loadChannelData();
    }
  }, [id, channels]);

  const loadChannelData = async () => {
    try {
      setLoading(true);
      setError('');
      
      // First ensure channels are loaded
      if (channels.length === 0) {
        await fetchChannels();
      }
      
      // Get channel from the loaded channels
      const channelData = getChannelById(id);
      
      if (!channelData) {
        throw new Error('Channel not found');
      }

      setCurrentChannel(channelData);
      // Set user role from channel data (assuming it comes with role info)
      setUserRole(channelData?.user_role || 'viewer');
      
      // Also fetch all videos to show channel videos
      await fetchVideos();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleChannelUpdated = (updatedChannel) => {
    setCurrentChannel(updatedChannel);
    setShowSettingsModal(false);
  };

  const handleUploadVideo = () => {
    // Navigate to upload page with channel context
    navigate('/upload', { state: { channelId: id } });
  };

  const handleRecordVideo = () => {
    // Navigate to record page with channel context
    navigate('/record', { state: { channelId: id } });
  };

  // Filter videos that belong to this channel (if video has channel_id)
  const channelVideos = videos.filter(video => video.channel_id === id);

  const canManage = userRole === 'owner';
  const canUpload = userRole === 'owner' || userRole === 'uploader';
  const isPersonalTenant = currentTenant?.tenant?.is_personal || false;

  if (loading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="loading loading-spinner loading-lg"></div>
        </div>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout>
        <div className="container mx-auto px-4 py-6">
          <div className="alert alert-error">
            <svg className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" 
                    d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>{error}</span>
          </div>
          <button 
            onClick={() => navigate('/channels')}
            className="btn btn-primary mt-4"
          >
            ← Back to Channels
          </button>
        </div>
      </Layout>
    );
  }

  if (!currentChannel) {
    return (
      <Layout>
        <div className="container mx-auto px-4 py-6">
          <div className="text-center py-12">
            <div className="flex justify-center mb-4">
              <svg className="w-16 h-16 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2">Channel not found</h3>
            <button 
              onClick={() => navigate('/channels')}
              className="btn btn-primary"
            >
              ← Back to Channels
            </button>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="container mx-auto px-4 py-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 mb-6 text-sm text-base-content/60">
        <button 
          onClick={() => navigate('/channels')}
          className="flex items-center gap-2 hover:text-base-content transition-colors"
        >
          {isPersonalTenant ? (
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          ) : (
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
          )}
          <span>{currentTenant?.tenant?.name || 'Workspace'}</span>
        </button>
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
        </svg>
        <span className="text-base-content">{currentChannel.name}</span>
      </div>

      {/* Header */}
      <div className="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 mb-8">
        <div className="flex items-center gap-4">
          <button 
            onClick={() => navigate('/channels')}
            className="btn btn-ghost btn-circle"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          
          <div>
            <h1 className="text-3xl font-bold text-base-content flex items-center gap-3">
              <svg className="w-8 h-8 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
              {currentChannel.name}
            </h1>
            <p className="text-base-content/70 mt-1">
              {currentChannel.description || 'No description provided'}
            </p>
            <div className="flex items-center gap-4 mt-2 text-sm text-base-content/60">
              <span className="flex items-center gap-1">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                        d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
                {channelVideos.length} videos
              </span>
              <div className={`badge badge-sm ${
                userRole === 'owner' ? 'badge-primary' : 
                userRole === 'uploader' ? 'badge-secondary' : 
                'badge-accent'
              }`}>
                {userRole === 'owner' ? (
                  <span className="flex items-center gap-1">
                    <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                    </svg>
                    Owner
                  </span>
                ) : userRole === 'uploader' ? (
                  <span className="flex items-center gap-1">
                    <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    Uploader
                  </span>
                ) : (
                  <span className="flex items-center gap-1">
                    <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    Viewer
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>
        
        {/* Action Buttons */}
        <div className="flex items-center gap-2">
          {canUpload && (
            <>
              <button 
                onClick={handleUploadVideo}
                className="btn btn-primary gap-2"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                        d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
                Upload Video
              </button>
              
              <button 
                onClick={handleRecordVideo}
                className="btn btn-secondary gap-2"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                        d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
                Record Video
              </button>
            </>
          )}
          
          {canManage && (
            <div className="dropdown dropdown-end">
              <div tabIndex={0} role="button" className="btn btn-outline gap-2">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                Manage
              </div>
              <ul tabIndex={0} className="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                {/* Only show Manage Members for non-personal tenants */}
                {!isPersonalTenant && (
                  <li>
                    <button onClick={() => setShowMembersModal(true)}>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                              d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5 0a6 6 0 01-4.5 5.197" />
                      </svg>
                      Manage Members
                    </button>
                  </li>
                )}
                <li>
                  <button onClick={() => setShowSettingsModal(true)}>
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Channel Settings
                  </button>
                </li>
              </ul>
            </div>
          )}
        </div>
      </div>

      {/* Videos Grid */}
      {channelVideos.length === 0 ? (
        <div className="text-center py-12">
          <div className="flex justify-center mb-4">
            <svg className="w-16 h-16 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold mb-2">No videos yet</h3>
          <p className="text-base-content/70 mb-6">
            {canUpload 
              ? "Upload your first video to get started" 
              : "No videos have been uploaded to this channel"
            }
          </p>
          {canUpload && (
            <button 
              onClick={handleUploadVideo}
              className="btn btn-primary gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              Upload First Video
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {channelVideos.map((video) => (
            <div key={video.id} className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow duration-300">
              <figure className="aspect-video bg-base-300">
                {video.thumbnailUrl ? (
                  <img 
                    src={video.thumbnailUrl} 
                    alt={video.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="flex items-center justify-center w-full h-full text-base-content/40">
                    <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                            d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                  </div>
                )}
              </figure>
              <div className="card-body">
                <h2 className="card-title text-sm">{video.title}</h2>
                <div className="text-xs text-base-content/60 space-y-1">
                  <p>{video.duration || 'Unknown duration'}</p>
                  <p>{new Date(video.createdAt).toLocaleDateString()}</p>
                </div>
                <div className="card-actions justify-end">
                  <button 
                    onClick={() => navigate(`/video/${video.id}`)}
                    className="btn btn-primary btn-sm"
                  >
                    Watch
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modals */}
      {showMembersModal && (
        <ManageMembersModal
          isOpen={showMembersModal}
          onClose={() => setShowMembersModal(false)}
          channel={currentChannel}
        />
      )}

      {showSettingsModal && (
        <ChannelSettingsModal
          isOpen={showSettingsModal}
          onClose={() => setShowSettingsModal(false)}
          channel={currentChannel}
          onChannelUpdated={handleChannelUpdated}
        />
      )}
    </div>
    </Layout>
  );
};

export default ChannelPage; 