-- OFA Center Database Migration Script
-- Version: v7.1.0
-- Description: Complete database schema for Center service

-- === Core Tables ===

-- Agents table (existing)
CREATE TABLE IF NOT EXISTS agents (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    device_info JSONB,
    capabilities JSONB,
    identity_id VARCHAR(64),  -- v7.0.0: bound identity
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tasks table (existing)
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(64) PRIMARY KEY,
    parent_task_id VARCHAR(64),
    skill_id VARCHAR(255),
    input JSONB,
    output JSONB,
    status VARCHAR(32) NOT NULL,
    priority INTEGER DEFAULT 0,
    target_agent VARCHAR(64),
    source_agent VARCHAR(64),
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms BIGINT,
    timeout_ms BIGINT,
    metadata JSONB
);

-- Messages table (existing)
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(64) PRIMARY KEY,
    from_agent VARCHAR(64) NOT NULL,
    to_agent VARCHAR(64) NOT NULL,
    action VARCHAR(255),
    payload JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ttl INTEGER,
    delivered BOOLEAN DEFAULT FALSE,
    headers JSONB
);

-- === Identity Tables (v2.x) ===

-- Personal Identity table
CREATE TABLE IF NOT EXISTS personal_identities (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    nickname VARCHAR(255),
    avatar TEXT,
    birthday TIMESTAMP,
    gender VARCHAR(16),
    location VARCHAR(255),
    occupation VARCHAR(255),
    languages JSONB,
    timezone VARCHAR(64),

    -- Personality (Big Five + MBTI)
    personality JSONB,

    -- Value System
    value_system JSONB,

    -- Voice Profile
    voice_profile JSONB,

    -- Writing Style
    writing_style JSONB,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Interests table
CREATE TABLE IF NOT EXISTS interests (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),
    category VARCHAR(64),
    name VARCHAR(255),
    level FLOAT,
    keywords JSONB,
    description TEXT,
    since TIMESTAMP,
    last_active TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_interests_identity ON interests(identity_id);

-- Behavior Observations table
CREATE TABLE IF NOT EXISTS behavior_observations (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),
    behavior_type VARCHAR(64) NOT NULL,
    description TEXT,
    context JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_behaviors_identity ON behavior_observations(identity_id);
CREATE INDEX IF NOT EXISTS idx_behaviors_type ON behavior_observations(behavior_type);
CREATE INDEX IF NOT EXISTS idx_behaviors_timestamp ON behavior_observations(timestamp);

-- === Device Tables (v2.x) ===

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),
    agent_id VARCHAR(64),  -- Linked agent
    device_type VARCHAR(32),  -- phone, watch, tablet, glasses, etc.
    device_name VARCHAR(255),
    os VARCHAR(32),
    os_version VARCHAR(64),
    model VARCHAR(255),
    manufacturer VARCHAR(255),

    -- Status
    status VARCHAR(32) DEFAULT 'offline',
    last_sync TIMESTAMP,
    last_heartbeat TIMESTAMP,

    -- Capabilities
    capabilities JSONB,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_identity ON devices(identity_id);
CREATE INDEX IF NOT EXISTS idx_devices_agent ON devices(agent_id);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);

-- Device Groups table
CREATE TABLE IF NOT EXISTS device_groups (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),
    name VARCHAR(255),
    description TEXT,
    device_ids JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_device_groups_identity ON device_groups(identity_id);

-- === Emotion Tables (v4.0.0) ===

-- Emotion States table
CREATE TABLE IF NOT EXISTS emotion_states (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),

    -- Emotion values
    joy FLOAT DEFAULT 0.5,
    anger FLOAT DEFAULT 0.0,
    sadness FLOAT DEFAULT 0.0,
    fear FLOAT DEFAULT 0.0,
    love FLOAT DEFAULT 0.0,
    disgust FLOAT DEFAULT 0.0,
    desire FLOAT DEFAULT 0.5,

    -- Current mood
    current_mood VARCHAR(32),
    dominant_emotion VARCHAR(32),

    -- Duration
    duration INTEGER DEFAULT 0,

    -- Timestamps
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_emotion_states_identity ON emotion_states(identity_id);
CREATE INDEX IF NOT EXISTS idx_emotion_states_timestamp ON emotion_states(timestamp);

-- Emotion Triggers table
CREATE TABLE IF NOT EXISTS emotion_triggers (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),
    trigger_type VARCHAR(64),
    trigger_desc TEXT,
    emotion_type VARCHAR(32),
    intensity FLOAT,
    duration INTEGER,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_emotion_triggers_identity ON emotion_triggers(identity_id);

