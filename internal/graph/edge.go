package graph

import (
	"math"
	"sync/atomic"
	"waze/internal/config"
)

type Edge struct {
	Id         int     `json:"id"`
	From       int     `json:"from"`
	To         int     `json:"to"`
	Length     float64 `json:"length"`     // in KM
	SpeedLimit float64 `json:"speedlimit"` // in KM/hour

	currentSpeed uint64 // in KM per hour
}

/*
type Historical_stats struct{
	// bla
	// bla
	// bla
}
*/

func (e *Edge) GetCurrentSpeed() float64 {
	bits := atomic.LoadUint64(&e.currentSpeed)
	return math.Float64frombits(bits)
}

func (e *Edge) SetCurrentSpeed(speed float64) {
	atomic.StoreUint64(&e.currentSpeed, math.Float64bits(speed))
}

func (e *Edge) UpdateSpeed(measuredSpeed float64) {
	// check for negative time
	if measuredSpeed <= 0 {
		return
	}

	alpha := config.Global.Physics.Alpha

	for {
		oldBits := atomic.LoadUint64(&e.currentSpeed)
		currentSpeed := math.Float64frombits(oldBits)

		if currentSpeed <= 0 {
			currentSpeed = e.SpeedLimit
		}
		// calculate current and measured time
		currentTime := e.Length / currentSpeed
		measuredTime := e.Length / measuredSpeed

		// calculate the new expected time
		updateTime := (alpha * measuredTime) + ((1.0 - alpha) * currentTime)

		// calculate the the speed
		newSpeed := e.Length / updateTime

		// check new speed boundries. not much more then speed limit
		if newSpeed > (e.SpeedLimit * 1.5) {
			newSpeed = e.SpeedLimit * 1.5
		}
		if newSpeed < 1 {
			newSpeed = 1
		}

		newBits := math.Float64bits(newSpeed)

		if atomic.CompareAndSwapUint64(&e.currentSpeed, oldBits, newBits) {
			return
		}
	}
}
