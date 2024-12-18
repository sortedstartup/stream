package api

import (
	"net/http"

	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/proto"
)

type VideoAPI struct {
	//implemented proto server
	proto.UnimplementedVideoServiceServer
	ServerMux *http.ServeMux
}

func NewVideoAPIProduction(config config.VideoServiceConfig) (*VideoAPI, error) {

	ServerMux := http.NewServeMux()
	ServerMux.HandleFunc("/upload", uploadHandler)

	return &VideoAPI{
		ServerMux: ServerMux,
	}, nil

	// return &VideoAPI{}, nil
}

func (s *VideoAPI) Start() error {
	return nil
}

func (s *VideoAPI) Init() error {
	return nil
}
