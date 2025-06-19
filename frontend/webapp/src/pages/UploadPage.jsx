import React, { useState } from 'react';
import { useStore } from '@nanostores/react';
import { $authToken } from '../auth/store/auth';
import { useNavigate } from 'react-router';
import { Layout } from '../components/layout/Layout';

export const UploadPage = () => {
  const [videoFile, setVideoFile] = useState(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [statusMessage, setStatusMessage] = useState('');
  const [isUploading, setIsUploading] = useState(false);

  const authToken = useStore($authToken);
  const navigate = useNavigate();

  const handleUpload = async () => {
    if (!videoFile || !title || !description) {
      setStatusMessage('Please provide all required fields.');
      return;
    }

    const formData = new FormData();
    formData.append('title', title);
    formData.append('description', description);
    formData.append('video', videoFile);

    setIsUploading(true);
    setStatusMessage('Uploading...');

    try {
      const response = await fetch(import.meta.env.VITE_PUBLIC_API_URL + '/api/videoservice/upload', {
        method: 'POST',
        headers: {
          authorization: authToken,
        },
        body: formData,
      });

      if (!response.ok) throw new Error(`Status: ${response.status}`);

      await response.json();
      setStatusMessage('Upload successful!');
      navigate('/videos');
    } catch (err) {
      console.error(err);
      setStatusMessage('Upload failed. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <Layout>
      <div className="max-w-xl mx-auto space-y-6 p-4">
        <div className="text-center">
          <h2 className="text-3xl font-bold">Upload an Existing Video</h2>
          <p className="mt-2 text-base-content/70">Add a previously recorded video and manage it like other recordings</p>
        </div>

        <div className="space-y-4">
          <input
            type="file"
            accept="video/*"
            onChange={(e) => setVideoFile(e.target.files[0])}
            className="file-input file-input-bordered w-full"
          />
          <input
            type="text"
            placeholder="Title"
            className="input input-bordered w-full"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <textarea
            placeholder="Description"
            className="textarea textarea-bordered w-full"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />

          <button
            className="btn btn-primary w-full"
            onClick={handleUpload}
            disabled={isUploading}
          >
            {isUploading ? 'Uploading...' : 'Upload Video'}
          </button>

          {statusMessage && (
            <div
              className={`alert mt-4 ${
                statusMessage.includes('failed') || statusMessage.includes('required')
                  ? 'alert-error'
                  : 'alert-success'
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
