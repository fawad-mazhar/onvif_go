package events

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

const (
	MaxSubscriptions           = 16
	DefaultPullPointTimeout    = 60 * time.Second
	DefaultSubscriptionTimeout = 600 * time.Second
)

// ServiceContext holds the configuration and state for the events service
type ServiceContext struct {
	Port          int
	Events        []Event
	subscriptions map[int]*Subscription
	nextSubID     int
	subMutex      sync.RWMutex
}

// Event represents an event configuration
type Event struct {
	Topic       string
	Producer    string
	SourceName  string
	SourceType  string
	SourceValue string
	State       bool
	Timestamp   time.Time
}

// Subscription represents an event subscription
type Subscription struct {
	ID              int
	Type            SubscriptionType
	Address         string
	TopicExpression string
	ExpireTime      time.Time
	Created         time.Time
	PendingMessages []EventMessage
	messageMutex    sync.RWMutex
}

// SubscriptionType defines the type of subscription
type SubscriptionType int

const (
	PullPointSubscription SubscriptionType = iota
	BaseSubscription
)

// EventMessage represents an event message
type EventMessage struct {
	Topic       string
	Timestamp   time.Time
	Property    string
	SourceName  string
	SourceValue string
	DataName    string
	DataValue   string
}

// NewServiceContext creates a new events service context
func NewServiceContext() *ServiceContext {
	return &ServiceContext{
		subscriptions: make(map[int]*Subscription),
		nextSubID:     1,
	}
}

// createEventElement creates an XML element for an event
func (s *ServiceContext) createEventElement(event Event) string {
	eventElement := fmt.Sprintf(`
                <tev:EventTopic>
                    <tt:Topic>%s</tt:Topic>
                    <tt:Producer>%s</tt:Producer>
                </tev:EventTopic>`, event.Topic, event.Producer)

	return eventElement
}

// getSubscriptionByID retrieves a subscription by ID
func (s *ServiceContext) getSubscriptionByID(id int) (*Subscription, bool) {
	s.subMutex.RLock()
	defer s.subMutex.RUnlock()
	sub, exists := s.subscriptions[id]
	return sub, exists
}

// createSubscription creates a new subscription
func (s *ServiceContext) createSubscription(subType SubscriptionType, address, topicExpression string, duration time.Duration) *Subscription {
	s.subMutex.Lock()
	defer s.subMutex.Unlock()

	if len(s.subscriptions) >= MaxSubscriptions {
		return nil
	}

	sub := &Subscription{
		ID:              s.nextSubID,
		Type:            subType,
		Address:         address,
		TopicExpression: topicExpression,
		Created:         time.Now(),
		ExpireTime:      time.Now().Add(duration),
		PendingMessages: make([]EventMessage, 0),
	}

	s.subscriptions[s.nextSubID] = sub
	s.nextSubID++
	if s.nextSubID > 65535 {
		s.nextSubID = 1
	}

	logger.Debug("Created subscription ID=%d, type=%d, expires=%s", sub.ID, sub.Type, sub.ExpireTime.Format(time.RFC3339))
	return sub
}

// removeSubscription removes a subscription by ID
func (s *ServiceContext) removeSubscription(id int) bool {
	s.subMutex.Lock()
	defer s.subMutex.Unlock()

	if _, exists := s.subscriptions[id]; exists {
		delete(s.subscriptions, id)
		logger.Debug("Removed subscription ID=%d", id)
		return true
	}
	return false
}

// cleanExpiredSubscriptions removes expired subscriptions
func (s *ServiceContext) cleanExpiredSubscriptions() {
	s.subMutex.Lock()
	defer s.subMutex.Unlock()

	now := time.Now()
	for id, sub := range s.subscriptions {
		if now.After(sub.ExpireTime) {
			delete(s.subscriptions, id)
			logger.Debug("Expired subscription ID=%d removed", id)
		}
	}
}

// isTopicMatching checks if a topic matches the expression
func isTopicMatching(expression, topic string) bool {
	if expression == "" {
		return true
	}
	// Simple topic matching - in real implementation, this would be more complex
	return expression == topic
}

// AddEventMessage adds an event message to matching subscriptions (public for external event generation)
func (s *ServiceContext) AddEventMessage(topic string, state bool, timestamp time.Time) {
	s.subMutex.RLock()
	defer s.subMutex.RUnlock()

	for _, sub := range s.subscriptions {
		if isTopicMatching(sub.TopicExpression, topic) {
			sub.messageMutex.Lock()

			dataName := "State"
			dataValue := "false"
			if state {
				dataValue = "true"
			}
			if topic == "tns1:Device/Trigger/Relay" {
				dataName = "LogicalState"
				if state {
					dataValue = "active"
				} else {
					dataValue = "inactive"
				}
			}

			msg := EventMessage{
				Topic:       topic,
				Timestamp:   timestamp,
				Property:    "Changed",
				SourceName:  "VideoSourceToken",
				SourceValue: "VideoSource_1",
				DataName:    dataName,
				DataValue:   dataValue,
			}

			sub.PendingMessages = append(sub.PendingMessages, msg)
			sub.messageMutex.Unlock()
		}
	}
}

