package workflow

import (
	"context"
	"log"
	"sync"
	"time"
)

// WorkflowScheduler schedules and triggers workflows
type WorkflowScheduler struct {
	engine   *WorkflowEngine
	triggers sync.Map // map[string]*ScheduledTrigger

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// ScheduledTrigger represents a scheduled trigger
type ScheduledTrigger struct {
	WorkflowID string
	Trigger    Trigger
	NextRun    time.Time
	LastRun    time.Time
	Active     bool
}

// NewWorkflowScheduler creates a new workflow scheduler
func NewWorkflowScheduler(engine *WorkflowEngine) *WorkflowScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkflowScheduler{
		engine: engine,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins scheduling
func (s *WorkflowScheduler) Start() {
	go s.scheduleLoop()
}

// Stop stops the scheduler
func (s *WorkflowScheduler) Stop() {
	s.cancel()
}

// scheduleLoop checks for scheduled workflows to run
func (s *WorkflowScheduler) scheduleLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkTriggers()
		}
	}
}

// checkTriggers checks all triggers and fires if needed
func (s *WorkflowScheduler) checkTriggers() {
	s.triggers.Range(func(key, value interface{}) bool {
		trigger := value.(*ScheduledTrigger)

		if !trigger.Active {
			return true
		}

		if time.Now().After(trigger.NextRun) || time.Now().Equal(trigger.NextRun) {
			s.fireTrigger(trigger)
			s.updateNextRun(trigger)
		}

		return true
	})
}

// fireTrigger fires a scheduled trigger
func (s *WorkflowScheduler) fireTrigger(trigger *ScheduledTrigger) {
	log.Printf("Firing scheduled trigger for workflow %s", trigger.WorkflowID)

	execID, err := s.engine.ExecuteWorkflow(trigger.WorkflowID, nil)
	if err != nil {
		log.Printf("Failed to execute scheduled workflow %s: %v", trigger.WorkflowID, err)
		return
	}

	trigger.LastRun = time.Now()
	log.Printf("Scheduled workflow %s started: %s", trigger.WorkflowID, execID)
}

// updateNextRun updates the next run time
func (s *WorkflowScheduler) updateNextRun(trigger *ScheduledTrigger) {
	switch trigger.Trigger.Type {
	case TriggerTypeSchedule:
		s.updateScheduleTrigger(trigger)
	default:
		// One-time trigger, deactivate
		trigger.Active = false
	}
}

// updateScheduleTrigger calculates next run time
func (s *WorkflowScheduler) updateScheduleTrigger(trigger *ScheduledTrigger) {
	config := trigger.Trigger.Config

	// Parse schedule config
	if interval, ok := config["interval"].(string); ok {
		// Parse interval (e.g., "1h", "30m", "24h")
		duration, err := time.ParseDuration(interval)
		if err == nil {
			trigger.NextRun = trigger.LastRun.Add(duration)
			return
		}
	}

	if cronExpr, ok := config["cron"].(string); ok {
		// Parse cron expression (simplified)
		trigger.NextRun = s.parseCron(cronExpr, trigger.LastRun)
		return
	}

	// Default: daily at midnight
	trigger.NextRun = trigger.LastRun.Add(24 * time.Hour)
}

// parseCron parses a simplified cron expression
func (s *WorkflowScheduler) parseCron(expr string, lastRun time.Time) time.Time {
	// Simplified cron: only handles "daily" and hourly patterns
	switch expr {
	case "0 0 * * *": // Daily at midnight
		next := lastRun.Add(24 * time.Hour)
		return next.Truncate(24 * time.Hour)
	case "0 * * * *": // Hourly
		return lastRun.Add(1 * time.Hour).Truncate(1 * time.Hour)
	default:
		return lastRun.Add(24 * time.Hour)
	}
}

