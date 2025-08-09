-- Add file_size column to track video file sizes for storage usage calculation
ALTER TABLE videoservice_videos ADD COLUMN file_size_bytes INTEGER DEFAULT 0; 