package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os" // Add this import

	"google.golang.org/protobuf/types/known/timestamppb"
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/proto"
)

type VideoAPI struct {
	config        config.VideoServiceConfig
	HTTPServerMux *http.ServeMux
	dbQueries     *db.Queries
	log           *log.Logger // Added logger field
	proto.UnimplementedVideoServiceServer
}

// NewVideoAPIProduction initializes the VideoAPI with routes and database connections.
func NewVideoAPIProduction(config config.VideoServiceConfig) (*VideoAPI, error) {
	// Initialize the database connection
	dbConn, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	dbQueries := db.New(dbConn)

	// Initialize the API instance
	api := &VideoAPI{
		config:    config,
		dbQueries: dbQueries,
		log:       log.New(os.Stdout, "VIDEO_API: ", log.LstdFlags), // Set up the logger here
	}

	// Create and configure HTTP routes
	api.HTTPServerMux = http.NewServeMux()
	api.HTTPServerMux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		uploadHandler(w, r, api.dbQueries)
	})
	api.HTTPServerMux.HandleFunc("/videos", api.videosHandler)

	return api, nil
}

// Init initializes the service, such as applying database migrations.
func (s *VideoAPI) Init() error {
	// Migrate the database
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}
	return nil
}

// Start is a placeholder for starting the service (if needed).
func (s *VideoAPI) Start() error {
	return nil
}

// ListVideos fetches all videos from the database using the gRPC protocol.
func (s *VideoAPI) ListVideos(ctx context.Context, req *proto.ListVideosRequest) (*proto.ListVideosResponse, error) {
	// Fetch videos from the database
	videos, err := s.dbQueries.GetAllVideo(ctx)
	if err != nil {
		return nil, err
	}

	// Log the fetched videos
	if len(videos) == 0 {
		s.log.Println("No videos found in the database")
	} else {
		s.log.Printf("Fetched %d videos from the database", len(videos))
	}

	// Map the database records to gRPC response objects
	protoVideos := make([]*proto.Video, 0, len(videos))
	for _, video := range videos {
		protoVideos = append(protoVideos, &proto.Video{
			Id:          video.ID,
			Title:       video.Title,
			Description: video.Description,
			Url:         video.Url,
			CreatedAt:   timestamppb.New(video.CreatedAt),
		})
	}

	return &proto.ListVideosResponse{Videos: protoVideos}, nil
}

// videosHandler serves the `/videos` route for HTTP clients.
func (s *VideoAPI) videosHandler(w http.ResponseWriter, r *http.Request) {
	// Call the gRPC ListVideos function
	resp, err := s.ListVideos(context.Background(), &proto.ListVideosRequest{})
	if err != nil {
		http.Error(w, "Failed to fetch videos", http.StatusInternalServerError)
		return
	}

	// Log the response
	if len(resp.Videos) == 0 {
		s.log.Println("No videos available to list")
	} else {
		s.log.Printf("Serving %d videos", len(resp.Videos))
	}

	// Serialize the response to JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Videos)
}
