package helpers

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
)

type LogPayload struct {
	App        core.App
	SiteID     string
	Method     string
	Path       string
	IP         string
	UserAgent  string
	DurationMs int64
}

var LogQueue chan LogPayload

// InitLogger initializes a pool of workers to process request logs.
func InitLogger(workers int, queueSize int) {
	LogQueue = make(chan LogPayload, queueSize)
	for i := 0; i < workers; i++ {
		go logWorker()
	}
}

func logWorker() {
	for payload := range LogQueue {
		collection, err := payload.App.FindCollectionByNameOrId("request_logs")
		if err != nil {
			log.Println("Error finding request_logs collection:", err)
			continue
		}
		record := core.NewRecord(collection)
		record.Set("site", payload.SiteID)
		record.Set("method", payload.Method)
		record.Set("path", payload.Path)
		record.Set("ip", payload.IP)
		record.Set("user_agent", payload.UserAgent)
		record.Set("duration_ms", payload.DurationMs)
		
		if err := payload.App.Save(record); err != nil {
			log.Println("Error saving request log:", err)
		}
	}
}
