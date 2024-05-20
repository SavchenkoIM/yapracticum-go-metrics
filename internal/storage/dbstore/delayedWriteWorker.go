package dbstore

import (
	"context"
	"time"
	"yaprakticum-go-track2/internal/shared"
)

func (ms *DBStore) delayedWriteWorker(ctx context.Context) {
	shared.Logger.Info("Cached write worker routine started")
	for ctx.Err() == nil {
		time.Sleep(ms.cachedWriteInterval)
		// Forbid data changes
		ms.wgWorker.Add(1)
		// Wait until all handlers finished filling data
		ms.wgServer.Wait()

		// Write data
		ms.delayedWriteResult = ms.WriteDataMultiBatchRaw(ctx, ms.cachedGauges.GetData(), ms.cachedCounters.GetData())
		ms.cachedCounters.Clear()
		ms.cachedGauges.Clear()

		// Report operation finished
		ms.delayedWriteCond.Broadcast()

		// Allow data changes
		ms.wgWorker.Done()
	}
	shared.Logger.Info("Cached write worker routine terminated")
}
