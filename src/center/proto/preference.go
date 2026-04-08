package proto

// === Preference System Messages ===

// Preference message
type Preference struct {
	Id          string      `json:"id"`
	UserId      string      `json:"user_id"`
	Category    string      `json:"category"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ValueType   string      `json:"value_type"`
	Confidence  float64     `json:"confidence"`
	Source      string      `json:"source"`
	Context     string      `json:"context"`
	Conditions  []Condition `json:"conditions"`
	AccessCount int32       `json:"access_count"`
	ConfirmCount int32      `json:"confirm_count"`
	RejectCount int32       `json:"reject_count"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	LastUsed    int64       `json:"last_used"`
	Tags        []string    `json:"tags"`
	Notes       string      `json:"notes"`
}

// Condition message
type Condition struct {
	Type     string      `json:"type"`
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"`
}

// PreferenceQuery message
type PreferenceQuery struct {
	UserId        string   `json:"user_id"`
	Categories    []string `json:"categories"`
	Keys          []string `json:"keys"`
	MinConfidence float64  `json:"min_confidence"`
	Source        string   `json:"source"`
	Tags          []string `json:"tags"`
	Limit         int32    `json:"limit"`
	Offset        int32    `json:"offset"`
}

// PreferenceStats message
type PreferenceStats struct {
	UserId          string                    `json:"user_id"`
	TotalCount      int32                     `json:"total_count"`
	CountByCategory map[string]int32          `json:"count_by_category"`
	CountBySource   map[string]int32          `json:"count_by_source"`
	AvgConfidence   float64                   `json:"avg_confidence"`
	TopPreferences  map[string][]*Preference  `json:"top_preferences"`
}

// PreferenceLearningEvent message
type PreferenceLearningEvent struct {
	Id        string                 `json:"id"`
	UserId    string                 `json:"user_id"`
	Type      string                 `json:"type"`
	Category  string                 `json:"category"`
	Context   map[string]interface{} `json:"context"`
	Options   []interface{}          `json:"options"`
	Selected  interface{}            `json:"selected"`
	Feedback  string                 `json:"feedback"`
	Sentiment float64                `json:"sentiment"`
	Timestamp int64                  `json:"timestamp"`
}

// Recommendation message
type Recommendation struct {
	Id        string                 `json:"id"`
	UserId    string                 `json:"user_id"`
	Category  string                 `json:"category"`
	Item      interface{}            `json:"item"`
	Score     float64                `json:"score"`
	Reason    string                 `json:"reason"`
	BasedOn   []string               `json:"based_on"`
	Context   map[string]interface{} `json:"context"`
	ExpiresAt int64                  `json:"expires_at"`
	CreatedAt int64                  `json:"created_at"`
}

// === API Requests/Responses ===

// SetPreferenceRequest is defined in api.go
// SetPreferenceResponse is defined in api.go
// GetPreferenceRequest is defined in api.go
// GetPreferenceResponse is defined in api.go

// GetPreferencesRequest message
type GetPreferencesRequest struct {
	Query *PreferenceQuery `json:"query"`
}

// GetPreferencesResponse message
type GetPreferencesResponse struct {
	Success     bool          `json:"success"`
	Preferences []*Preference `json:"preferences"`
	Total       int32         `json:"total"`
	Error       string        `json:"error"`
}

// DeletePreferenceRequest is defined in api.go
// DeletePreferenceResponse is defined in api.go

// ConfirmPreferenceRequest message
type ConfirmPreferenceRequest struct {
	Id string `json:"id"`
}

// ConfirmPreferenceResponse message
type ConfirmPreferenceResponse struct {
	Success    bool        `json:"success"`
	Preference *Preference `json:"preference"`
	Error      string      `json:"error"`
}

// RejectPreferenceRequest message
type RejectPreferenceRequest struct {
	Id string `json:"id"`
}

// RejectPreferenceResponse message
type RejectPreferenceResponse struct {
	Success    bool        `json:"success"`
	Preference *Preference `json:"preference"`
	Error      string      `json:"error"`
}

// LearnFromEventRequest message
type LearnFromEventRequest struct {
	Event *PreferenceLearningEvent `json:"event"`
}

// LearnFromEventResponse message
type LearnFromEventResponse struct {
	Success    bool        `json:"success"`
	Preference *Preference `json:"preference"`
	Error      string      `json:"error"`
}

// InferPreferenceRequest message
type InferPreferenceRequest struct {
	UserId   string `json:"user_id"`
	Category string `json:"category"`
}

// InferPreferenceResponse message
type InferPreferenceResponse struct {
	Success     bool          `json:"success"`
	Preferences []*Preference `json:"preferences"`
	Error       string        `json:"error"`
}

// GetContextualPreferencesRequest message
type GetContextualPreferencesRequest struct {
	UserId  string                 `json:"user_id"`
	Context map[string]interface{} `json:"context"`
}

// GetContextualPreferencesResponse message
type GetContextualPreferencesResponse struct {
	Success     bool          `json:"success"`
	Preferences []*Preference `json:"preferences"`
	Error       string        `json:"error"`
}

// GetRecommendationsRequest message
type GetRecommendationsRequest struct {
	UserId   string                 `json:"user_id"`
	Category string                 `json:"category"`
	Context  map[string]interface{} `json:"context"`
}

// GetRecommendationsResponse message
type GetRecommendationsResponse struct {
	Success        bool              `json:"success"`
	Recommendations []*Recommendation `json:"recommendations"`
	Error          string            `json:"error"`
}

// GetPreferenceStatsRequest message
type GetPreferenceStatsRequest struct {
	UserId string `json:"user_id"`
}

// GetPreferenceStatsResponse message
type GetPreferenceStatsResponse struct {
	Success bool            `json:"success"`
	Stats   *PreferenceStats `json:"stats"`
	Error   string          `json:"error"`
}