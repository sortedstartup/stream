import React, { useState } from 'react';
import { useStore } from '@nanostores/react';
import { $channels } from '../stores/channels';
import { moveVideoToChannel, removeVideoFromChannel, deleteVideo, updateVideo } from '../stores/videos';

const VideoActionsMenu = ({ video, userRole, onActionStart, onActionComplete, onActionError }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [showMoveModal, setShowMoveModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [selectedChannelId, setSelectedChannelId] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editTitle, setEditTitle] = useState(video.title);
  const [editDescription, setEditDescription] = useState(video.description);
  const [editIsPrivate, setEditIsPrivate] = useState(video.visibility === 0); // private = 0

  const channels = useStore($channels);

  // Determine what actions are available based on video location and user permissions
  const isInChannel = video.channel_id && video.channel_id !== '';
  const canMove = !isInChannel; // Can only move tenant-level videos to channels
  const canRemove = isInChannel && (userRole === 'owner'); // Only channel owners can remove from channel
  const canDelete = isInChannel ? (userRole === 'owner') : true; // Channel owners for channel videos, anyone for their own tenant videos

  const handleMoveToChannel = async () => {
    if (!selectedChannelId || selectedChannelId === video.channel_id) return;
    
    setIsLoading(true);
    onActionStart?.('Moving video to channel...');
    
    try {
      await moveVideoToChannel(video.id, selectedChannelId);
      onActionComplete?.('Video moved to channel successfully');
      setShowMoveModal(false);
      setSelectedChannelId('');
      setIsOpen(false);
    } catch (error) {
      onActionError?.('Failed to move video to channel');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRemoveFromChannel = async () => {
    setIsLoading(true);
    onActionStart?.('Removing video from channel...');
    
    try {
      await removeVideoFromChannel(video.id);
      onActionComplete?.('Video removed from channel successfully');
      setIsOpen(false);
    } catch (error) {
      onActionError?.('Failed to remove video from channel');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteVideo = async () => {
    setIsLoading(true);
    onActionStart?.('Deleting video...');
    
    try {
      await deleteVideo(video.id);
      onActionComplete?.('Video deleted successfully');
      setShowDeleteConfirm(false);
      setIsOpen(false);
    } catch (error) {
      onActionError?.('Failed to delete video');
    } finally {
      setIsLoading(false);
    }
  };

  // Filter available channels for moving (exclude current channel if video is in one)
  const availableChannels = channels.filter(channel => 
    channel.id !== video.channel_id && 
    (channel.user_role === 'owner' || channel.user_role === 'uploader')
  );

  const hasAnyActions = canMove || canRemove || canDelete;

  if (!hasAnyActions) {
    return null; // Don't show menu if no actions available
  }

  return (
    <div className="relative">
      {/* Actions Button - More visible three dots */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="btn btn-ghost btn-sm hover:bg-base-200 p-1 min-h-0 h-6 w-6"
        disabled={isLoading}
        title="Video actions"
      >
        <svg className="w-4 h-4 text-base-content" fill="currentColor" viewBox="0 0 20 20">
          <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
        </svg>
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div 
            className="fixed inset-0 z-10" 
            onClick={() => setIsOpen(false)}
          />
          
          {/* Menu */}
          <div className="absolute right-0 top-full mt-1 w-48 bg-base-100 rounded-lg shadow-lg border border-base-300 z-20">
            <div className="py-1">
              {canMove && availableChannels.length > 0 && (
                <button
                  onClick={() => setShowMoveModal(true)}
                  className="w-full px-4 py-2 text-left hover:bg-base-200 flex items-center gap-2 cursor-pointer"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                  </svg>
                  Move to Channel
                </button>
              )}
              
              {canRemove && (
                <button
                  onClick={handleRemoveFromChannel}
                  className="w-full px-4 py-2 text-left hover:bg-base-200 flex items-center gap-2 cursor-pointer"
                  disabled={isLoading}
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                  </svg>
                  Remove from Channel
                </button>
              )}

              <button
                onClick={() => {
                  setEditTitle(video.title);
                  setEditDescription(video.description);
                  setEditIsPrivate(video.visibility === 0);
                  setShowEditModal(true);
                }}
                className="w-full px-4 py-2 text-left hover:bg-base-200 flex items-center gap-2 cursor-pointer"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v12a2 2 0 002 2h12a2 2 0 002-2v-5m-5-5l5 5M14 3l7 7" />
                </svg>
                Edit Details
              </button>

              
              {canDelete && (
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="w-full px-4 py-2 text-left hover:bg-error hover:text-error-content flex items-center gap-2 text-error cursor-pointer"
                  disabled={isLoading}
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  Delete Video
                </button>
              )}
            </div>
          </div>
        </>
      )}

      {/* Move to Channel Modal - Fixed background and improved UX */}
      {showMoveModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-base-100 rounded-lg p-6 w-full max-w-md mx-4 max-h-[80vh] overflow-y-auto">
            <h3 className="text-lg font-semibold mb-4">Move Video to Channel</h3>
            <p className="text-sm text-base-content/70 mb-4">
              <strong>Warning:</strong> Moving this video to a channel will transfer ownership to the channel owner.
            </p>
            
            {/* Channel Selection */}
            <div className="mb-6">
              <label className="block text-sm font-medium mb-3">Select a channel:</label>
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {availableChannels.map((channel) => (
                  <label
                    key={channel.id}
                    className={`flex items-center p-3 rounded-lg border cursor-pointer transition-colors ${
                      selectedChannelId === channel.id
                        ? 'border-primary bg-primary/10'
                        : 'border-base-300 hover:border-base-400 hover:bg-base-200'
                    }`}
                  >
                    <input
                      type="radio"
                      name="channel"
                      value={channel.id}
                      checked={selectedChannelId === channel.id}
                      onChange={(e) => setSelectedChannelId(e.target.value)}
                      className="radio radio-primary radio-sm mr-3"
                    />
                    <div className="flex-1">
                      <div className="font-medium">{channel.name}</div>
                      <div className="text-sm text-base-content/70">
                        {channel.description || 'No description'}
                      </div>
                    </div>
                  </label>
                ))}
              </div>
            </div>
            
            <div className="flex gap-2">
              <button
                onClick={() => {
                  setShowMoveModal(false);
                  setSelectedChannelId('');
                }}
                className="btn btn-ghost flex-1"
                disabled={isLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleMoveToChannel}
                className="btn btn-primary flex-1"
                disabled={isLoading || !selectedChannelId}
              >
                {isLoading ? (
                  <>
                    <span className="loading loading-spinner loading-sm"></span>
                    Moving...
                  </>
                ) : (
                  'Move to Channel'
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal - Fixed background */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-base-100 rounded-lg p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold mb-4 text-error">Delete Video</h3>
            <p className="text-sm text-base-content/70 mb-6">
              Are you sure you want to delete "<strong>{video.title}</strong>"? This action cannot be undone.
            </p>
            
            <div className="flex gap-2">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="btn btn-ghost flex-1"
                disabled={isLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteVideo}
                className="btn btn-error flex-1"
                disabled={isLoading}
              >
                {isLoading ? (
                  <>
                    <span className="loading loading-spinner loading-sm"></span>
                    Deleting...
                  </>
                ) : (
                  'Delete'
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {showEditModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-base-100 rounded-lg p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold mb-4">Edit Video Details</h3>

            <div className="mb-4">
              <label className="block text-sm font-medium mb-1">Title</label>
              <input
                type="text"
                className="input input-bordered w-full"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                className="textarea textarea-bordered w-full"
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                rows={3}
              />
            </div>

            <div className="form-control mb-6">
              <label className="cursor-pointer label justify-start gap-4">
                <input
                  type="checkbox"
                  className="checkbox"
                  checked={editIsPrivate}
                  onChange={() => setEditIsPrivate(!editIsPrivate)}
                />
                <span className="label-text">Private Video</span>
              </label>
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => setShowEditModal(false)}
                className="btn btn-ghost flex-1"
                disabled={isLoading}
              >
                Cancel
              </button>
              <button
                onClick={async () => {
                  setIsLoading(true);
                  onActionStart?.("Updating video details...");

                  try {
                    await updateVideo(video.id, editTitle, editDescription, editIsPrivate);
                    onActionComplete?.("Video updated successfully.");
                    setShowEditModal(false);
                    setIsOpen(false);
                  } catch (error) {
                    onActionError?.("Failed to update video.");
                  } finally {
                    setIsLoading(false);
                  }
                }}
                className="btn btn-primary flex-1"
                disabled={isLoading}
              >
                {isLoading ? (
                  <>
                    <span className="loading loading-spinner loading-sm"></span>
                    Saving...
                  </>
                ) : (
                  "Save Changes"
                )}
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
};

export default VideoActionsMenu; 