// StartEventGenerator starts a background goroutine that generates test events
func (s *ServiceContext) StartEventGenerator() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		eventState := false
		for range ticker.C {
			// Generate sample events for testing
			s.AddEventMessage("tns1:VideoSource/MotionAlarm", eventState, time.Now())
			s.AddEventMessage("tns1:Device/Trigger/Relay", eventState, time.Now())
			eventState = !eventState
			logger.Debug("Generated test events with state: %v", eventState)
		}
	}()
}

// HTTP-compatible methods that write to http.ResponseWriter

// GetServiceCapabilitiesHTTP handles the GetServiceCapabilities ONVIF events service method via HTTP
func (s *ServiceContext) GetServiceCapabilitiesHTTP(w http.ResponseWriter) error {
	replacements := map[string]string{
		"%EVENTS_BASESUBSCRIPTION%": "true",
		"%EVENTS_PULLPOINT%":        "true",
	}

	return utils.ProcessServiceTemplate(w, "events", "GetServiceCapabilities", replacements)
}

// CreatePullPointSubscriptionHTTP handles the CreatePullPointSubscription ONVIF events service method via HTTP
func (s *ServiceContext) CreatePullPointSubscriptionHTTP(w http.ResponseWriter, r *http.Request) error {
	logger.Info("CreatePullPointSubscription received")

	// Clean expired subscriptions
	s.cleanExpiredSubscriptions()

	// Parse timeout duration (default 1 minute for pull point)
	duration := DefaultPullPointTimeout

	// Create pull point subscription
	address := fmt.Sprintf("http://localhost:%d/onvif/events_service", s.Port)
	sub := s.createSubscription(PullPointSubscription, address, "", duration)
	if sub == nil {
		return fmt.Errorf("failed to create subscription: maximum subscriptions reached")
	}

	// Add subscription ID to address
	subscriptionAddress := fmt.Sprintf("%s?sub=%d", address, sub.ID)

	now := time.Now()
	replacements := map[string]string{
		"%ADDRESS%":          subscriptionAddress,
		"%CURRENT_TIME%":     now.Format(time.RFC3339),
		"%TERMINATION_TIME%": sub.ExpireTime.Format(time.RFC3339),
	}

	return utils.ProcessServiceTemplate(w, "events", "CreatePullPointSubscription", replacements)
}

