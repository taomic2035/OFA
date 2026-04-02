package proto

// === Decision System Messages ===

// Decision message
type Decision struct {
	Id                 string            `json:"id"`
	UserId             string            `json:"user_id"`
	Scenario           string            `json:"scenario"`
	ScenarioType       string            `json:"scenario_type"`
	Context            map[string]interface{} `json:"context"`
	Options            []DecisionOption  `json:"options"`
	SelectedIndex      int32             `json:"selected_index"`
	SelectedOption     *DecisionOption   `json:"selected_option"`
	SelectedReason     string            `json:"selected_reason"`
	AppliedValues      []string          `json:"applied_values"`
	AppliedRules       []string          `json:"applied_rules"`
	AppliedPreferences []string          `json:"applied_preferences"`
	ScoreDetails       map[string]float64 `json:"score_details"`
	Ranking            []int32           `json:"ranking"`
	Outcome            string            `json:"outcome"`
	UserFeedback       string            `json:"user_feedback"`
	OutcomeScore       float64           `json:"outcome_score"`
	ExecutedAt         int64             `json:"executed_at"`
	CompletedAt        int64             `json:"completed_at"`
	AutoDecided        bool              `json:"auto_decided"`
	Confidence         float64           `json:"confidence"`
	Tags               []string          `json:"tags"`
	CreatedAt          int64             `json:"created_at"`
	UpdatedAt          int64             `json:"updated_at"`
}

// DecisionOption message
type DecisionOption struct {
	Id            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Attributes    map[string]interface{} `json:"attributes"`
	Score         float64                `json:"score"`
	ScoreBreakdown map[string]float64    `json:"score_breakdown"`
	Pros          []string               `json:"pros"`
	Cons          []string               `json:"cons"`
	Rank          int32                  `json:"rank"`
}

// DecisionContext message
type DecisionContext struct {
	UserId            string                 `json:"user_id"`
	Personality       *Personality           `json:"personality"`
	ValueSystem       *ValueSystem           `json:"value_system"`
	Interests         []*Interest            `json:"interests"`
	SpeakingTone      string                 `json:"speaking_tone"`
	ResponseLength    string                 `json:"response_length"`
	ValuePriority     []string               `json:"value_priority"`
	RecentDecisions   []*Decision            `json:"recent_decisions"`
	ActivePreferences map[string]interface{} `json:"active_preferences"`
}

// DecisionResult message
type DecisionResult struct {
	Decision        *Decision        `json:"decision"`
	Alternatives    []*DecisionOption `json:"alternatives"`
	Explanation     string           `json:"explanation"`
	Confidence      float64          `json:"confidence"`
	NeedsUserInput  bool             `json:"needs_user_input"`
	UncertainReason string           `json:"uncertain_reason"`
}

// DecisionQuery message
type DecisionQuery struct {
	UserId      string `json:"user_id"`
	Scenario    string `json:"scenario"`
	Outcome     string `json:"outcome"`
	AutoDecided *bool  `json:"auto_decided"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Limit       int32  `json:"limit"`
	Offset      int32  `json:"offset"`
}

// DecisionStats message
type DecisionStats struct {
	UserId           string            `json:"user_id"`
	TotalDecisions   int32             `json:"total_decisions"`
	AutoDecisions    int32             `json:"auto_decisions"`
	ManualDecisions  int32             `json:"manual_decisions"`
	SatisfiedCount   int32             `json:"satisfied_count"`
	UnsatisfiedCount int32             `json:"unsatisfied_count"`
	AvgOutcomeScore  float64           `json:"avg_outcome_score"`
	CountByScenario  map[string]int32  `json:"count_by_scenario"`
	TopScenarios     []string          `json:"top_scenarios"`
	ValueUsage       map[string]int32  `json:"value_usage"`
	PreferenceHits   map[string]int32  `json:"preference_hits"`
}

// ScoringRule message
type ScoringRule struct {
	Id            string  `json:"id"`
	Name          string  `json:"name"`
	Attribute     string  `json:"attribute"`
	Weight        float64 `json:"weight"`
	ValueMatch    string  `json:"value_match"`
	ScoreIfMatch  float64 `json:"score_if_match"`
	ScoreIfNoMatch float64 `json:"score_if_no_match"`
}

// Constraint message
type Constraint struct {
	Id          string      `json:"id"`
	Type        string      `json:"type"`
	Attribute   string      `json:"attribute"`
	Operator    string      `json:"operator"`
	Value       interface{} `json:"value"`
	Penalty     float64     `json:"penalty"`
	Description string      `json:"description"`
}

// DecisionScenario message
type DecisionScenario struct {
	Id             string          `json:"id"`
	Name           string          `json:"name"`
	Category       string          `json:"category"`
	Description    string          `json:"description"`
	RequiredValues []string        `json:"required_values"`
	RequiredPrefs  []string        `json:"required_prefs"`
	ScoringRules   []ScoringRule   `json:"scoring_rules"`
	Constraints    []Constraint    `json:"constraints"`
}

// === API Requests/Responses ===

// DecideRequest message
type DecideRequest struct {
	UserId    string                 `json:"user_id"`
	Scenario  string                 `json:"scenario"`
	Options   []DecisionOption       `json:"options"`
	Context   map[string]interface{} `json:"context"`
}

// DecideResponse message
type DecideResponse struct {
	Success bool            `json:"success"`
	Result  *DecisionResult `json:"result"`
	Error   string          `json:"error"`
}

// QuickDecideRequest message
type QuickDecideRequest struct {
	UserId    string                 `json:"user_id"`
	Scenario  string                 `json:"scenario"`
	Options   []DecisionOption       `json:"options"`
}

// QuickDecideResponse message
type QuickDecideResponse struct {
	Success bool            `json:"success"`
	Result  *DecisionResult `json:"result"`
	Error   string          `json:"error"`
}

// ConfirmDecisionRequest message
type ConfirmDecisionRequest struct {
	DecisionId  string `json:"decision_id"`
	OptionIndex int32  `json:"option_index"`
}

// ConfirmDecisionResponse message
type ConfirmDecisionResponse struct {
	Success  bool       `json:"success"`
	Decision *Decision  `json:"decision"`
	Error    string     `json:"error"`
}

// RecordOutcomeRequest message
type RecordOutcomeRequest struct {
	DecisionId string  `json:"decision_id"`
	Outcome    string  `json:"outcome"`
	Feedback   string  `json:"feedback"`
	Score      float64 `json:"score"`
}

// RecordOutcomeResponse message
type RecordOutcomeResponse struct {
	Success  bool      `json:"success"`
	Decision *Decision `json:"decision"`
	Error    string    `json:"error"`
}

// GetDecisionRequest message
type GetDecisionRequest struct {
	Id string `json:"id"`
}

// GetDecisionResponse message
type GetDecisionResponse struct {
	Success  bool      `json:"success"`
	Decision *Decision `json:"decision"`
	Error    string    `json:"error"`
}

// GetDecisionHistoryRequest message
type GetDecisionHistoryRequest struct {
	Query *DecisionQuery `json:"query"`
}

// GetDecisionHistoryResponse message
type GetDecisionHistoryResponse struct {
	Success   bool        `json:"success"`
	Decisions []*Decision `json:"decisions"`
	Total     int32       `json:"total"`
	Error     string      `json:"error"`
}

// GetDecisionStatsRequest message
type GetDecisionStatsRequest struct {
	UserId string `json:"user_id"`
}

// GetDecisionStatsResponse message
type GetDecisionStatsResponse struct {
	Success bool            `json:"success"`
	Stats   *DecisionStats  `json:"stats"`
	Error   string          `json:"error"`
}