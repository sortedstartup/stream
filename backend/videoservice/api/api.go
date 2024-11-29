package api

import (
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/proto"
)

type VideoAPI struct {
	//implemented proto server
	proto.UnimplementedVideoServiceServer
}

func NewVideoAPIProduction(config config.VideoServiceConfig) (*VideoAPI, error) {
	return &VideoAPI{}, nil
}

func (s *VideoAPI) Start() error {
	return nil
}

func (s *VideoAPI) Init() error {
	return nil
}
