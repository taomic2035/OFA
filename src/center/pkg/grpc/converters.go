package grpc

import (
	"time"

	"github.com/ofa/center/internal/models"
	"github.com/ofa/center/proto"
)

// === Model <-> Proto Converters ===

// PersonalityToProto converts models.Personality to proto.Personality
func PersonalityToProto(p *models.Personality) *proto.Personality {
	if p == nil {
		return nil
	}
	return &proto.Personality{
		MbtiType:         p.MBTIType,
		MbtiEi:           p.MBTI_EI,
		MbtiSn:           p.MBTI_SN,
		MbtiTf:           p.MBTI_TF,
		MbtiJp:           p.MBTI_JP,
		MbtiConfidence:   p.MBTIConfidence,
		Openness:         p.Openness,
		Conscientiousness: p.Conscientiousness,
		Extraversion:     p.Extraversion,
		Agreeableness:    p.Agreeableness,
		Neuroticism:      p.Neuroticism,
		StabilityScore:   p.StabilityScore,
		ObservedCount:    int32(p.ObservedCount),
		SpeakingTone:     p.SpeakingTone,
		ResponseLength:   p.ResponseLength,
		EmojiUsage:       p.EmojiUsage,
		Tags:             p.Tags,
		CustomTraits:     p.CustomTraits,
	}
}

// PersonalityToModel converts proto.Personality to models.Personality
func PersonalityToModel(p *proto.Personality) *models.Personality {
	if p == nil {
		return nil
	}
	return &models.Personality{
		MBTIType:         p.MbtiType,
		MBTI_EI:          p.MbtiEi,
		MBTI_SN:          p.MbtiSn,
		MBTI_TF:          p.MbtiTf,
		MBTI_JP:          p.MbtiJp,
		MBTIConfidence:   p.MbtiConfidence,
		Openness:         p.Openness,
		Conscientiousness: p.Conscientiousness,
		Extraversion:     p.Extraversion,
		Agreeableness:    p.Agreeableness,
		Neuroticism:      p.Neuroticism,
		StabilityScore:   p.StabilityScore,
		ObservedCount:    int(p.ObservedCount),
		SpeakingTone:     p.SpeakingTone,
		ResponseLength:   p.ResponseLength,
		EmojiUsage:       p.EmojiUsage,
		Tags:             p.Tags,
		CustomTraits:     p.CustomTraits,
	}
}

// ValueSystemToProto converts models.ValueSystem to proto.ValueSystem
func ValueSystemToProto(vs *models.ValueSystem) *proto.ValueSystem {
	if vs == nil {
		return nil
	}
	return &proto.ValueSystem{
		Privacy:       vs.Privacy,
		Efficiency:    vs.Efficiency,
		Health:        vs.Health,
		Family:        vs.Family,
		Career:        vs.Career,
		Entertainment: vs.Entertainment,
		Learning:      vs.Learning,
		Social:        vs.Social,
		Finance:       vs.Finance,
		Environment:   vs.Environment,
	}
}

// ValueSystemToModel converts proto.ValueSystem to models.ValueSystem
func ValueSystemToModel(vs *proto.ValueSystem) *models.ValueSystem {
	if vs == nil {
		return nil
	}
	return &models.ValueSystem{
		Privacy:       vs.Privacy,
		Efficiency:    vs.Efficiency,
		Health:        vs.Health,
		Family:        vs.Family,
		Career:        vs.Career,
		Entertainment: vs.Entertainment,
		Learning:      vs.Learning,
		Social:        vs.Social,
		Finance:       vs.Finance,
		Environment:   vs.Environment,
	}
}

// InterestsToProto converts models.Interest slice to proto.Interest slice
func InterestsToProto(interests []models.Interest) []*proto.Interest {
	result := make([]*proto.Interest, len(interests))
	for i, interest := range interests {
		lastActive := int64(0)
		if !interest.LastActive.IsZero() {
			lastActive = interest.LastActive.Unix()
		}
		result[i] = &proto.Interest{
			Id:         interest.ID,
			Category:   interest.Category,
			Name:       interest.Name,
			Keywords:   interest.Keywords,
			Level:      interest.Level,
			LastActive: lastActive,
		}
	}
	return result
}

// InterestsToModel converts proto.Interest slice to models.Interest slice
func InterestsToModel(interests []*proto.Interest) []models.Interest {
	result := make([]models.Interest, len(interests))
	for i, interest := range interests {
		result[i] = models.Interest{
			ID:         interest.Id,
			Category:   interest.Category,
			Name:       interest.Name,
			Keywords:   interest.Keywords,
			Level:      interest.Level,
			LastActive: time.Unix(interest.LastActive, 0),
		}
	}
	return result
}

