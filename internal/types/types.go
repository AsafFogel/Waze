package types

// format of sending a traffic report
type TrafficReport struct {
	CarID  int     `json:"car_id"`
	EdgeID int     `json:"edge_id"`
	Speed  float64 `json:"speed"`
	// Position  float64
	Timestamp int64 `json:"timestamp"`
}

// format of asking a navigation request
type NavigationRequest struct {
	FromNodeId int `json:"from_node"`
	ToNodeId   int `json:"toNode"`
}

// format of recieving a navigation request answer
type NavigationResponse struct {
	RouteNodes []int   `json:"route"`
	ETA        float64 `json:"eta"`
	Distance   float64 `json:"distance"`
	Err        error   `json:"error"`
}