// PullMessagesHTTP handles the PullMessages ONVIF events service method via HTTP
func (s *ServiceContext) PullMessagesHTTP(w http.ResponseWriter, r *http.Request) error {
	logger.Info("PullMessages request received")

	// Parse subscription ID from query parameters
	subIDStr := r.URL.Query().Get("sub")
	if subIDStr == "" {
		return fmt.Errorf("no subscription ID provided")
	}

	subID, err := strconv.Atoi(subIDStr)
	if err != nil || subID <= 0 || subID > 65535 {
		return fmt.Errorf("invalid subscription ID")
	}

	sub, exists := s.getSubscriptionByID(subID)
	if !exists {
		return fmt.Errorf("subscription not found")
	}

	if sub.Type != PullPointSubscription {
		return fmt.Errorf("not a pull point subscription")
	}

	// Wait for messages or timeout (simplified - in real implementation would parse timeout from request)
	timeout := 30 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		sub.messageMutex.RLock()
		hasMessages := len(sub.PendingMessages) > 0
		sub.messageMutex.RUnlock()

		if hasMessages {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	now := time.Now()
	replacements := map[string]string{
		"%CURRENT_TIME%":     now.Format(time.RFC3339),
		"%TERMINATION_TIME%": sub.ExpireTime.Format(time.RFC3339),
	}

	return utils.ProcessServiceTemplate(w, "events", "PullMessages_1", replacements)
}

// SubscribeHTTP handles the Subscribe ONVIF events service method via HTTP
func (s *ServiceContext) SubscribeHTTP(w http.ResponseWriter, r *http.Request) error {
	logger.Info("Subscribe request received")

	// Clean expired subscriptions
	s.cleanExpiredSubscriptions()

	// In a real implementation, we would parse the consumer reference from the SOAP request
	consumerReference := "http://consumer/notify"

	// Create base subscription
	sub := s.createSubscription(BaseSubscription, consumerReference, "", DefaultSubscriptionTimeout)
	if sub == nil {
		return fmt.Errorf("failed to create subscription: maximum subscriptions reached")
	}

	subscriptionAddress := fmt.Sprintf("http://localhost:%d/onvif/events_service?sub=%d", s.Port, sub.ID)

	now := time.Now()
	replacements := map[string]string{
		"%MSG_UUID%":         fmt.Sprintf("uuid-%d", time.Now().UnixNano()),
		"%REL_TO_UUID%":      "uuid-relates-to", // Would parse from request
		"%REFERENCE%":        subscriptionAddress,
		"%CURRENT_TIME%":     now.Format(time.RFC3339),
		"%TERMINATION_TIME%": sub.ExpireTime.Format(time.RFC3339),
	}

	return utils.ProcessServiceTemplate(w, "events", "Subscribe", replacements)
}

// RenewHTTP handles the Renew ONVIF events service method via HTTP
func (s *ServiceContext) RenewHTTP(w http.ResponseWriter, r *http.Request) error {
	logger.Info("Renew request received")

	subIDStr := r.URL.Query().Get("sub")
	if subIDStr == "" {
		return fmt.Errorf("no subscription ID provided")
	}

	subID, err := strconv.Atoi(subIDStr)
	if err != nil || subID <= 0 || subID > 65535 {
		return fmt.Errorf("invalid subscription ID")
	}

	sub, exists := s.getSubscriptionByID(subID)
	if !exists {
		return fmt.Errorf("subscription not found")
	}

	// Extend subscription (default extension)
	s.subMutex.Lock()
	sub.ExpireTime = time.Now().Add(DefaultSubscriptionTimeout)
	s.subMutex.Unlock()

	now := time.Now()
	replacements := map[string]string{
		"%MSG_UUID%":         fmt.Sprintf("uuid-%d", time.Now().UnixNano()),
		"%REL_TO_UUID%":      "uuid-relates-to", // Would parse from request
		"%CURRENT_TIME%":     now.Format(time.RFC3339),
		"%TERMINATION_TIME%": sub.ExpireTime.Format(time.RFC3339),
	}

	return utils.ProcessServiceTemplate(w, "events", "Renew", replacements)
}

// UnsubscribeHTTP handles the Unsubscribe ONVIF events service method via HTTP
func (s *ServiceContext) UnsubscribeHTTP(w http.ResponseWriter, r *http.Request) error {
	logger.Info("Unsubscribe request received")

	subIDStr := r.URL.Query().Get("sub")
	if subIDStr == "" {
		return fmt.Errorf("no subscription ID provided")
	}

	subID, err := strconv.Atoi(subIDStr)
	if err != nil || subID <= 0 || subID > 65535 {
		return fmt.Errorf("invalid subscription ID")
	}

	if !s.removeSubscription(subID) {
		return fmt.Errorf("subscription not found")
	}

	replacements := map[string]string{}
	return utils.ProcessServiceTemplate(w, "events", "Unsubscribe", replacements)
}

// GetEventPropertiesHTTP handles the GetEventProperties ONVIF events service method via HTTP
func (s *ServiceContext) GetEventPropertiesHTTP(w http.ResponseWriter) error {
	return utils.ProcessCollectionServiceTemplate(
		w,
		s.Events,
		s.createEventElement,
		"%EVENTS%",
		"events",
		"GetEventProperties",
	)
}

// SetSynchronizationPointHTTP handles the SetSynchronizationPoint ONVIF events service method via HTTP
func (s *ServiceContext) SetSynchronizationPointHTTP(w http.ResponseWriter, r *http.Request) error {
	logger.Info("SetSynchronizationPoint request received")

	subIDStr := r.URL.Query().Get("sub")
	if subIDStr == "" {
		return fmt.Errorf("no subscription ID provided")
	}

	subID, err := strconv.Atoi(subIDStr)
	if err != nil || subID <= 0 || subID > 65535 {
		return fmt.Errorf("invalid subscription ID")
	}

	sub, exists := s.getSubscriptionByID(subID)
	if !exists {
		return fmt.Errorf("subscription not found")
	}

	// Force initialization messages for all matching events
	now := time.Now()
	for _, event := range s.Events {
		if isTopicMatching(sub.TopicExpression, event.Topic) {
			sub.messageMutex.Lock()

			dataName := "State"
			dataValue := "false"
			if event.State {
				dataValue = "true"
			}
			if event.Topic == "tns1:Device/Trigger/Relay" {
				dataName = "LogicalState"
				if event.State {
					dataValue = "active"
				} else {
					dataValue = "inactive"
				}
			}

			msg := EventMessage{
				Topic:       event.Topic,
				Timestamp:   now,
				Property:    "Initialized",
				SourceName:  event.SourceName,
				SourceValue: event.SourceValue,
				DataName:    dataName,
				DataValue:   dataValue,
			}

			sub.PendingMessages = append(sub.PendingMessages, msg)
			sub.messageMutex.Unlock()
		}
	}

	replacements := map[string]string{}
	return utils.ProcessServiceTemplate(w, "events", "SetSynchronizationPoint", replacements)
}
