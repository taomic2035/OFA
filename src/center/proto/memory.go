package proto

// === Memory System Messages ===

// MemoryType enum
type MemoryType string

const (
	MemoryTypeEpisodic   MemoryType = "episodic"
	MemoryTypeSemantic   MemoryType = "semantic"
	MemoryTypeProcedural MemoryType = "procedural"
	MemoryTypePreference MemoryType = "preference"
	MemoryTypeEmotional  MemoryType = "emotional"
	MemoryTypeFact       MemoryType = "fact"
	MemoryTypeSkill      MemoryType = "skill"
)

// MemoryLayer enum
type MemoryLayer string

const (
	MemoryLayerL1 MemoryLayer = "L1"
	MemoryLayerL2 MemoryLayer = "L2"
	MemoryLayerL3 MemoryLayer = "L3"
)

// Memory message
type Memory struct {
	Id           string            `json:"id"`
	UserId       string            `json:"user_id"`
	Type         MemoryType        `json:"type"`
	Category     string            `json:"category"`
	Content      string            `json:"content"`
	Summary      string            `json:"summary"`
	Importance   float64           `json:"importance"`
	Priority     int32             `json:"priority"`
	Emotion      string            `json:"emotion"`
	EmotionScore float64           `json:"emotion_score"`
	Tags         []string          `json:"tags"`
	Entities     []string          `json:"entities"`
	Source       string            `json:"source"`
	SourceAgent  string            `json:"source_agent"`
	SourceApp    string            `json:"source_app"`
	Timestamp    int64             `json:"timestamp"`
	LastAccessed int64             `json:"last_accessed"`
	AccessCount  int32             `json:"access_count"`
	Layer        MemoryLayer       `json:"layer"`
	DecayFactor  float64           `json:"decay_factor"`
	RelatedIds   []string          `json:"related_ids"`
	ParentId     string            `json:"parent_id"`
	CreatedAt    int64             `json:"created_at"`
	UpdatedAt    int64             `json:"updated_at"`
}

// MemoryStats message
type MemoryStats struct {
	UserId         string          `json:"user_id"`
	TotalCount     int32           `json:"total_count"`
	CountByType    map[string]int32 `json:"count_by_type"`
	CountByLayer   map[string]int32 `json:"count_by_layer"`
	TotalSize      int64           `json:"total_size"`
	OldestMemory   int64           `json:"oldest_memory"`
	NewestMemory   int64           `json:"newest_memory"`
	AvgImportance  float64         `json:"avg_importance"`
	AvgAccessCount float64         `json:"avg_access_count"`
	TopCategories  []string        `json:"top_categories"`
	TopTags        []string        `json:"top_tags"`
}

// MemoryQuery message
type MemoryQuery struct {
	UserId        string       `json:"user_id"`
	Types         []MemoryType `json:"types"`
	Categories    []string     `json:"categories"`
	Tags          []string     `json:"tags"`
	Layer         MemoryLayer  `json:"layer"`
	Keywords      string       `json:"keywords"`
	Semantic      string       `json:"semantic"`
	StartTime     int64        `json:"start_time"`
	EndTime       int64        `json:"end_time"`
	MinImportance float64      `json:"min_importance"`
	Emotion       string       `json:"emotion"`
	Source        string       `json:"source"`
	Limit         int32        `json:"limit"`
	Offset        int32        `json:"offset"`
	SortBy        string       `json:"sort_by"`
	SortDesc      bool         `json:"sort_desc"`
}

// MemoryRecallResult message
type MemoryRecallResult struct {
	Memories  []*Memory `json:"memories"`
	Total     int32     `json:"total"`
	QueryTime int64     `json:"query_time_ms"`
	RecallType string   `json:"recall_type"`
}

