package proto

// === Helper Functions ===

// PreferencesToMap converts preference list to map
func PreferencesToMap(prefs []*Preference) map[string]interface{} {
	result := make(map[string]interface{})
	for _, p := range prefs {
		result[p.Key] = p.Value
	}
	return result
}

// IdentityToDecisionContext creates DecisionContext from PersonalIdentity
func IdentityToDecisionContext(identity *PersonalIdentity, prefs []*Preference, decisions []*Decision) *DecisionContext {
	return &DecisionContext{
		UserId:            identity.Id,
		Personality:       identity.Personality,
		ValueSystem:       identity.ValueSystem,
		Interests:         identity.Interests,
		SpeakingTone:      "",
		ResponseLength:    "",
		ValuePriority:     []string{},
		RecentDecisions:   decisions,
		ActivePreferences: PreferencesToMap(prefs),
	}
}

// NewDecisionOption creates a new decision option
func NewDecisionOption(name string, description string, attributes map[string]interface{}) *DecisionOption {
	return &DecisionOption{
		Id:          GenerateID(),
		Name:        name,
		Description: description,
		Attributes:  attributes,
		Pros:        []string{},
		Cons:        []string{},
	}
}

// GenerateID generates a unique ID (placeholder)
func GenerateID() string {
	// In real implementation, use UUID or similar
	return "id_" + RandString(16)
}

// RandString generates random string (placeholder)
func RandString(n int) string {
	// Simplified placeholder - real implementation should use crypto/rand
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// === Memory Helper Functions ===

// NewMemory creates a new memory entry
func NewMemory(userID string, memType MemoryType, content string, importance float64) *Memory {
	return &Memory{
		Id:          GenerateID(),
		UserId:      userID,
		Type:        memType,
		Content:     content,
		Importance:  importance,
		Layer:       MemoryLayerL1,
		DecayFactor: 1.0,
		Timestamp:   Now(),
		LastAccessed: Now(),
		AccessCount: 0,
		Tags:        []string{},
	}
}

// Now returns current timestamp (placeholder)
func Now() int64 {
	// Real implementation should use time.Now().Unix()
	return 1700000000
}

// === Value Helper Functions ===

// NewValueSystem creates default value system
func NewValueSystem() *ValueSystem {
	return &ValueSystem{
		Privacy:       0.5,
		Efficiency:    0.5,
		Health:        0.5,
		Family:        0.5,
		Career:        0.5,
		Entertainment: 0.5,
		Learning:      0.5,
		Social:        0.5,
		Finance:       0.5,
		Environment:   0.5,
	}
}

// GetPriority returns priority score for a value key
func (vs *ValueSystem) GetPriority(key string) float64 {
	switch key {
	case "privacy":
		return vs.Privacy
	case "efficiency":
		return vs.Efficiency
	case "health":
		return vs.Health
	case "family":
		return vs.Family
	case "career":
		return vs.Career
	case "entertainment":
		return vs.Entertainment
	case "learning":
		return vs.Learning
	case "social":
		return vs.Social
	case "finance":
		return vs.Finance
	case "environment":
		return vs.Environment
	default:
		return 0.5
	}
}

// === Personality Helper Functions ===

// NewPersonality creates default personality
func NewPersonality() *Personality {
	return &Personality{
		MbtiType:         "INFP", // Default type
		MbtiEi:          0.0,    // Balanced
		MbtiSn:          0.0,
		MbtiTf:          0.0,
		MbtiJp:          0.0,
		MbtiConfidence:   0.0,
		Openness:          0.5,
		Conscientiousness: 0.5,
		Extraversion:      0.5,
		Agreeableness:     0.5,
		Neuroticism:       0.5,
		StabilityScore:    0.0,
		ObservedCount:     0,
	}
}

// GetMBTITypeFromDimensions determines MBTI type from dimensions
func GetMBTITypeFromDimensions(ei, sn, tf, jp float64) string {
	// EI: negative = Introvert, positive = Extrovert
	// SN: negative = Sensing, positive = Intuition
	// TF: negative = Thinking, positive = Feeling
	// JP: negative = Judging, positive = Perceiving

	e := "E"
	if ei < 0 {
		e = "I"
	}

	n := "N"
	if sn < 0 {
		n = "S"
	}

	f := "F"
	if tf < 0 {
		f = "T"
	}

	p := "P"
	if jp < 0 {
		p = "J"
	}

	return e + n + f + p
}

// === Interest Helper Functions ===

// NewInterest creates a new interest
func NewInterest(category, name string, keywords []string, level float64) *Interest {
	return &Interest{
		Id:       GenerateID(),
		Category: category,
		Name:     name,
		Keywords: keywords,
		Level:    level,
		Since:    Now(),
		LastActive: Now(),
	}
}

// === Preference Helper Functions ===

// NewPreference creates a new preference
func NewPreference(userID, key string, value interface{}, confidence float64, source string) *Preference {
	return &Preference{
		Id:         GenerateID(),
		UserId:     userID,
		Key:        key,
		Value:      value,
		Confidence: confidence,
		Source:     source,
		CreatedAt:  Now(),
		UpdatedAt:  Now(),
		AccessCount: 0,
	}
}

// === Voice Helper Functions ===

// NewVoiceProfile creates default voice profile
func NewVoiceProfile(userID string) *VoiceProfile {
	return &VoiceProfile{
		Id:            GenerateID(),
		VoiceType:     "preset",
		PresetVoiceId: "default",
		Pitch:         1.0,
		Speed:         1.0,
		Volume:        1.0,
		CreatedAt:     Now(),
		UpdatedAt:     Now(),
	}
}

// === Session Helper Functions ===

// NewSession creates a new session
func NewSession(userID, agentID string, context map[string]interface{}) *Session {
	return &Session{
		Id:           GenerateID(),
		UserId:       userID,
		AgentId:      agentID,
		Status:       "active",
		Context:      context,
		ActiveMemory: []*Memory{},
		StartedAt:    Now(),
		LastActiveAt: Now(),
	}
}