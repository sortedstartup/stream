import React, { useState } from 'react';
import { createChannel } from '../../stores/channels';

const CreateChannelModal = ({ isOpen, onClose, onChannelCreated }) => {
  const [formData, setFormData] = useState({
    name: '',
    description: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    // Clear error when user starts typing
    if (error) setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      setError('Channel name is required');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const channel = await createChannel(
        formData.name.trim(),
        formData.description.trim()
      );

      if (channel) {
        onChannelCreated(channel);
        // Reset form
        setFormData({ name: '', description: '' });
      } else {
        setError('Failed to create channel');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setFormData({ name: '', description: '' });
      setError('');
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal modal-open">
      <div className="modal-box">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-bold text-lg flex items-center gap-2">
            <span className="text-2xl">📁</span>
            Create New Channel
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
              placeholder="e.g., Marketing Team, Product Updates..."
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
              <span className="label-text font-medium">Description (Optional)</span>
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
              disabled={loading || !formData.name.trim()}
            >
              {loading ? 'Creating...' : 'Create Channel'}
            </button>
          </div>
        </form>

        {/* Info Box */}
        <div className="bg-base-200 rounded-lg p-4 mt-4">
          <div className="flex items-start gap-3">
            <div className="text-info text-xl">💡</div>
            <div className="text-sm">
              <p className="font-medium mb-1">Channel Information</p>
              <ul className="text-base-content/70 space-y-1">
                <li>• You'll be the channel owner and can manage members</li>
                <li>• Channel members can view and upload videos (based on their role)</li>
                <li>• You can change these settings later</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CreateChannelModal; 