// MemoryAssociation message
type MemoryAssociation struct {
	SourceId     string  `json:"source_id"`
	TargetId     string  `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Strength     float64 `json:"strength"`
	CreatedAt    int64   `json:"created_at"`
}

// ConsolidationResult message
type ConsolidationResult struct {
	UserId        string   `json:"user_id"`
	PromotedToL2  []string `json:"promoted_to_l2"`
	PromotedToL3  []string `json:"promoted_to_l3"`
	Demoted       []string `json:"demoted"`
	Forgotten     []string `json:"forgotten"`
	Merged        []string `json:"merged"`
	ConsolidatedAt int64   `json:"consolidated_at"`
}

// === API Requests/Responses ===

// RememberRequest message
type RememberRequest struct {
	UserId       string      `json:"user_id"`
	Type         MemoryType  `json:"type"`
	Content      string      `json:"content"`
	Importance   float64     `json:"importance"`
	Priority     int32       `json:"priority"`
	Category     string      `json:"category"`
	Tags         []string    `json:"tags"`
	Source       string      `json:"source"`
	Emotion      string      `json:"emotion"`
	EmotionScore float64     `json:"emotion_score"`
	Timestamp    int64       `json:"timestamp"`
}

// RememberResponse message
type RememberResponse struct {
	Success bool    `json:"success"`
	Memory  *Memory `json:"memory"`
	Error   string  `json:"error"`
}

// RecallRequest message
type RecallRequest struct {
	Query *MemoryQuery `json:"query"`
}

// RecallResponse message
type RecallResponse struct {
	Success bool              `json:"success"`
	Result  *MemoryRecallResult `json:"result"`
	Error   string            `json:"error"`
}

// RecallByIDRequest message
type RecallByIDRequest struct {
	Id string `json:"id"`
}

// RecallByIDResponse message
type RecallByIDResponse struct {
	Success bool    `json:"success"`
	Memory  *Memory `json:"memory"`
	Error   string  `json:"error"`
}

// RecallByTypeRequest message
type RecallByTypeRequest struct {
	UserId string     `json:"user_id"`
	Type   MemoryType `json:"type"`
	Limit  int32      `json:"limit"`
}

// RecallByTypeResponse message
type RecallByTypeResponse struct {
	Success  bool      `json:"success"`
	Memories []*Memory `json:"memories"`
	Error    string    `json:"error"`
}

// RecallByTimeRequest message
type RecallByTimeRequest struct {
	UserId string `json:"user_id"`
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
	Limit     int32 `json:"limit"`
}

// RecallByTimeResponse message
type RecallByTimeResponse struct {
	Success  bool      `json:"success"`
	Memories []*Memory `json:"memories"`
	Error    string    `json:"error"`
}

// RecallRecentRequest message
type RecallRecentRequest struct {
	UserId string `json:"user_id"`
	Limit  int32  `json:"limit"`
}

// RecallRecentResponse message
type RecallRecentResponse struct {
	Success  bool      `json:"success"`
	Memories []*Memory `json:"memories"`
	Error    string    `json:"error"`
}

// RecallSemanticRequest message
type RecallSemanticRequest struct {
	UserId string `json:"user_id"`
	Query  string `json:"query"`
	Limit  int32  `json:"limit"`
}

// RecallSemanticResponse message
type RecallSemanticResponse struct {
	Success  bool      `json:"success"`
	Memories []*Memory `json:"memories"`
	Error    string    `json:"error"`
}

// ForgetRequest message
type ForgetRequest struct {
	Id string `json:"id"`
}

// ForgetResponse message
type ForgetResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// UpdateMemoryRequest message
type UpdateMemoryRequest struct {
	Id      string                 `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}

// UpdateMemoryResponse message
type UpdateMemoryResponse struct {
	Success bool    `json:"success"`
	Memory  *Memory `json:"memory"`
	Error   string  `json:"error"`
}

// AssociateMemoriesRequest message
type AssociateMemoriesRequest struct {
	SourceId     string  `json:"source_id"`
	TargetId     string  `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Strength     float64 `json:"strength"`
}

// AssociateMemoriesResponse message
type AssociateMemoriesResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ConsolidateRequest message
type ConsolidateRequest struct {
	UserId string `json:"user_id"`
}

// ConsolidateResponse message
type ConsolidateResponse struct {
	Success  bool                  `json:"success"`
	Result   *ConsolidationResult  `json:"result"`
	Error    string                `json:"error"`
}

// GetMemoryStatsRequest message
type GetMemoryStatsRequest struct {
	UserId string `json:"user_id"`
}

// GetMemoryStatsResponse message
type GetMemoryStatsResponse struct {
	Success bool         `json:"success"`
	Stats   *MemoryStats `json:"stats"`
	Error   string       `json:"error"`
}