-- Emotion Profiles table
CREATE TABLE IF NOT EXISTS emotion_profiles (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Baselines
    baseline_joy FLOAT DEFAULT 0.5,
    baseline_anger FLOAT DEFAULT 0.0,
    baseline_sadness FLOAT DEFAULT 0.0,
    baseline_fear FLOAT DEFAULT 0.0,
    baseline_love FLOAT DEFAULT 0.3,
    baseline_disgust FLOAT DEFAULT 0.0,

    -- Emotional range
    emotional_range FLOAT DEFAULT 0.5,
    recovery_rate FLOAT DEFAULT 0.1,
    stability_score FLOAT DEFAULT 0.7,

    -- Trigger sensitivity
    trigger_sensitivity JSONB,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Desire States table
CREATE TABLE IF NOT EXISTS desire_states (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),

    -- Maslow hierarchy
    physiological FLOAT DEFAULT 0.7,
    safety FLOAT DEFAULT 0.7,
    love_belonging FLOAT DEFAULT 0.6,
    esteem FLOAT DEFAULT 0.5,
    self_actualization FLOAT DEFAULT 0.4,

    -- Current state
    primary_desire VARCHAR(32),
    satisfaction_level FLOAT,

    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_desire_states_identity ON desire_states(identity_id);

-- === Philosophy Tables (v4.1.0) ===

-- Worldview table
CREATE TABLE IF NOT EXISTS worldviews (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- World essence
    world_essence VARCHAR(64),  -- deterministic, probabilistic, complex_adaptive
    human_nature VARCHAR(64),   -- good, neutral, evil, mixed
    social_cognition VARCHAR(64), -- cooperative, competitive, cooperative_competitive

    -- Control perception
    control_perception VARCHAR(64), -- internal, external, mixed
    change_attitude VARCHAR(64),    -- embrace, accept, resist

    -- Confidence
    confidence_score FLOAT DEFAULT 0.5,

    -- Evidence
    evidence_type VARCHAR(64),  -- experience, observation, reasoning

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Life View table
CREATE TABLE IF NOT EXISTS life_views (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Life meaning
    life_meaning VARCHAR(64),  -- growth_contribution, pleasure_comfort, relationships_connection

    -- Purpose
    purpose_statement TEXT,
    purpose_clarity FLOAT DEFAULT 0.5,

    -- Values priority
    value_priority JSONB,

    -- Time orientation
    time_orientation VARCHAR(64),  -- past_oriented, present_oriented, future_oriented

    -- Death attitude
    death_attitude VARCHAR(64),

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced Value System table
CREATE TABLE IF NOT EXISTS enhanced_value_systems (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Core values
    core_values JSONB,

    -- Moral framework
    moral_framework VARCHAR(64),
    ethical_principles JSONB,

    -- Decision weights
    decision_weights JSONB,

    -- Alignment
    alignment_score FLOAT,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- === Social Identity Tables (v4.2.0) ===

-- Identity Profile table
CREATE TABLE IF NOT EXISTS identity_profiles (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Identity confidence
    identity_confidence FLOAT DEFAULT 0.6,

    -- Self-concept
    self_concept JSONB,

    -- Social comparison
    comparison_tendency FLOAT DEFAULT 0.3,

    -- Identity stability
    stability_score FLOAT DEFAULT 0.7,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Education Profile table
CREATE TABLE IF NOT EXISTS education_profiles (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Education level
    education_level VARCHAR(64),
    school_tier VARCHAR(64),
    major VARCHAR(255),
    school_name VARCHAR(255),

    -- Academic performance
    academic_performance FLOAT,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Career Profile table
CREATE TABLE IF NOT EXISTS career_profiles (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Career basics
    occupation VARCHAR(255),
    industry VARCHAR(64),
    company VARCHAR(255),
    stage VARCHAR(64),
    work_mode VARCHAR(64),  -- office, remote, hybrid, field

    -- Satisfaction
    satisfaction FLOAT DEFAULT 0.5,
    ambition_level FLOAT DEFAULT 0.6,

    -- Income tier
    income_tier VARCHAR(64),

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Social Class Profile table
CREATE TABLE IF NOT EXISTS social_class_profiles (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Capitals
    economic_capital JSONB,
    cultural_capital JSONB,
    social_capital JSONB,

    -- Class perception
    class_identity VARCHAR(64),
    class_mobility FLOAT,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- === Culture Tables (v4.3.0) ===

-- Regional Culture table
CREATE TABLE IF NOT EXISTS regional_cultures (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Location
    country VARCHAR(64),
    province VARCHAR(64),
    city VARCHAR(255),
    city_level VARCHAR(64),  -- first_tier, new_first_tier, second_tier, etc.

    -- Duration
    residence_duration INTEGER,
    migration_history JSONB,

    -- Hofstede dimensions
    hofstede_dimensions JSONB,

    -- Communication style
    communication_style VARCHAR(64),  -- direct, indirect, context_dependent

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- === Life Stage Tables (v4.4.0) ===

-- Life Stage table
CREATE TABLE IF NOT EXISTS life_stages (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Current stage
    current_stage VARCHAR(64),  -- childhood, adolescence, youth, early_adult, etc.
    stage_start TIMESTAMP,
    stage_progress FLOAT,

    -- Age
    chronological_age INTEGER,
    psychological_age INTEGER,

    -- Stage characteristics
    stage_characteristics JSONB,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Life Events table
CREATE TABLE IF NOT EXISTS life_events (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),

    -- Event details
    event_type VARCHAR(64),
    event_desc TEXT,
    event_year INTEGER,

    -- Impact
    impact FLOAT,
    emotion_impact JSONB,

    -- Status
    processed BOOLEAN DEFAULT FALSE,

    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_life_events_identity ON life_events(identity_id);

-- === Relationship Tables (v4.6.0) ===

-- Relationships table
CREATE TABLE IF NOT EXISTS relationships (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),

    -- Person
    person_id VARCHAR(64),
    person_name VARCHAR(255),

    -- Relationship type
    relationship_type VARCHAR(64),  -- family, friend, colleague, romantic, etc.

    -- Scores
    intimacy FLOAT DEFAULT 0.5,
    trust FLOAT DEFAULT 0.5,
    importance FLOAT DEFAULT 0.5,
    compatibility FLOAT DEFAULT 0.5,

    -- Interaction
    interaction_frequency VARCHAR(64),
    last_interaction TIMESTAMP,

    -- Metadata
    metadata JSONB,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_relationships_identity ON relationships(identity_id);
CREATE INDEX IF NOT EXISTS idx_relationships_person ON relationships(person_id);

-- Relationship Profile table
CREATE TABLE IF NOT EXISTS relationship_profiles (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Attachment style
    attachment_style VARCHAR(64),  -- secure, anxious, avoidant, disorganized

    -- Social style
    social_style VARCHAR(64),
    conflict_style VARCHAR(64),

    -- Network
    network_size INTEGER DEFAULT 10,
    network_density FLOAT DEFAULT 0.3,

    -- Social capital
    social_capital FLOAT DEFAULT 0.5,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- === Sync Tables (v2.x) ===

-- Sync States table
CREATE TABLE IF NOT EXISTS sync_states (
    identity_id VARCHAR(64) PRIMARY KEY REFERENCES personal_identities(id),

    -- Versions
    identity_version BIGINT DEFAULT 0,
    memory_version BIGINT DEFAULT 0,
    preference_version BIGINT DEFAULT 0,
    emotion_version BIGINT DEFAULT 0,

    -- Device count
    device_count INTEGER DEFAULT 0,

    -- Last sync
    last_full_sync TIMESTAMP,
    last_partial_sync TIMESTAMP,

    -- Conflict tracking
    pending_conflicts INTEGER DEFAULT 0,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sync History table
CREATE TABLE IF NOT EXISTS sync_history (
    id VARCHAR(64) PRIMARY KEY,
    identity_id VARCHAR(64) NOT NULL REFERENCES personal_identities(id),
    agent_id VARCHAR(64),

    -- Sync details
    sync_type VARCHAR(64),  -- full, partial, conflict_resolution
    data_types JSONB,

    -- Status
    status VARCHAR(32),  -- pending, completed, failed, conflict
    conflict_count INTEGER DEFAULT 0,

    -- Timing
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_ms BIGINT,

    -- Result
    result JSONB,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sync_history_identity ON sync_history(identity_id);
CREATE INDEX IF NOT EXISTS idx_sync_history_agent ON sync_history(agent_id);

-- === WebSocket Sessions (v7.0.0) ===

-- WebSocket Sessions table
CREATE TABLE IF NOT EXISTS ws_sessions (
    session_id VARCHAR(64) PRIMARY KEY,
    agent_id VARCHAR(64) NOT NULL,
    identity_id VARCHAR(64) REFERENCES personal_identities(id),

    -- Connection info
    connection_status VARCHAR(32),
    connected_at TIMESTAMP,
    disconnected_at TIMESTAMP,

    -- Stats
    messages_sent INTEGER DEFAULT 0,
    messages_received INTEGER DEFAULT 0,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ws_sessions_agent ON ws_sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_ws_sessions_identity ON ws_sessions(identity_id);

-- === Indexes ===

-- Additional indexes for performance
CREATE INDEX IF NOT EXISTS idx_agents_identity ON agents(identity_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_target ON tasks(target_agent);
CREATE INDEX IF NOT EXISTS idx_messages_to ON messages(to_agent);
CREATE INDEX IF NOT EXISTS idx_messages_delivered ON messages(delivered);