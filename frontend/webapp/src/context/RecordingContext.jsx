import React, { createContext, useContext, useRef, useState, useEffect } from "react";
import { useStore } from "@nanostores/react";
import { $authToken } from "../auth/store/auth";
import { $currentTenant } from "../stores/tenants";

const RecordingContext = createContext();

export function useRecordingController() {
  return useContext(RecordingContext);
}

export function RecordingProvider({ children }) {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const [currentVideoBlob, setCurrentVideoBlob] = useState(null);
  const [statusMessage, setStatusMessage] = useState(null);

  const mediaRecorder = useRef(null);
  const writableStreamRef = useRef(null);

  const authToken = useStore($authToken);
  const currentTenant = useStore($currentTenant);

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
      // File may not exist
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
      }
    } catch (_) {
      // No previous recording
    }
  };

  useEffect(() => {
    loadPreviousRecording();
  }, []);

  const startRecording = async () => {
    if (currentVideoBlob) {
      setStatusMessage("Please upload or download the current recording before starting a new one.");
      return;
    }

    await deleteRecordingFromOPFS();

    try {
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
        setStatusMessage("Recording saved. Ready to upload.");
      };

      mediaRecorder.current.start();
      setIsRecording(true);
      setStatusMessage("Recording started...");
    } catch (err) {
      console.error("Error starting recording:", err);
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
        a.download = "recording.webm";
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
    };

    const handleReupload = () => {
        if (currentVideoBlob) {
            uploadRecording(title, description);
        }
    };

  const uploadRecording = async (title, description, onUploadSuccess, onUploadError) => {
    if (!title || !description) {
      setStatusMessage("Please enter title and description.");
      return;
    }

    if (!currentVideoBlob) {
      setStatusMessage("No video to upload.");
      return;
    }

    if (!currentTenant?.tenant?.id) {
      setStatusMessage("No tenant selected. Please select a tenant first.");
      return;
    }

    setStatusMessage("Uploading video...");

    const formData = new FormData();
    formData.append("title", title);
    formData.append("description", description);
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

      await deleteRecordingFromOPFS();
      setVideoUrl(null);
      setCurrentVideoBlob(null);
      setStatusMessage("Video uploaded successfully!");

      onUploadSuccess && onUploadSuccess();
    } catch (error) {
      console.error("Upload error:", error);
      setStatusMessage("Upload failed. Please try again.");
      onUploadError && onUploadError(error);
    }
  };

  return (
    <RecordingContext.Provider
      value={{
        isRecording,
        videoUrl,
        downloadRecording,
        handleReupload,
        currentVideoBlob,
        statusMessage,
        startRecording,
        stopRecording,
        uploadRecording,
      }}
    >
      {children}
    </RecordingContext.Provider>
  );
}
