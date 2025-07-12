import React, { useState, useEffect } from 'react';
import { useStore } from '@nanostores/react';
import { $currentTenant } from '../../stores/tenants';
import { updateChannel } from '../../stores/channels';

const ChannelSettingsModal = ({ isOpen, onClose, channel, onChannelUpdated }) => {
  const currentTenant = useStore($currentTenant);
  const [formData, setFormData] = useState({
    name: '',
    description: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    if (isOpen && channel) {
      setFormData({
        name: channel.name || '',
        description: channel.description || ''
      });
      setHasChanges(false);
      setError('');
    }
  }, [isOpen, channel]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Check if there are changes
    const newData = { ...formData, [name]: value };
    setHasChanges(
      newData.name !== (channel?.name || '') ||
      newData.description !== (channel?.description || '')
    );
    
    // Clear error when user starts typing
    if (error) setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      setError('Channel name is required');
      return;
    }

    if (!hasChanges) {
      onClose();
      return;
    }

    setLoading(true);
    setError('');

    try {
      const updatedChannel = await updateChannel(
        channel.id,
        formData.name.trim(),
        formData.description.trim()
      );

      if (updatedChannel) {
        onChannelUpdated(updatedChannel);
        onClose();
      } else {
        setError('Failed to update channel');
      }
    } catch (err) {
      setError(err.message || 'Failed to update channel');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      if (hasChanges) {
        if (confirm('You have unsaved changes. Are you sure you want to close?')) {
          onClose();
        }
      } else {
        onClose();
      }
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal modal-open">
      <div className="modal-box">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-bold text-lg flex items-center gap-2">
            <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                    d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Channel Settings
          </h3>
          <button 
            className="btn btn-sm btn-circle btn-ghost"
            onClick={handleClose}
            disabled={loading}
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Channel Name */}
          <div className="form-control">
            <label className="label">
              <span className="label-text font-medium">Channel Name *</span>
            </label>
            <input
              type="text"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="Enter channel name..."
              className={`input input-bordered w-full ${error && !formData.name.trim() ? 'input-error' : ''}`}
              disabled={loading}
              maxLength={100}
            />
            <label className="label">
              <span className="label-text-alt text-base-content/60">
                {formData.name.length}/100 characters
              </span>
            </label>
          </div>

          {/* Channel Description */}
          <div className="form-control">
            <label className="label">
              <span className="label-text font-medium">Description</span>
            </label>
            <textarea
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              placeholder="Describe what this channel is for..."
              className="textarea textarea-bordered w-full"
              disabled={loading}
              rows={3}
              maxLength={500}
            />
            <label className="label">
              <span className="label-text-alt text-base-content/60">
                {formData.description.length}/500 characters
              </span>
            </label>
          </div>

          {/* Error Message */}
          {error && (
            <div className="alert alert-error">
              <svg className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" 
                      d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <span>{error}</span>
            </div>
          )}

          {/* Action Buttons */}
          <div className="modal-action">
            <button
              type="button"
              className="btn btn-ghost"
              onClick={handleClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className={`btn btn-primary ${loading ? 'loading' : ''}`}
              disabled={loading || !formData.name.trim() || !hasChanges}
            >
              {loading ? 'Saving...' : hasChanges ? 'Save Changes' : 'No Changes'}
            </button>
          </div>
        </form>

        {/* Channel Info */}
        <div className="bg-base-200 rounded-lg p-4 mt-4">
          <div className="flex items-start gap-3">
            <div className="text-info">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <div className="text-sm">
              <p className="font-medium mb-1">Channel Statistics</p>
              <ul className="text-base-content/70 space-y-1">
                <li>• Created: {channel?.created_at ? new Date(channel.created_at.seconds * 1000).toLocaleDateString() : 'Unknown'}</li>
                <li>• Videos: {channel?.videoCount || 0}</li>
                <li>• Members: {channel?.memberCount || 0}</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Danger Zone */}
        <div className="bg-error/10 border border-error/20 rounded-lg p-4 mt-4">
          <div className="flex items-start gap-3">
            <div className="text-error">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div className="text-sm">
              <p className="font-medium mb-1 text-error">Danger Zone</p>
              <p className="text-base-content/70 mb-3">
                Permanently delete this channel and all its videos. This action cannot be undone.
              </p>
              <button 
                className="btn btn-error btn-sm"
                disabled={loading}
                onClick={() => alert('Channel deletion will be implemented in a future update.')}
              >
                Delete Channel
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChannelSettingsModal; 