// IdentityToProto converts models.PersonalIdentity to proto.PersonalIdentity
func IdentityToProto(id *models.PersonalIdentity) *proto.PersonalIdentity {
	if id == nil {
		return nil
	}
	birthday := int64(0)
	if !id.Birthday.IsZero() {
		birthday = id.Birthday.Unix()
	}
	return &proto.PersonalIdentity{
		Id:              id.ID,
		Name:            id.Name,
		Nickname:        id.Nickname,
		Avatar:          id.Avatar,
		Birthday:        birthday,
		Gender:          id.Gender,
		Location:        id.Location,
		Occupation:      id.Occupation,
		Languages:       id.Languages,
		Timezone:        id.Timezone,
		Personality:     PersonalityToProto(id.Personality),
		ValueSystem:     ValueSystemToProto(id.ValueSystem),
		VoiceProfile:    IdentityVoiceProfileToProto(id.VoiceProfile),
		WritingStyle:    WritingStyleToProto(id.WritingStyle),
		CreatedAt:       id.CreatedAt.Unix(),
		UpdatedAt:       id.UpdatedAt.Unix(),
	}
}

// MemoryToProto converts models.Memory to proto.Memory
func MemoryToProto(m *models.Memory) *proto.Memory {
	if m == nil {
		return nil
	}
	timestamp := int64(0)
	if !m.Timestamp.IsZero() {
		timestamp = m.Timestamp.Unix()
	}
	lastAccessed := int64(0)
	if !m.LastAccessed.IsZero() {
		lastAccessed = m.LastAccessed.Unix()
	}
	return &proto.Memory{
		Id:           m.ID,
		UserId:       m.UserID,
		Type:         proto.MemoryType(m.Type),
		Category:     m.Category,
		Content:      m.Content,
		Summary:      m.Summary,
		Importance:   m.Importance,
		Priority:     int32(m.Priority),
		Emotion:      m.Emotion,
		EmotionScore: m.EmotionScore,
		Tags:         m.Tags,
		Entities:     m.Entities,
		Source:       m.Source,
		SourceAgent:  m.SourceAgent,
		SourceApp:    m.SourceApp,
		Timestamp:    timestamp,
		LastAccessed: lastAccessed,
		AccessCount:  int32(m.AccessCount),
		Layer:        proto.MemoryLayer(m.Layer),
		DecayFactor:  m.DecayFactor,
		RelatedIds:   m.RelatedIDs,
		ParentId:     m.ParentID,
		CreatedAt:    m.CreatedAt.Unix(),
		UpdatedAt:    m.UpdatedAt.Unix(),
	}
}

// MemoryToModel converts proto.Memory to models.Memory
func MemoryToModel(m *proto.Memory) *models.Memory {
	if m == nil {
		return nil
	}
	return &models.Memory{
		ID:           m.Id,
		UserID:       m.UserId,
		Type:         models.MemoryType(m.Type),
		Category:     m.Category,
		Content:      m.Content,
		Summary:      m.Summary,
		Importance:   m.Importance,
		Priority:     int(m.Priority),
		Emotion:      m.Emotion,
		EmotionScore: m.EmotionScore,
		Tags:         m.Tags,
		Entities:     m.Entities,
		Source:       m.Source,
		SourceAgent:  m.SourceAgent,
		SourceApp:    m.SourceApp,
		Timestamp:    time.Unix(m.Timestamp, 0),
		LastAccessed: time.Unix(m.LastAccessed, 0),
		AccessCount:  int(m.AccessCount),
		Layer:        models.MemoryLayer(m.Layer),
		DecayFactor:  m.DecayFactor,
		RelatedIDs:   m.RelatedIds,
		ParentID:     m.ParentId,
		CreatedAt:    time.Unix(m.CreatedAt, 0),
		UpdatedAt:    time.Unix(m.UpdatedAt, 0),
	}
}

// MemoriesToProto converts models.Memory slice to proto.Memory slice
func MemoriesToProto(memories []*models.Memory) []*proto.Memory {
	result := make([]*proto.Memory, len(memories))
	for i, m := range memories {
		result[i] = MemoryToProto(m)
	}
	return result
}