// RegisterTrigger registers a trigger for a workflow
func (s *WorkflowScheduler) RegisterTrigger(workflowID string, trigger Trigger) {
	scheduled := &ScheduledTrigger{
		WorkflowID: workflowID,
		Trigger:    trigger,
		Active:     trigger.Enabled,
	}

	// Calculate initial next run
	if trigger.Type == TriggerTypeSchedule {
		s.updateScheduleTrigger(scheduled)
	} else {
		// Manual/event triggers start inactive
		scheduled.Active = false
	}

	triggerID := workflowID + "-" + trigger.Type
	s.triggers.Store(triggerID, scheduled)
}

// UnregisterTriggers removes all triggers for a workflow
func (s *WorkflowScheduler) UnregisterTriggers(workflowID string) {
	s.triggers.Range(func(key, value interface{}) bool {
		trigger := value.(*ScheduledTrigger)
		if trigger.WorkflowID == workflowID {
			s.triggers.Delete(key)
		}
		return true
	})
}

// TriggerWorkflow manually triggers a workflow
func (s *WorkflowScheduler) TriggerWorkflow(workflowID string, input map[string]interface{}) (string, error) {
	return s.engine.ExecuteWorkflow(workflowID, input)
}

// TriggerByEvent triggers workflows by event
func (s *WorkflowScheduler) TriggerByEvent(eventType string, eventData map[string]interface{}) {
	s.triggers.Range(func(key, value interface{}) bool {
		trigger := value.(*ScheduledTrigger)

		if trigger.Trigger.Type == TriggerTypeEvent {
			config := trigger.Trigger.Config
			if eventType == config["event_type"] {
				// Check filter if exists
				if filter, ok := config["filter"].(map[string]interface{}); ok {
					if !s.matchEventFilter(eventData, filter) {
						return true
					}
				}

				execID, err := s.engine.ExecuteWorkflow(trigger.WorkflowID, eventData)
				if err != nil {
					log.Printf("Failed to trigger workflow %s by event: %v", trigger.WorkflowID, err)
				} else {
					log.Printf("Workflow %s triggered by event: %s", trigger.WorkflowID, execID)
				}
			}
		}

		return true
	})
}

// TriggerByMessage triggers workflows by message
func (s *WorkflowScheduler) TriggerByMessage(msgType string, msgData map[string]interface{}) {
	s.triggers.Range(func(key, value interface{}) bool {
		trigger := value.(*ScheduledTrigger)

		if trigger.Trigger.Type == TriggerTypeMessage {
			config := trigger.Trigger.Config
			if msgType == config["message_type"] {
				execID, err := s.engine.ExecuteWorkflow(trigger.WorkflowID, msgData)
				if err != nil {
					log.Printf("Failed to trigger workflow %s by message: %v", trigger.WorkflowID, err)
				} else {
					log.Printf("Workflow %s triggered by message: %s", trigger.WorkflowID, execID)
				}
			}
		}

		return true
	})
}

// matchEventFilter checks if event data matches filter
func (s *WorkflowScheduler) matchEventFilter(data, filter map[string]interface{}) bool {
	for k, v := range filter {
		if dataValue, ok := data[k]; ok {
			if dataValue != v {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// GetScheduledTriggers returns all scheduled triggers
func (s *WorkflowScheduler) GetScheduledTriggers() []*ScheduledTrigger {
	var triggers []*ScheduledTrigger
	s.triggers.Range(func(key, value interface{}) bool {
		triggers = append(triggers, value.(*ScheduledTrigger))
		return true
	})
	return triggers
}

// GetNextRun returns the next run time for a workflow
func (s *WorkflowScheduler) GetNextRun(workflowID string) time.Time {
	var nextRun time.Time
	s.triggers.Range(func(key, value interface{}) bool {
		trigger := value.(*ScheduledTrigger)
		if trigger.WorkflowID == workflowID && trigger.Active {
			if nextRun.IsZero() || trigger.NextRun.Before(nextRun) {
				nextRun = trigger.NextRun
			}
		}
		return true
	})
	return nextRun
}