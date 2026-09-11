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

const logBatchSize = 100

// InitLogger initializes a pool of workers to process request logs.
func InitLogger(workers int, queueSize int) {
	LogQueue = make(chan LogPayload, queueSize)
	for i := 0; i < workers; i++ {
		go logWorker()
	}
}

func logWorker() {
	var collection *core.Collection
	for first := range LogQueue {
		batch := make([]LogPayload, 1, logBatchSize)
		batch[0] = first

		for len(batch) < logBatchSize {
			select {
			case payload := <-LogQueue:
				batch = append(batch, payload)
			default:
				collection = persistLogBatch(batch, collection)
				batch = nil
			}
			if batch == nil {
				break
			}
		}
		if batch != nil {
			collection = persistLogBatch(batch, collection)
		}
	}
}

func persistLogBatch(batch []LogPayload, collection *core.Collection) *core.Collection {
	app := batch[0].App
	if collection == nil {
		var err error
		collection, err = app.FindCollectionByNameOrId("request_logs")
		if err != nil {
			log.Println("Error finding request_logs collection:", err)
			return nil
		}
	}

	err := app.RunInTransaction(func(txApp core.App) error {
		for _, payload := range batch {
			record := core.NewRecord(collection)
			record.Set("site", payload.SiteID)
			record.Set("method", payload.Method)
			record.Set("path", payload.Path)
			record.Set("ip", payload.IP)
			record.Set("user_agent", payload.UserAgent)
			record.Set("duration_ms", payload.DurationMs)
			if err := txApp.Save(record); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println("Error saving request log batch:", err)
	}
	return collection
}
