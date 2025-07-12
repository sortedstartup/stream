import React, { useState } from "react";
import { useRecordingController } from "../context/RecordingContext";

export default function ScreenRecorder({ onUploadSuccess, onUploadError }) {
  const {
    isRecording,
    videoUrl,
    statusMessage,
    startRecording,
    stopRecording,
    uploadRecording,
    downloadRecording,
    handleReupload
  } = useRecordingController();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [uploading, setUploading] = useState(false);
  const [uploadFailed, setUploadFailed] = useState(false);
  const [localStatus, setLocalStatus] = useState(null);

  const handleUpload = async () => {
    setUploading(true);
    setUploadFailed(false);
    setLocalStatus(null);

    try {
      await uploadRecording(
        title,
        description,
        () => {
          setLocalStatus("Upload successful!");
          onUploadSuccess && onUploadSuccess();
          setTitle("");
          setDescription("");
        },
        (error) => {
          setLocalStatus("Upload failed. Please try again.");
          setUploadFailed(true);
          onUploadError && onUploadError(error);
        }
      );
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="space-y-4">
     {statusMessage && statusMessage !== "Video uploaded successfully!" && (
        <div
          className={`alert ${
            statusMessage.toLowerCase().includes("failed")
              ? "alert-error"
              : "alert-info"
          } shadow-lg`}
        >
          <span>{statusMessage}</span>
        </div>
      )}
      <div className="flex justify-center gap-4">
        {!isRecording ? (
          <button
            className="btn btn-primary"
            onClick={startRecording}
            disabled={uploading}
          >
            Start Recording
          </button>
        ) : (
          <button
            className="btn btn-error"
            onClick={stopRecording}
            disabled={uploading}
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

          <div className="space-y-2">
            <label className="block">
              Title:
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Enter video title"
                className="input input-bordered w-full mt-1"
                disabled={uploading}
              />
            </label>
            <label className="block">
              Description:
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Enter video description"
                className="textarea textarea-bordered w-full mt-1"
                disabled={uploading}
              />
            </label>

            <div className="flex flex-wrap justify-center gap-4 mt-4">
              <button
                className="btn btn-success"
                onClick={handleUpload}
                disabled={uploading}
              >
                {uploading ? "Uploading..." : "Upload Video"}
              </button>

              <button
                className="btn btn-secondary"
                onClick={downloadRecording}
                disabled={uploading}
              >
                Download Video
              </button>

              {uploadFailed && !uploading && (
                <button className="btn btn-primary" onClick={handleReupload}>
                  Re-upload Video
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
