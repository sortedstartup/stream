import React, { useState, useEffect } from 'react';
import { useStore } from '@nanostores/react';
import { useLocation, useNavigate } from 'react-router';
import { $authToken } from '../auth/store/auth';
import { $currentTenant } from '../stores/tenants';
import { $channels, fetchChannels } from '../stores/channels';
import { Layout } from '../components/layout/Layout';

export const UploadPage = () => {
  const [videoFile, setVideoFile] = useState(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [selectedChannelId, setSelectedChannelId] = useState('');
  const [statusMessage, setStatusMessage] = useState('');
  const [isUploading, setIsUploading] = useState(false);

  const authToken = useStore($authToken);
  const currentTenant = useStore($currentTenant);
  const channels = useStore($channels);
  const navigate = useNavigate();
  const location = useLocation();

  // Get channel context from navigation state (if coming from channel page)
  const contextChannelId = location.state?.channelId;

  useEffect(() => {
    // Fetch channels when component mounts
    if (currentTenant?.tenant?.id) {
      fetchChannels();
    }
  }, [currentTenant]);

  useEffect(() => {
    // Auto-select channel if coming from channel page
    if (contextChannelId) {
      setSelectedChannelId(contextChannelId);
    }
  }, [contextChannelId]);

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    setVideoFile(file);
    
    // Auto-generate title from filename if title is empty
    if (file && !title) {
      const filename = file.name;
      const nameWithoutExtension = filename.substring(0, filename.lastIndexOf('.')) || filename;
      setTitle(nameWithoutExtension);
    }
  };

  const handleUpload = async () => {
    if (!videoFile) {
      setStatusMessage('Please select a video file.');
      return;
    }

    if (!currentTenant?.tenant?.id) {
      setStatusMessage('No tenant selected. Please select a tenant first.');
      return;
    }

    const formData = new FormData();
    
    // Use auto-generated title if user hasn't provided one
    const finalTitle = title.trim() || videoFile.name.substring(0, videoFile.name.lastIndexOf('.')) || `Recording ${new Date().toLocaleString()}`;
    
    formData.append('title', finalTitle);
    formData.append('description', description.trim()); // Optional, can be empty
    formData.append('channel_id', selectedChannelId); // Optional, can be empty for personal videos
    formData.append('video', videoFile);

    setIsUploading(true);
    setStatusMessage('Uploading...');

    try {
      const response = await fetch(import.meta.env.VITE_PUBLIC_API_URL + '/api/videoservice/upload', {
        method: 'POST',
        headers: {
          authorization: authToken,
          'x-tenant-id': currentTenant.tenant.id,
        },
        body: formData,
      });

      if (!response.ok) throw new Error(`Status: ${response.status}`);

      await response.json();
      setStatusMessage('Upload successful!');
      
      // Navigate back to appropriate page
      if (selectedChannelId) {
        navigate(`/channel/${selectedChannelId}`);
      } else {
        navigate('/channels');
      }
    } catch (err) {
      console.error(err);
      setStatusMessage('Upload failed. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  // Determine if we should show channel selector
  const shouldShowChannelSelector = !contextChannelId && channels.length > 0;

  return (
    <Layout>
      <div className="max-w-xl mx-auto space-y-6 p-4">
        <div className="text-center">
          <h2 className="text-3xl font-bold">Upload Video</h2>
          <p className="mt-2 text-base-content/70">
            {contextChannelId 
              ? `Uploading to channel: ${channels.find(c => c.id === contextChannelId)?.name || 'Selected Channel'}`
              : 'Upload a video to your workspace'
            }
          </p>
        </div>

        <div className="space-y-4">
          {/* File Input */}
          <div>
            <label className="block text-sm font-medium mb-2">Video File *</label>
            <input
              type="file"
              accept="video/*"
              onChange={handleFileChange}
              className="file-input file-input-bordered w-full"
            />
          </div>

          {/* Title Input */}
          <div>
            <label className="block text-sm font-medium mb-2">Title</label>
            <input
              type="text"
              placeholder="Auto-generated from filename"
              className="input input-bordered w-full"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
            <div className="text-xs text-base-content/60 mt-1">
              Leave empty to use filename as title
            </div>
          </div>

          {/* Description Input */}
          <div>
            <label className="block text-sm font-medium mb-2">Description (Optional)</label>
            <textarea
              placeholder="Add a description for your video"
              className="textarea textarea-bordered w-full"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          {/* Channel Selector - Only show if needed */}
          {shouldShowChannelSelector && (
            <div>
              <label className="block text-sm font-medium mb-2">Channel</label>
              <select
                className="select select-bordered w-full"
                value={selectedChannelId}
                onChange={(e) => setSelectedChannelId(e.target.value)}
              >
                <option value="">My Videos (Personal)</option>
                {channels.map((channel) => (
                  <option key={channel.id} value={channel.id}>
                    {channel.name}
                  </option>
                ))}
              </select>
              <div className="text-xs text-base-content/60 mt-1">
                Select a channel to share with team members
              </div>
            </div>
          )}

          {/* Upload Button */}
          <button
            className="btn btn-primary w-full"
            onClick={handleUpload}
            disabled={isUploading || !videoFile}
          >
            {isUploading ? 'Uploading...' : 'Upload Video'}
          </button>

          {/* Status Message */}
          {statusMessage && (
            <div
              className={`alert mt-4 ${
                statusMessage.includes('failed') || statusMessage.includes('select')
                  ? 'alert-error'
                  : statusMessage.includes('successful')
                  ? 'alert-success'
                  : 'alert-info'
              }`}
            >
              <span>{statusMessage}</span>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
};
