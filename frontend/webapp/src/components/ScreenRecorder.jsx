import { useState, useRef } from "react";
import { $authToken } from "../auth/store/auth";
import { useStore } from "@nanostores/react";

export default function ScreenRecorder({ onUploadSuccess, onUploadError }) {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [currentVideoBlob, setCurrentVideoBlob] = useState(null);
  const [uploadFailed, setUploadFailed] = useState(false);
  const mediaRecorder = useRef(null);
  const recordedChunks = useRef([]);
  const authToken = useStore($authToken);

  const startRecording = async () => {
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

      mediaRecorder.current = new MediaRecorder(combinedStream);
      recordedChunks.current = [];

      mediaRecorder.current.ondataavailable = (event) => {
        if (event.data.size > 0) recordedChunks.current.push(event.data);
      };

      mediaRecorder.current.onstop = () => {
        const blob = new Blob(recordedChunks.current, { type: "video/webm" });
        setVideoUrl(URL.createObjectURL(blob));
        setCurrentVideoBlob(blob);
        uploadVideo(blob);
      };

      mediaRecorder.current.start();
      setIsRecording(true);
    } catch (error) {
      console.error("Error starting recording:", error);
    }
  };

  const stopRecording = () => {
    if (mediaRecorder.current) {
      mediaRecorder.current.stop();
      setIsRecording(false);
      mediaRecorder.current.stream.getTracks().forEach((track) => track.stop());
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

  const uploadVideo = async (videoBlob) => {
    if (isUploading) return;
    
    setIsUploading(true);
    setUploadFailed(false);
    const formData = new FormData();
    formData.append("video", videoBlob, "recording.webm");

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
      let message = "Video uploaded successfully!";
      
      if (responseText) {
        const data = JSON.parse(responseText);
        message = data.message || message;
      }

      onUploadSuccess && onUploadSuccess({ message });
      setUploadFailed(false);
    } catch (error) {
      console.error("Error uploading video:", error);
      onUploadError && onUploadError(error);
      setUploadFailed(true);
    } finally {
      setIsUploading(false);
    }
  };

  const handleReupload = () => {
    if (currentVideoBlob) {
      uploadVideo(currentVideoBlob);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-center gap-4">
        {!isRecording ? (
          <button 
            className="btn btn-primary" 
            onClick={startRecording}
            disabled={isUploading}
          >
            Start Recording
          </button>
        ) : (
          <button 
            className="btn btn-error" 
            onClick={stopRecording}
            disabled={isUploading}
          >
            Stop Recording
          </button>
        )}
      </div>

      {videoUrl && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Recording Preview:</h3>
          <video 
            controls 
            src={videoUrl} 
            className="w-full max-w-2xl mx-auto rounded-lg shadow-lg"
          />
          <div className="flex justify-center gap-4">
            <button 
              className="btn btn-secondary"
              onClick={downloadRecording}
              disabled={isUploading}
            >
              Download Video
            </button>
            {uploadFailed && !isUploading && (
              <button 
                className="btn btn-primary"
                onClick={handleReupload}
              >
                Re-upload Video
              </button>
            )}
          </div>
          {isUploading && (
            <div className="flex justify-center items-center gap-2">
              <span className="loading loading-spinner loading-md"></span>
              <span>Uploading video...</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
