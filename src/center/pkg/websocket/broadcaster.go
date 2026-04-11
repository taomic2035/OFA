package websocket

import (
	"log"
	"sync"
)

// StateBroadcaster handles state push notifications to agents
type StateBroadcaster struct {
	mu sync.RWMutex

	// Connection manager reference
	connManager *ConnectionManager

	// State subscribers (identity_id -> []agent_id)
	subscribers sync.Map

	// Version tracking (identity_id -> last_version)
	versions sync.Map

	// Pending updates queue
	pendingQueue chan *StateUpdatePayload
}

// NewStateBroadcaster creates a new state broadcaster
func NewStateBroadcaster(connManager *ConnectionManager) *StateBroadcaster {
	return &StateBroadcaster{
		connManager:  connManager,
		pendingQueue: make(chan *StateUpdatePayload, 1000),
	}
}

// Start starts the broadcaster background worker
func (b *StateBroadcaster) Start() {
	go b.dispatchWorker()
	log.Printf("StateBroadcaster started")
}

// Stop stops the broadcaster
func (b *StateBroadcaster) Stop() {
	// Queue will be drained by dispatchWorker when context ends
	log.Printf("StateBroadcaster stopped")
}

// === Subscription Management ===

// Subscribe registers an agent to receive state updates for an identity
func (b *StateBroadcaster) Subscribe(agentID, identityID string) {
	agents, _ := b.subscribers.LoadOrStore(identityID, &[]string{})
	agentList := agents.(*[]string)

	// Add agent if not already subscribed
	for _, a := range *agentList {
		if a == agentID {
			return
		}
	}
	*agentList = append(*agentList, agentID)

	log.Printf("Agent %s subscribed to identity %s updates", agentID, identityID)
}

// Unsubscribe removes an agent from state updates for an identity
func (b *StateBroadcaster) Unsubscribe(agentID, identityID string) {
	if agents, ok := b.subscribers.Load(identityID); ok {
		agentList := agents.(*[]string)
		newList := make([]string, 0)
		for _, a := range *agentList {
			if a != agentID {
				newList = append(newList, a)
			}
		}
		if len(newList) == 0 {
			b.subscribers.Delete(identityID)
		} else {
			b.subscribers.Store(identityID, &newList)
		}
	}

	log.Printf("Agent %s unsubscribed from identity %s updates", agentID, identityID)
}

// UnsubscribeAll removes an agent from all subscriptions
func (b *StateBroadcaster) UnsubscribeAll(agentID string) {
	b.subscribers.Range(func(key, value interface{}) bool {
		identityID := key.(string)
		agents := value.(*[]string)
		newList := make([]string, 0)
		for _, a := range *agents {
			if a != agentID {
				newList = append(newList, a)
			}
		}
		if len(newList) == 0 {
			b.subscribers.Delete(identityID)
		} else {
			b.subscribers.Store(identityID, &newList)
		}
		return true
	})
}

// GetSubscribers returns agents subscribed to an identity
func (b *StateBroadcaster) GetSubscribers(identityID string) []string {
	if agents, ok := b.subscribers.Load(identityID); ok {
		return *agents.(*[]string)
	}
	return []string{}
}

// === State Update Broadcasting ===

// BroadcastIdentityUpdate broadcasts identity state update
func (b *StateBroadcaster) BroadcastIdentityUpdate(identityID string, data map[string]interface{}, version int64) {
	update := &StateUpdatePayload{
		IdentityID: identityID,
		UpdateType: "identity",
		Data:       data,
		Version:    version,
		Timestamp:  getCurrentTimestamp(),
	}
	b.queueUpdate(update)
}

// BroadcastEmotionUpdate broadcasts emotion state update
func (b *StateBroadcaster) BroadcastEmotionUpdate(identityID string, dominantEmotion string, intensity float64, mood string) {
	data := map[string]interface{}{
		"dominant_emotion": dominantEmotion,
		"intensity":        intensity,
		"mood":             mood,
	}

	update := &StateUpdatePayload{
		IdentityID: identityID,
		UpdateType: "emotion",
		Data:       data,
		Version:    getCurrentTimestamp(),
		Timestamp:  getCurrentTimestamp(),
	}
	b.queueUpdate(update)
}

// BroadcastDeviceUpdate broadcasts device state update
func (b *StateBroadcaster) BroadcastDeviceUpdate(identityID, agentID string, data map[string]interface{}) {
	data["agent_id"] = agentID

	update := &StateUpdatePayload{
		IdentityID: identityID,
		UpdateType: "device",
		Data:       data,
		Version:    getCurrentTimestamp(),
		Timestamp:  getCurrentTimestamp(),
	}
	b.queueUpdate(update)
}

// BroadcastSyncUpdate broadcasts sync state update
func (b *StateBroadcaster) BroadcastSyncUpdate(identityID string, dataType string, data map[string]interface{}, version int64) {
	data["data_type"] = dataType

	update := &StateUpdatePayload{
		IdentityID: identityID,
		UpdateType: "sync",
		Data:       data,
		Version:    version,
		Timestamp:  getCurrentTimestamp(),
	}
	b.queueUpdate(update)
}

