package server

import "waze/internal/types"

type PathRequest struct {
	StartNodeId int
	EndNodeId   int

	// channel to notify when the response is ready
	ResponseChannel chan PathResult
}

type PathResult struct {
	Response types.NavigationResponse
	Err      error
}