// PreferenceToProto converts models.Preference to proto.Preference
func PreferenceToProto(p *models.Preference) *proto.Preference {
	if p == nil {
		return nil
	}
	lastUsed := int64(0)
	if !p.LastUsed.IsZero() {
		lastUsed = p.LastUsed.Unix()
	}
	return &proto.Preference{
		Id:           p.ID,
		UserId:       p.UserID,
		Category:     p.Category,
		Key:          p.Key,
		Value:        p.Value,
		ValueType:    p.ValueType,
		Confidence:   p.Confidence,
		Source:       p.Source,
		Context:      p.Context,
		Conditions:   ConditionsToProto(p.Conditions),
		AccessCount:  int32(p.AccessCount),
		ConfirmCount: int32(p.ConfirmCount),
		RejectCount:  int32(p.RejectCount),
		CreatedAt:    p.CreatedAt.Unix(),
		UpdatedAt:    p.UpdatedAt.Unix(),
		LastUsed:     lastUsed,
		Tags:         p.Tags,
		Notes:        p.Notes,
	}
}

// PreferenceToModel converts proto.Preference to models.Preference
func PreferenceToModel(p *proto.Preference) *models.Preference {
	if p == nil {
		return nil
	}
	return &models.Preference{
		ID:           p.Id,
		UserID:       p.UserId,
		Category:     p.Category,
		Key:          p.Key,
		Value:        p.Value,
		ValueType:    p.ValueType,
		Confidence:   p.Confidence,
		Source:       p.Source,
		Context:      p.Context,
		Conditions:   ConditionsToModel(p.Conditions),
		AccessCount:  int(p.AccessCount),
		ConfirmCount: int(p.ConfirmCount),
		RejectCount:  int(p.RejectCount),
		CreatedAt:    time.Unix(p.CreatedAt, 0),
		UpdatedAt:    time.Unix(p.UpdatedAt, 0),
		LastUsed:     time.Unix(p.LastUsed, 0),
		Tags:         p.Tags,
		Notes:        p.Notes,
	}
}

// PreferencesToProto converts models.Preference slice to proto.Preference slice
func PreferencesToProto(prefs []*models.Preference) []*proto.Preference {
	result := make([]*proto.Preference, len(prefs))
	for i, p := range prefs {
		result[i] = PreferenceToProto(p)
	}
	return result
}

// ConditionsToProto converts models.Condition slice to proto.Condition slice
func ConditionsToProto(conditions []models.Condition) []proto.Condition {
	result := make([]proto.Condition, len(conditions))
	for i, c := range conditions {
		result[i] = proto.Condition{
			Type:     c.Type,
			Key:      c.Key,
			Value:    c.Value,
			Operator: c.Operator,
		}
	}
	return result
}

// ConditionsToModel converts proto.Condition slice to models.Condition slice
func ConditionsToModel(conditions []proto.Condition) []models.Condition {
	result := make([]models.Condition, len(conditions))
	for i, c := range conditions {
		result[i] = models.Condition{
			Type:     c.Type,
			Key:      c.Key,
			Value:    c.Value,
			Operator: c.Operator,
		}
	}
	return result
}

// DecisionToProto converts models.Decision to proto.Decision
func DecisionToProto(d *models.Decision) *proto.Decision {
	if d == nil {
		return nil
	}
	options := make([]proto.DecisionOption, len(d.Options))
	for i, o := range d.Options {
		options[i] = *DecisionOptionToProto(&o)
	}

	executedAt := int64(0)
	if d.ExecutedAt != nil {
		executedAt = d.ExecutedAt.Unix()
	}
	completedAt := int64(0)
	if d.CompletedAt != nil {
		completedAt = d.CompletedAt.Unix()
	}

	return &proto.Decision{
		Id:                 d.ID,
		UserId:             d.UserID,
		Scenario:           d.Scenario,
		ScenarioType:       d.ScenarioType,
		Context:            d.Context,
		Options:            options,
		SelectedIndex:      int32(d.SelectedIndex),
		SelectedOption:     DecisionOptionToProto(d.SelectedOption),
		SelectedReason:     d.SelectedReason,
		AppliedValues:      d.AppliedValues,
		AppliedRules:       d.AppliedRules,
		AppliedPreferences: d.AppliedPreferences,
		ScoreDetails:       d.ScoreDetails,
		Ranking:            int32SliceToProto(d.Ranking),
		Outcome:            d.Outcome,
		UserFeedback:       d.UserFeedback,
		OutcomeScore:       d.OutcomeScore,
		ExecutedAt:         executedAt,
		CompletedAt:        completedAt,
		AutoDecided:        d.AutoDecided,
		Confidence:         d.Confidence,
		Tags:               d.Tags,
		CreatedAt:          d.CreatedAt.Unix(),
		UpdatedAt:          d.UpdatedAt.Unix(),
	}
}

// DecisionOptionToProto converts models.DecisionOption to proto.DecisionOption
func DecisionOptionToProto(o *models.DecisionOption) *proto.DecisionOption {
	if o == nil {
		return nil
	}
	return &proto.DecisionOption{
		Id:             o.ID,
		Name:           o.Name,
		Description:    o.Description,
		Attributes:     o.Attributes,
		Score:          o.Score,
		ScoreBreakdown: o.ScoreBreakdown,
		Pros:           o.Pros,
		Cons:           o.Cons,
		Rank:           int32(o.Rank),
	}
}

