import { useState, useRef } from "react";
import { $authToken } from "../auth/store/auth";
import { useStore } from "@nanostores/react";

export default function ScreenRecorder() {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const mediaRecorder = useRef(null);
  const recordedChunks = useRef([]);
  const authToken = useStore($authToken)

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

  // Upload the video to the server
  const uploadVideo = async (videoBlob) => {
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
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Handle empty response body gracefully
      const responseText = await response.text();
      console.log("Server response:", responseText);

      if (responseText) {
        const data = JSON.parse(responseText);
        console.log("Upload success:", data);
        alert(data.message || "Video uploaded successfully!");
      } else {
        console.warn("Server returned an empty response.");
        alert("Upload completed, but server returned no data.");
      }
    } catch (error) {
      console.error("Error uploading video:", error);
      alert("Error uploading video!");
    }
  };

  return (
    <div style={{ padding: "20px", textAlign: "center" }}>
      <h1>Screen Recorder</h1>
      {!isRecording ? (
        <button onClick={startRecording}>Start Recording</button>
      ) : (
        <button onClick={stopRecording}>Stop Recording</button>
      )}

      {videoUrl && (
        <div>
          <h3>Recording Preview:</h3>
          <video controls src={videoUrl} style={{ width: "100%", maxWidth: "600px" }}></video>
          <button onClick={downloadRecording}>Download Video</button>
        </div>
      )}
    </div>
  );
}
