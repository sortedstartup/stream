import fs from "fs";
import path from "path";

export const config = {
  api: {
    bodyParser: false, // Disable default body parser for file uploads
  },
};

export default function handler(req, res) {
  if (req.method === "POST") {
    const videoFolderPath = path.join(process.cwd(), "uploads");

    // Ensure the uploads directory exists
    if (!fs.existsSync(videoFolderPath)) {
      fs.mkdirSync(videoFolderPath);
    }

    const filePath = path.join(videoFolderPath, "recording.webm");

    // Handle incoming chunks
    const chunks = [];
    req.on("data", (chunk) => {
      chunks.push(chunk);
    });

    req.on("end", () => {
      try {
        // Append the chunk to the file
        fs.appendFileSync(filePath, Buffer.concat(chunks));
        res.status(200).json({ message: "Chunk uploaded successfully" });
      } catch (error) {
        console.error("Error saving chunk: ", error);
        res.status(500).json({ message: "Failed to upload chunk" });
      }
    });
  } else {
    res.setHeader("Allow", ["POST"]);
    res.status(405).end(`Method ${req.method} Not Allowed`);
  }
}