// BroadcastCustomUpdate broadcasts a custom state update
func (b *StateBroadcaster) BroadcastCustomUpdate(identityID, updateType string, data map[string]interface{}) {
	update := &StateUpdatePayload{
		IdentityID: identityID,
		UpdateType: updateType,
		Data:       data,
		Version:    getCurrentTimestamp(),
		Timestamp:  getCurrentTimestamp(),
	}
	b.queueUpdate(update)
}

// === Task Broadcasting ===

// BroadcastTaskAssign broadcasts task assignment to agents
func (b *StateBroadcaster) BroadcastTaskAssign(agentID string, task *TaskAssignPayload) error {
	msg := NewMessage(MsgTypeTaskAssign, task)
	return b.connManager.SendMessage(agentID, msg)
}

// BroadcastTaskAssignToIdentity broadcasts task to all agents of an identity
func (b *StateBroadcaster) BroadcastTaskAssignToIdentity(identityID string, task *TaskAssignPayload) int {
	agents := b.connManager.GetAgentsForIdentity(identityID)
	count := 0
	for _, agentID := range agents {
		if b.BroadcastTaskAssign(agentID, task) == nil {
			count++
		}
	}
	return count
}

// === Version Management ===

// GetLastVersion returns the last broadcast version for an identity
func (b *StateBroadcaster) GetLastVersion(identityID string) int64 {
	if version, ok := b.versions.Load(identityID); ok {
		return version.(int64)
	}
	return 0
}

// UpdateVersion updates the version for an identity
func (b *StateBroadcaster) UpdateVersion(identityID string, version int64) {
	b.versions.Store(identityID, version)
}

// === Internal Methods ===

// queueUpdate queues a state update for broadcasting
func (b *StateBroadcaster) queueUpdate(update *StateUpdatePayload) {
	select {
	case b.pendingQueue <- update:
		// Queued successfully
	default:
		log.Printf("State update queue full, dropping update for %s", update.IdentityID)
	}
}

// dispatchWorker dispatches queued state updates to subscribers
func (b *StateBroadcaster) dispatchWorker() {
	for update := range b.pendingQueue {
		b.dispatchUpdate(update)
	}
}

// dispatchUpdate dispatches a state update to all subscribers
func (b *StateBroadcaster) dispatchUpdate(update *StateUpdatePayload) {
	// Get subscribers for this identity
	agents := b.GetSubscribers(update.IdentityID)

	// Also include agents that are bound to this identity
	boundAgents := b.connManager.GetAgentsForIdentity(update.IdentityID)
	for _, agentID := range boundAgents {
		found := false
		for _, a := range agents {
			if a == agentID {
				found = true
				break
			}
		}
		if !found {
			agents = append(agents, agentID)
		}
	}

	if len(agents) == 0 {
		log.Printf("No subscribers for identity %s update", update.IdentityID)
		return
	}

	// Create message
	msg := NewMessage(MsgTypeStateUpdate, update)

	// Send to all agents
	successCount := 0
	for _, agentID := range agents {
		if b.connManager.IsAgentOnline(agentID) {
			if err := b.connManager.SendMessage(agentID, msg); err != nil {
				log.Printf("Failed to send update to agent %s: %v", agentID, err)
			} else {
				successCount++
			}
		}
	}

	// Update version
	b.UpdateVersion(update.IdentityID, update.Version)

	log.Printf("Broadcasted %s update to %d/%d agents (identity: %s, version: %d)",
		update.UpdateType, successCount, len(agents), update.IdentityID, update.Version)
}

// getCurrentTimestamp returns current timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return getCurrentTimestampMs()
}

// EmotionBroadcaster is a specialized broadcaster for emotion updates
type EmotionBroadcaster struct {
	broadcaster *StateBroadcaster
}

// NewEmotionBroadcaster creates a new emotion broadcaster
func NewEmotionBroadcaster(broadcaster *StateBroadcaster) *EmotionBroadcaster {
	return &EmotionBroadcaster{
		broadcaster: broadcaster,
	}
}

// BroadcastEmotionChange broadcasts when dominant emotion changes
func (e *EmotionBroadcaster) BroadcastEmotionChange(identityID string, oldEmotion, newEmotion string, intensity float64) {
	data := map[string]interface{}{
		"old_emotion":     oldEmotion,
		"new_emotion":     newEmotion,
		"dominant_emotion": newEmotion,
		"intensity":       intensity,
		"change_type":     "dominant_change",
	}
	e.broadcaster.BroadcastCustomUpdate(identityID, "emotion_change", data)
}

// BroadcastEmotionTrigger broadcasts when emotion is triggered by an event
func (e *EmotionBroadcaster) BroadcastEmotionTrigger(identityID string, emotionType string, trigger string, intensity float64) {
	data := map[string]interface{}{
		"emotion_type":   emotionType,
		"trigger":        trigger,
		"intensity":      intensity,
		"change_type":    "trigger",
	}
	e.broadcaster.BroadcastCustomUpdate(identityID, "emotion_trigger", data)
}

// BroadcastMoodChange broadcasts when mood changes
func (e *EmotionBroadcaster) BroadcastMoodChange(identityID string, oldMood, newMood string) {
	data := map[string]interface{}{
		"old_mood": oldMood,
		"new_mood": newMood,
		"change_type": "mood_change",
	}
	e.broadcaster.BroadcastCustomUpdate(identityID, "mood_change", data)
}