// int32SliceToProto converts int slice to int32 slice
func int32SliceToProto(s []int) []int32 {
	result := make([]int32, len(s))
	for i, v := range s {
		result[i] = int32(v)
	}
	return result
}

// DecisionResultToProto converts models.DecisionResult to proto.DecisionResult
func DecisionResultToProto(r *models.DecisionResult) *proto.DecisionResult {
	if r == nil {
		return nil
	}
	alternatives := make([]*proto.DecisionOption, len(r.Alternatives))
	for i, a := range r.Alternatives {
		alternatives[i] = DecisionOptionToProto(a)
	}
	return &proto.DecisionResult{
		Decision:        DecisionToProto(r.Decision),
		Alternatives:    alternatives,
		Explanation:     r.Explanation,
		Confidence:      r.Confidence,
		NeedsUserInput:  r.NeedsUserInput,
		UncertainReason: r.UncertainReason,
	}
}

// IdentityVoiceProfileToProto converts models.IdentityVoiceProfile to proto.VoiceProfile
func IdentityVoiceProfileToProto(v *models.IdentityVoiceProfile) *proto.VoiceProfile {
	if v == nil {
		return nil
	}
	return &proto.VoiceProfile{
		Id:               v.ID,
		VoiceType:        v.VoiceType,
		PresetVoiceId:    v.PresetVoiceID,
		CloneReferenceId: v.CloneReferenceID,
		Pitch:            v.Pitch,
		Speed:            v.Speed,
		Volume:           v.Volume,
		Tone:             v.Tone,
		Accent:           v.Accent,
		EmotionLevel:     v.EmotionLevel,
		PausePattern:     v.PausePattern,
		EmphasisStyle:    v.EmphasisStyle,
		CreatedAt:        v.CreatedAt.Unix(),
		UpdatedAt:        v.UpdatedAt.Unix(),
	}
}

// IdentityVoiceProfileToModel converts proto.VoiceProfile to models.IdentityVoiceProfile
func IdentityVoiceProfileToModel(v *proto.VoiceProfile) *models.IdentityVoiceProfile {
	if v == nil {
		return nil
	}
	return &models.IdentityVoiceProfile{
		ID:               v.Id,
		VoiceType:        v.VoiceType,
		PresetVoiceID:    v.PresetVoiceId,
		CloneReferenceID: v.CloneReferenceId,
		Pitch:            v.Pitch,
		Speed:            v.Speed,
		Volume:           v.Volume,
		Tone:             v.Tone,
		Accent:           v.Accent,
		EmotionLevel:     v.EmotionLevel,
		PausePattern:     v.PausePattern,
		EmphasisStyle:    v.EmphasisStyle,
		CreatedAt:        time.Unix(v.CreatedAt, 0),
		UpdatedAt:        time.Unix(v.UpdatedAt, 0),
	}
}

// WritingStyleToProto converts models.WritingStyle to proto.WritingStyle
func WritingStyleToProto(w *models.WritingStyle) *proto.WritingStyle {
	if w == nil {
		return nil
	}
	return &proto.WritingStyle{
		Formality:         w.Formality,
		Verbosity:         w.Verbosity,
		Humor:             w.Humor,
		Technicality:      w.Technicality,
		UseEmoji:          w.UseEmoji,
		UseGifs:           w.UseGIFs,
		UseMarkdown:       w.UseMarkdown,
		SignaturePhrase:   w.SignaturePhrase,
		FrequentWords:     w.FrequentWords,
		AvoidWords:        w.AvoidWords,
		PreferredGreeting: w.PreferredGreeting,
		PreferredClosing:  w.PreferredClosing,
	}
}

// WritingStyleToModel converts proto.WritingStyle to models.WritingStyle
func WritingStyleToModel(w *proto.WritingStyle) *models.WritingStyle {
	if w == nil {
		return nil
	}
	return &models.WritingStyle{
		Formality:         w.Formality,
		Verbosity:         w.Verbosity,
		Humor:             w.Humor,
		Technicality:      w.Technicality,
		UseEmoji:          w.UseEmoji,
		UseGIFs:           w.UseGifs,
		UseMarkdown:       w.UseMarkdown,
		SignaturePhrase:   w.SignaturePhrase,
		FrequentWords:     w.FrequentWords,
		AvoidWords:        w.AvoidWords,
		PreferredGreeting: w.PreferredGreeting,
		PreferredClosing:  w.PreferredClosing,
	}
}