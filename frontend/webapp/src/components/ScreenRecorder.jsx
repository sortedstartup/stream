import React, { useState, useRef, useEffect } from "react";
import { $authToken } from "../auth/store/auth";
import { $currentTenant } from "../stores/tenants";
import { $channels, fetchChannels } from "../stores/channels";
import { useStore } from "@nanostores/react";

export default function ScreenRecorder({ onUploadSuccess, onUploadError, contextChannelId }) {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [currentVideoBlob, setCurrentVideoBlob] = useState(null);
  const [uploadFailed, setUploadFailed] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [selectedChannelId, setSelectedChannelId] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [statusMessage, setStatusMessage] = useState(null);
  const [statusType, setStatusType] = useState("info");

  const mediaRecorder = useRef(null);
  const writableStreamRef = useRef(null);
  const authToken = useStore($authToken);
  const currentTenant = useStore($currentTenant);
  const channels = useStore($channels);

  // --- OPFS Helpers ---
  const fileName = "recording.webm";

  useEffect(() => {
    // Fetch channels when component mounts
    if (currentTenant?.tenant?.id) {
      fetchChannels();
    }
  }, [currentTenant]);

  useEffect(() => {
    // Auto-select channel if context is provided
    if (contextChannelId) {
      setSelectedChannelId(contextChannelId);
    }
  }, [contextChannelId]);

  const getRecordingFileHandle = async () => {
    const root = await navigator.storage.getDirectory();
    return await root.getFileHandle(fileName, { create: true });
  };

  const deleteRecordingFromOPFS = async () => {
    try {
      const root = await navigator.storage.getDirectory();
      await root.removeEntry(fileName);
    } catch (_) {
      // file may not exist; that's fine
    }
  };

    const discardRecording = async () => {
    try {
      await deleteRecordingFromOPFS();
      setCurrentVideoBlob(null);
      setVideoUrl(null);
      setShowForm(false);
      setTitle("");
      setDescription("");
      setSelectedChannelId(contextChannelId || "");
      setStatusMessage("Recording discarded.");
    } catch (error) {
      console.error("Error discarding recording:", error);
      setStatusMessage("Failed to discard recording.");
    }
  };

  const loadPreviousRecording = async () => {
    try {
      const root = await navigator.storage.getDirectory();
      const handle = await root.getFileHandle(fileName);
      const file = await handle.getFile();
      if (file.size > 0) {
        setCurrentVideoBlob(file);
        setVideoUrl(URL.createObjectURL(file));
        setShowForm(true);
        // Auto-generate title for existing recording
        if (!title) {
          setTitle(`Recording ${new Date().toLocaleString()}`);
        }
      }
    } catch (_) {
      // No previous recording
    }
  };

  useEffect(() => {
    loadPreviousRecording();
  }, []);

  const startRecording = async () => {
    try {
      if (currentVideoBlob) {
        setStatusMessage("Please upload or download the current recording before starting a new one.");
        return;
      }

      await deleteRecordingFromOPFS();

      const screenStream = await navigator.mediaDevices.getDisplayMedia({
        video: true,
        audio: true,
      });

      const audioStream = await navigator.mediaDevices.getUserMedia({
        audio: true,
      });

      const combinedStream = new MediaStream();
      screenStream.getTracks().forEach((track) => combinedStream.addTrack(track));
      audioStream.getAudioTracks().forEach((track) => combinedStream.addTrack(track));

      const fileHandle = await getRecordingFileHandle();
      const writable = await fileHandle.createWritable();
      writableStreamRef.current = writable;

      mediaRecorder.current = new MediaRecorder(combinedStream);

      mediaRecorder.current.ondataavailable = async (event) => {
        if (event.data.size > 0 && writableStreamRef.current) {
          await writableStreamRef.current.write(event.data);
        }
      };

      mediaRecorder.current.onstop = async () => {
        if (writableStreamRef.current) {
          await writableStreamRef.current.close();
          writableStreamRef.current = null;
        }

        const file = await fileHandle.getFile();
        setCurrentVideoBlob(file);
        setVideoUrl(URL.createObjectURL(file));
        setShowForm(true);
        setStatusMessage(null);
        
        // Auto-generate title with timestamp
        if (!title) {
          setTitle(`Recording ${new Date().toLocaleString()}`);
        }
      };

      mediaRecorder.current.start();
      setIsRecording(true);
      setStatusMessage("Recording started...");
    } catch (error) {
      console.error("Error starting recording:", error);
      setStatusMessage("Failed to start recording.");
    }
  };

  const stopRecording = () => {
    if (mediaRecorder.current) {
      mediaRecorder.current.stop();
      setIsRecording(false);
      mediaRecorder.current.stream.getTracks().forEach((track) => track.stop());
      setStatusMessage("Recording stopped.");
    }
  };

  const downloadRecording = () => {
    if (!videoUrl) return;
    const a = document.createElement("a");
    a.href = videoUrl;
    a.download = fileName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  };

  const uploadVideo = async () => {
    if (!currentVideoBlob) {
      setStatusMessage("No video to upload.");
      return;
    }

    if (!currentTenant?.tenant?.id) {
      setStatusMessage("No tenant selected. Please select a tenant first.");
      return;
    }

    setIsUploading(true);
    setUploadFailed(false);
    setStatusMessage("Uploading video...");

    const formData = new FormData();
    
    // Use auto-generated title if user hasn't provided one
    const finalTitle = title.trim() || `Recording ${new Date().toLocaleString()}`;
    
    formData.append("title", finalTitle);
    formData.append("description", description.trim()); // Optional, can be empty
    formData.append("channel_id", selectedChannelId); // Optional, can be empty for personal videos
    formData.append("video", currentVideoBlob, fileName);
    
    try {
      const response = await fetch(import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, "") + "/api/videoservice/upload", {
        method: "POST",
        body: formData,
        headers: {
          "authorization": authToken,
          "x-tenant-id": currentTenant.tenant.id,
        },
      });

      if (!response.ok) {
        throw new Error(`Upload failed with status: ${response.status}`);
      }

      const responseText = await response.text();
      const data = JSON.parse(responseText);
      const message = data.message || "Video uploaded successfully!";
      setStatusMessage(message);

      onUploadSuccess && onUploadSuccess({ message });

      await deleteRecordingFromOPFS();
      setShowForm(false);
      setVideoUrl(null);
      setCurrentVideoBlob(null);
      setTitle("");
      setDescription("");
      setSelectedChannelId(contextChannelId || "");
    } catch (error) {
      console.error("Error uploading video:", error);
      setUploadFailed(true);
      setStatusMessage("Upload failed. Please try again.");
      onUploadError && onUploadError(error);
    } finally {
      setIsUploading(false);
    }
  };

  const handleReupload = () => {
    if (currentVideoBlob) {
      uploadVideo();
    }
  };

  // Determine if we should show channel selector
  const shouldShowChannelSelector = !contextChannelId && channels.length > 0;

  return (
    <div className="space-y-4">
      {statusMessage === "Please upload or download the current recording before starting a new one." && (
        <div className="alert alert-error shadow-lg">
          <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M18.364 5.636L5.636 18.364M5.636 5.636l12.728 12.728" />
          </svg>
          <span>{statusMessage}</span>
        </div>
      )}

      <div className="flex justify-center gap-4">
        {!isRecording ? (
          <button className="btn btn-primary" onClick={startRecording} disabled={isUploading}>
            Start Recording
          </button>
        ) : (
          <button className="btn btn-error" onClick={stopRecording} disabled={isUploading}>
            Stop Recording
          </button>
        )}
      </div>

      {videoUrl && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Recording Preview:</h3>
          <video controls src={videoUrl} className="w-full max-w-2xl mx-auto rounded-lg shadow-lg" />

          {showForm && (
            <div className="space-y-4">
              {/* Title Input */}
              <div>
                <label className="block text-sm font-medium mb-2">Title</label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Auto-generated from timestamp"
                  className="input input-bordered w-full"
                />
                <div className="text-xs text-base-content/60 mt-1">
                  Leave empty to use timestamp as title
                </div>
              </div>

              {/* Description Input */}
              <div>
                <label className="block text-sm font-medium mb-2">Description (Optional)</label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Add a description for your recording"
                  className="textarea textarea-bordered w-full"
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

              {/* Context Channel Info */}
              {contextChannelId && (
                <div className="bg-base-200 p-3 rounded-lg">
                  <div className="text-sm font-medium text-base-content/70">
                    Recording will be saved to: {channels.find(c => c.id === contextChannelId)?.name || 'Selected Channel'}
                  </div>
                </div>
              )}

              <button className="btn btn-success w-full" onClick={uploadVideo} disabled={isUploading}>
                {isUploading ? 'Uploading...' : 'Upload Recording'}
              </button>
            </div>
          )}

          <div className="flex justify-center gap-4 mt-4">
            <button className="btn btn-secondary" onClick={downloadRecording} disabled={isUploading}>
              Download Video
            </button>

            <button className="btn btn-warning" onClick={discardRecording} disabled={isUploading}>
              Discard Recording
            </button>

            {uploadFailed && !isUploading && (
              <button className="btn btn-primary" onClick={handleReupload}>
                Re-upload Video
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
