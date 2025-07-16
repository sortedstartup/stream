import { atom, map } from "nanostores";
import { useStore } from "@nanostores/react";
import { $authToken } from "../auth/store/auth";
import { $currentTenant } from "../stores/tenants";

export interface RecordingState {
  isRecording: boolean;
  videoUrl: string | null;
  currentVideoBlob: Blob | null;
  statusMessage: string | null;
}

export const recordingState = map<RecordingState>({
  isRecording: false,
  videoUrl: null,
  currentVideoBlob: null,
  statusMessage: null,
});

// --- Utility: OPFS File Name ---
const fileName = "recording.webm";

// --- Helpers for OPFS ---
const getRecordingFileHandle = async (): Promise<FileSystemFileHandle> => {
  const root = await navigator.storage.getDirectory();
  return await root.getFileHandle(fileName, { create: true });
};

const deleteRecordingFromOPFS = async () => {
  try {
    const root = await navigator.storage.getDirectory();
    await root.removeEntry(fileName);
  } catch (_) {
    // Ignore if not exists
  }
};

export const loadPreviousRecording = async () => {
  try {
    const root = await navigator.storage.getDirectory();
    const handle = await root.getFileHandle(fileName);
    const file = await handle.getFile();
    if (file.size > 0) {
      recordingState.setKey("currentVideoBlob", file);
      recordingState.setKey("videoUrl", URL.createObjectURL(file));
    }
  } catch (_) {
    // No previous recording
  }
};

// --- Media Recorder Setup ---
let mediaRecorder: MediaRecorder | null = null;
let writableStream: FileSystemWritableFileStream | null = null;

export const startRecording = async () => {
  if (recordingState.get().currentVideoBlob) {
    recordingState.setKey(
      "statusMessage",
      "Please upload or download the current recording before starting a new one."
    );
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

    const combinedStream = new MediaStream([
      ...screenStream.getTracks(),
      ...audioStream.getAudioTracks(),
    ]);

    const fileHandle = await getRecordingFileHandle();
    writableStream = await fileHandle.createWritable();

    mediaRecorder = new MediaRecorder(combinedStream);

    mediaRecorder.ondataavailable = async (event) => {
      if (event.data.size > 0 && writableStream) {
        await writableStream.write(event.data);
      }
    };

    mediaRecorder.onstop = async () => {
      if (writableStream) {
        await writableStream.close();
        writableStream = null;
      }

      const file = await fileHandle.getFile();
      recordingState.setKey("currentVideoBlob", file);
      recordingState.setKey("videoUrl", URL.createObjectURL(file));
      recordingState.setKey("statusMessage", "Recording saved. Ready to upload.");
    };

    mediaRecorder.start();
    recordingState.setKey("isRecording", true);
    recordingState.setKey("statusMessage", "Recording started...");
  } catch (err) {
    console.error("Error starting recording:", err);
    recordingState.setKey("statusMessage", "Failed to start recording.");
  }
};

export const stopRecording = () => {
  if (mediaRecorder) {
    mediaRecorder.stop();
    mediaRecorder.stream.getTracks().forEach((track) => track.stop());
    recordingState.setKey("isRecording", false);
    recordingState.setKey("statusMessage", "Recording stopped.");
  }
};

export const handleReupload = (title: string, description: string) => {
  const { currentVideoBlob } = recordingState.get();
  if (currentVideoBlob) {
    uploadRecording(title, description);
  }
};

export const downloadRecording = () => {
  const { videoUrl } = recordingState.get();
  if (!videoUrl) return;
  const a = document.createElement("a");
  a.href = videoUrl;
  a.download = "recording.webm";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
};

export const uploadRecording = async (
  title: string,
  description: string,
  onUploadSuccess?: () => void,
  onUploadError?: (error: Error) => void
) => {
  const { currentVideoBlob } = recordingState.get();
  const authToken = $authToken.get();
  const currentTenant = $currentTenant.get();

  if (!title || !description) {
    recordingState.setKey("statusMessage", "Please enter title and description.");
    return;
  }

  if (!currentVideoBlob) {
    recordingState.setKey("statusMessage", "No video to upload.");
    return;
  }

  if (!currentTenant?.tenant?.id) {
    recordingState.setKey("statusMessage", "No tenant selected. Please select a tenant first.");
    return;
  }

  recordingState.setKey("statusMessage", "Uploading video...");

  const formData = new FormData();
  formData.append("title", title);
  formData.append("description", description);
  formData.append("video", currentVideoBlob, fileName);

  try {
    const response = await fetch(
      import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, "") + "/api/videoservice/upload",
      {
        method: "POST",
        body: formData,
        headers: {
          authorization: authToken,
          "x-tenant-id": currentTenant.tenant.id,
        },
      }
    );

    if (!response.ok) {
      throw new Error(`Upload failed with status: ${response.status}`);
    }

    await deleteRecordingFromOPFS();
    recordingState.set({
      isRecording: false,
      videoUrl: null,
      currentVideoBlob: null,
      statusMessage: "Video uploaded successfully!",
    });

    onUploadSuccess && onUploadSuccess();
  } catch (error) {
    console.error("Upload error:", error);
    recordingState.setKey("statusMessage", "Upload failed. Please try again.");
    onUploadError && onUploadError(error as Error);
  }
};
