import React, { useState, useRef, useEffect } from "react";
import { $authToken } from "../auth/store/auth";
import { useStore } from "@nanostores/react";

export default function ScreenRecorder({ onUploadSuccess, onUploadError }) {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [currentVideoBlob, setCurrentVideoBlob] = useState(null);
  const [uploadFailed, setUploadFailed] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [statusMessage, setStatusMessage] = useState(null);

  const mediaRecorder = useRef(null);
  const writableStreamRef = useRef(null);
  const authToken = useStore($authToken);

  // --- OPFS Helpers ---
  const fileName = "recording.webm";

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

  const loadPreviousRecording = async () => {
    try {
      const root = await navigator.storage.getDirectory();
      const handle = await root.getFileHandle(fileName);
      const file = await handle.getFile();
      if (file.size > 0) {
        setCurrentVideoBlob(file);
        setVideoUrl(URL.createObjectURL(file));
        setShowForm(true);
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
    if (!title || !description) {
      setStatusMessage("Please enter both title and description before uploading.");
      return;
    }

    if (!currentVideoBlob) {
      setStatusMessage("No video to upload.");
      return;
    }

    setIsUploading(true);
    setUploadFailed(false);
    setStatusMessage("Uploading video...");

    const formData = new FormData();
    formData.append("video", currentVideoBlob, fileName);
    formData.append("title", title);
    formData.append("description", description);

    try {
      const response = await fetch(import.meta.env.VITE_PUBLIC_API_URL + "/api/videoservice/upload", {
        method: "POST",
        body: formData,
        headers: {
          "authorization": authToken,
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

  return (
    <div className="space-y-4">
      {statusMessage && (
        <div className="text-sm text-center text-blue-600">{statusMessage}</div>
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
            <div className="space-y-2">
              <label className="block">
                Title:
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Enter video title"
                  className="input input-bordered w-full mt-1"
                />
              </label>
              <label className="block">
                Description:
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Enter video description"
                  className="textarea textarea-bordered w-full mt-1"
                />
              </label>
              <button className="btn btn-success w-full" onClick={uploadVideo} disabled={isUploading}>
                Upload Video
              </button>
            </div>
          )}

          <div className="flex justify-center gap-4 mt-4">
            <button className="btn btn-secondary" onClick={downloadRecording} disabled={isUploading}>
              Download Video
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
