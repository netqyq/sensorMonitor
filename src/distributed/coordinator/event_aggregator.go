package coordinator

import (
  "time"
)

type EventData struct {
	Name      string
	Value     float64
	Timestamp time.Time
}
