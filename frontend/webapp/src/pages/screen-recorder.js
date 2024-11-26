import { useState, useRef } from "react";

export default function ScreenRecorder() {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const mediaRecorder = useRef(null);

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getDisplayMedia({
        video: true,
        audio: true, // Include audio if needed
      });

      mediaRecorder.current = new MediaRecorder(stream);
      
      // Handle chunks as they become available
      mediaRecorder.current.ondataavailable = async (event) => {
        if (event.data.size > 0) {
          const formData = new FormData();
          formData.append("chunk", event.data);
          formData.append("fileName", "recording.webm"); // File name for chunks

          // Send chunk to server
          await fetch("/api/upload-chunk", {
            method: "POST",
            body: formData,
          });
        }
      };

      mediaRecorder.current.start(1000); // Collect chunks every second
      setIsRecording(true);
    } catch (error) {
      console.error("Error starting recording: ", error);
    }
  };

  const stopRecording = () => {
    if (mediaRecorder.current) {
      mediaRecorder.current.stop();
      setIsRecording(false);
    }
  };

  return (
    <div style={{ padding: "20px", textAlign: "center" }}>
      <h1>Screen Recorder</h1>
      {!isRecording ? (
        <button onClick={startRecording} style={{ padding: "10px 20px" }}>
          Start Recording
        </button>
      ) : (
        <button onClick={stopRecording} style={{ padding: "10px 20px" }}>
          Stop Recording
        </button>
      )}
    </div>
  );
}
