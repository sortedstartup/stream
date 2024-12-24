import { useState, useEffect } from "react";

export default function VideosList() {
  const [videos, setVideos] = useState([]);
  const [error, setError] = useState(null); // For error messages
  const [loading, setLoading] = useState(true); // For loading state

  useEffect(() => {
    const fetchVideos = async () => {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/api/videoservice/videos`
        );

        if (!response.ok) {
          throw new Error("Failed to fetch videos");
        }

        const contentType = response.headers.get("Content-Type");
        if (contentType?.includes("application/json")) {
          const data = await response.json();
          console.log("Fetched data:", data); // Log the data for debugging
          setVideos(data || []);
        } else {
          throw new Error("Invalid response type, expected JSON");
        }
      } catch (error) {
        console.error("Error fetching videos:", error);
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };

    fetchVideos();
  }, []);

  return (
    <div>
      <h1>Uploaded Videos</h1>
      {loading ? (
        <p>Loading videos...</p>
      ) : error ? (
        <p>Error: {error}</p>
      ) : videos.length > 0 ? (
        <ul>
          {videos.map((video) => (
            <li key={video.id} style={{ marginBottom: "20px" }}>
              <h3>{video.title || "Untitled Video"}</h3>
              <p>{video.description || "No description available."}</p>
              <p>
                <strong>Uploaded At:</strong>{" "}
                {video.created_at
                  ? new Date(video.created_at.seconds * 1000).toLocaleString()
                  : "Invalid date"}
              </p>
              <video controls width="400">
                <source
                  src={`${process.env.NEXT_PUBLIC_API_URL}${video.url}`}
                  type="video/webm"
                />
                Your browser does not support the video tag.
              </video>
            </li>
          ))}
        </ul>
      ) : (
        <p>No videos uploaded yet.</p>
      )}
    </div>
  );
}

