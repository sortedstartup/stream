import { useState, useRef } from "react";

export default function ScreenRecorder() {
  const [isRecording, setIsRecording] = useState(false);
  const [videoUrl, setVideoUrl] = useState(null);
  const mediaRecorder = useRef(null);
  const recordedChunks = useRef([]);

  const startRecording = async () => {
    try {
      // Request screen recording
      const screenStream = await navigator.mediaDevices.getDisplayMedia({
        video: true,
        audio: true,
      });

      // Request microphone audio
      const audioStream = await navigator.mediaDevices.getUserMedia({
        audio: true,
      });

      // Combine audio and video tracks
      const combinedStream = new MediaStream();

      screenStream.getTracks().forEach((track) => combinedStream.addTrack(track));
      audioStream.getAudioTracks().forEach((track) =>
        combinedStream.addTrack(track)
      );

      // Initialize MediaRecorder
      mediaRecorder.current = new MediaRecorder(combinedStream);
      recordedChunks.current = [];

      // Collect data chunks
      mediaRecorder.current.ondataavailable = (event) => {
        if (event.data.size > 0) recordedChunks.current.push(event.data);
      };

      // Save the video when recording stops
      mediaRecorder.current.onstop = () => {
        const blob = new Blob(recordedChunks.current, { type: "video/webm" });
        setVideoUrl(URL.createObjectURL(blob));
      };

      // Start recording
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

      // Stop all MediaStream tracks
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

  return (
    <div style={{ padding: "20px", textAlign: "center" }}>
      <h1>Screen Recorder</h1>
      {!isRecording ? (
        <button
          onClick={startRecording}
          style={{
            padding: "10px 20px",
            background: "#0070f3",
            color: "#fff",
            border: "none",
            borderRadius: "5px",
          }}
        >
          Start Recording
        </button>
      ) : (
        <button
          onClick={stopRecording}
          style={{
            padding: "10px 20px",
            background: "#f44336",
            color: "#fff",
            border: "none",
            borderRadius: "5px",
          }}
        >
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
          <button
            onClick={downloadRecording}
            style={{
              marginTop: "10px",
              padding: "10px 20px",
              background: "#0070f3",
              color: "#fff",
              border: "none",
              borderRadius: "5px",
            }}
          >
            Download Video
          </button>
        </div>
      )}
    </div>
  );
}
