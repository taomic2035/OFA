package grpc

import (
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
		MBTIType:         p.MBTIType,
		MBTI_EI:          p.MBTI_EI,
		MBTI_SN:          p.MBTI_SN,
		MBTI_TF:          p.MBTI_TF,
		MBTI_JP:          p.MBTI_JP,
		MBTIConfidence:   p.MBTIConfidence,
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
		MBTIType:         p.MBTIType,
		MBTI_EI:          p.MBTI_EI,
		MBTI_SN:          p.MBTI_SN,
		MBTI_TF:          p.MBTI_TF,
		MBTI_JP:          p.MBTI_JP,
		MBTIConfidence:   p.MBTIConfidence,
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
		result[i] = &proto.Interest{
			Id:         interest.ID,
			Category:   interest.Category,
			Name:       interest.Name,
			Keywords:   interest.Keywords,
			Level:      interest.Level,
			CreatedAt:  interest.CreatedAt,
			UpdatedAt:  interest.UpdatedAt,
			LastActive: interest.LastActive,
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
			CreatedAt:  interest.CreatedAt,
			UpdatedAt:  interest.UpdatedAt,
			LastActive: interest.LastActive,
		}
	}
	return result
}

// IdentityToProto converts models.PersonalIdentity to proto.Identity
func IdentityToProto(id *models.PersonalIdentity) *proto.Identity {
	if id == nil {
		return nil
	}
	return &proto.Identity{
		UserId:          id.ID,
		Name:            id.Name,
		Nickname:        id.Nickname,
		Avatar:          id.Avatar,
		Gender:          id.Gender,
		Birthday:        id.Birthday,
		Location:        id.Location,
		Occupation:      id.Occupation,
		Languages:       id.Languages,
		Timezone:        id.Timezone,
		Personality:     PersonalityToProto(id.Personality),
		ValueSystem:     ValueSystemToProto(id.ValueSystem),
		Interests:       InterestsToProto(id.Interests),
		SpeakingTone:    id.Personality.SpeakingTone,
		ResponseLength:  id.Personality.ResponseLength,
		ValuePriority:   id.GetValuePriority(),
		CreatedAt:       id.CreatedAt,
		UpdatedAt:       id.UpdatedAt,
	}
}

// MemoryToProto converts models.Memory to proto.Memory
func MemoryToProto(m *models.Memory) *proto.Memory {
	if m == nil {
		return nil
	}
	return &proto.Memory{
		Id:          m.ID,
		UserId:      m.UserID,
		Type:        proto.MemoryType(m.Type),
		Content:     m.Content,
		Importance:  m.Importance,
		Layer:       proto.MemoryLayer(m.Layer),
		DecayFactor: m.DecayFactor,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		AccessCount: int32(m.AccessCount),
		ExpiresAt:   m.ExpiresAt,
		SourceId:    m.SourceID,
		Tags:        m.Tags,
		Metadata:    m.Metadata,
	}
}

// MemoryToModel converts proto.Memory to models.Memory
func MemoryToModel(m *proto.Memory) *models.Memory {
	if m == nil {
		return nil
	}
	return &models.Memory{
		ID:          m.Id,
		UserID:      m.UserId,
		Type:        models.MemoryType(m.Type),
		Content:     m.Content,
		Importance:  m.Importance,
		Layer:       models.MemoryLayer(m.Layer),
		DecayFactor: m.DecayFactor,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		AccessCount: int(m.AccessCount),
		ExpiresAt:   m.ExpiresAt,
		SourceID:    m.SourceId,
		Tags:        m.Tags,
		Metadata:    m.Metadata,
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
	return &proto.Preference{
		Id:          p.ID,
		UserId:      p.UserID,
		Key:         p.Key,
		Value:       p.Value,
		Confidence:  p.Confidence,
		Source:      p.Source,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		AccessCount: int32(p.AccessCount),
		ExpiresAt:   p.ExpiresAt,
	}
}

// PreferenceToModel converts proto.Preference to models.Preference
func PreferenceToModel(p *proto.Preference) *models.Preference {
	if p == nil {
		return nil
	}
	return &models.Preference{
		ID:          p.Id,
		UserID:      p.UserId,
		Key:         p.Key,
		Value:       p.Value,
		Confidence:  p.Confidence,
		Source:      p.Source,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		AccessCount: int(p.AccessCount),
		ExpiresAt:   p.ExpiresAt,
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

// DecisionToProto converts models.Decision to proto.Decision
func DecisionToProto(d *models.Decision) *proto.Decision {
	if d == nil {
		return nil
	}
	options := make([]*proto.DecisionOption, len(d.Options))
	for i, o := range d.Options {
		options[i] = DecisionOptionToProto(&o)
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
		ExecutedAt:         d.ExecutedAt,
		CompletedAt:        d.CompletedAt,
		AutoDecided:        d.AutoDecided,
		Confidence:         d.Confidence,
		Tags:               d.Tags,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
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

// VoiceProfileToProto converts models.VoiceProfile to proto.VoiceProfile
func VoiceProfileToProto(v *models.VoiceProfile) *proto.VoiceProfile {
	if v == nil {
		return nil
	}
	customVoices := make([]*proto.CustomVoice, len(v.CustomVoices))
	for i, cv := range v.CustomVoices {
		customVoices[i] = &proto.CustomVoice{
			Id:          cv.ID,
			Name:        cv.Name,
			VoiceId:     cv.VoiceID,
			CreatedAt:   cv.CreatedAt,
			SampleCount: int32(cv.SampleCount),
			Quality:     cv.Quality,
		}
	}

	return &proto.VoiceProfile{
		UserId:            v.UserID,
		DefaultTtsVoice:   v.DefaultTTSVoice,
		SpeechRate:        v.SpeechRate,
		SpeechPitch:       v.SpeechPitch,
		Volume:            v.Volume,
		PreferredLanguage: v.PreferredLanguage,
		CustomVoices:      customVoices,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}

// VoiceProfileToModel converts proto.VoiceProfile to models.VoiceProfile
func VoiceProfileToModel(v *proto.VoiceProfile) *models.VoiceProfile {
	if v == nil {
		return nil
	}
	customVoices := make([]models.CustomVoice, len(v.CustomVoices))
	for i, cv := range v.CustomVoices {
		customVoices[i] = models.CustomVoice{
			ID:          cv.Id,
			Name:        cv.Name,
			VoiceID:     cv.VoiceId,
			CreatedAt:   cv.CreatedAt,
			SampleCount: int(cv.SampleCount),
			Quality:     cv.Quality,
		}
	}

	return &models.VoiceProfile{
		UserID:            v.UserId,
		DefaultTTSVoice:   v.DefaultTtsVoice,
		SpeechRate:        v.SpeechRate,
		SpeechPitch:       v.SpeechPitch,
		Volume:            v.Volume,
		PreferredLanguage: v.PreferredLanguage,
		CustomVoices:      customVoices,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}