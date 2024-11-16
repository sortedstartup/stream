import { useState, useRef } from "react";

export default function ScreenRecorder() {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const mediaRecorder = useRef(null);
  const chunks = useRef([]);

  const startRecording = async () => {
    try {
      // Request screen recording permission
      const stream = await navigator.mediaDevices.getDisplayMedia({
        video: true,
        audio: true, 
      });

      // Initialize MediaRecorder
      mediaRecorder.current = new MediaRecorder(stream);
      chunks.current = [];

      // Collect video data chunks
      mediaRecorder.current.ondataavailable = (event) => {
        if (event.data.size > 0) chunks.current.push(event.data);
      };

      // Save video when recording stops
      mediaRecorder.current.onstop = () => {
        const blob = new Blob(chunks.current, { type: "video/mp4" });
        setVideoUrl(URL.createObjectURL(blob));
      };

      // Start recording
      mediaRecorder.current.start();
      setIsRecording(true);
    } catch (error) {
      console.error("Error accessing display media: ", error);
    }
  };

  const stopRecording = () => {
    if (mediaRecorder.current && mediaRecorder.current.state !== "inactive") {
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

      {videoUrl && (
        <div style={{ marginTop: "20px" }}>
          <h3>Recording Preview:</h3>
          <video
            controls
            src={videoUrl}
            style={{ width: "100%", maxWidth: "600px", marginTop: "10px" }}
          ></video>
          <a
            href={videoUrl}
            download="recording.mp4"
            style={{
              display: "inline-block",
              marginTop: "10px",
              padding: "10px 20px",
              background: "#0070f3",
              color: "#fff",
              textDecoration: "none",
              borderRadius: "5px",
            }}
          >
            Download Video
          </a>
        </div>
      )}
    </div>
  );
}
