package websocket

import (
	"time"
)

// getCurrentTimestampMs returns current timestamp in milliseconds
func getCurrentTimestampMs() int64 {
	return time.Now().UnixMilli()
}

// SyncBroadcaster handles synchronization state broadcasting
type SyncBroadcaster struct {
	broadcaster *StateBroadcaster
}

// NewSyncBroadcaster creates a new sync broadcaster
func NewSyncBroadcaster(broadcaster *StateBroadcaster) *SyncBroadcaster {
	return &SyncBroadcaster{
		broadcaster: broadcaster,
	}
}

// BroadcastSyncComplete broadcasts when sync is complete
func (s *SyncBroadcaster) BroadcastSyncComplete(identityID string, dataType string, version int64) {
	data := map[string]interface{}{
		"status":     "complete",
		"data_type":  dataType,
		"version":    version,
	}
	s.broadcaster.BroadcastSyncUpdate(identityID, dataType, data, version)
}

// BroadcastSyncConflict broadcasts when sync conflict is detected
func (s *SyncBroadcaster) BroadcastSyncConflict(identityID string, dataType string, localVersion, remoteVersion int64) {
	data := map[string]interface{}{
		"status":         "conflict",
		"data_type":      dataType,
		"local_version":  localVersion,
		"remote_version": remoteVersion,
	}
	s.broadcaster.BroadcastSyncUpdate(identityID, dataType, data, max(localVersion, remoteVersion)+1)
}

// BroadcastSyncRequest broadcasts request for agent to sync
func (s *SyncBroadcaster) BroadcastSyncRequest(identityID string, dataType string, lastVersion int64) {
	data := map[string]interface{}{
		"status":       "request",
		"data_type":    dataType,
		"last_version": lastVersion,
	}
	s.broadcaster.BroadcastSyncUpdate(identityID, dataType, data, lastVersion)
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}