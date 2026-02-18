-- 000001_initial_schema.up.sql
-- Initial schema for DND MCP Server

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Campaigns table
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    dm_id VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_campaigns_dm_id ON campaigns(dm_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);

-- Characters table
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_npc BOOLEAN DEFAULT FALSE,
    npc_type VARCHAR(50),
    player_id VARCHAR(255),

    -- Basic attributes
    race VARCHAR(100) NOT NULL,
    class VARCHAR(100) NOT NULL,
    level INT DEFAULT 1,
    background VARCHAR(255),
    alignment VARCHAR(50),

    -- Ability scores and combat attributes
    abilities JSONB NOT NULL,
    hp JSONB NOT NULL,
    ac INT NOT NULL,
    speed INT DEFAULT 30,
    initiative INT DEFAULT 0,

    -- Skills and saves
    skills JSONB DEFAULT '{}',
    saves JSONB DEFAULT '{}',

    -- Equipment and inventory
    equipment JSONB DEFAULT '[]',
    inventory JSONB DEFAULT '[]',

    -- Status
    conditions JSONB DEFAULT '[]',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_characters_campaign_id ON characters(campaign_id);
CREATE INDEX idx_characters_is_npc ON characters(is_npc);
CREATE INDEX idx_characters_player_id ON characters(player_id);

-- Combats table
CREATE TABLE combats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'active',
    round INT DEFAULT 1,
    turn_index INT DEFAULT 0,
    participants JSONB NOT NULL,
    map_id UUID,
    log JSONB DEFAULT '[]',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_combats_campaign_id ON combats(campaign_id);
CREATE INDEX idx_combats_status ON combats(status);

-- Maps table
CREATE TABLE maps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    grid JSONB NOT NULL,
    locations JSONB DEFAULT '[]',
    tokens JSONB DEFAULT '[]',
    parent_id UUID REFERENCES maps(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_maps_campaign_id ON maps(campaign_id);
CREATE INDEX idx_maps_type ON maps(type);
CREATE INDEX idx_maps_parent_id ON maps(parent_id);

-- Game states table
CREATE TABLE game_states (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE UNIQUE,
    game_time JSONB NOT NULL,
    party_position JSONB,
    current_map_id UUID REFERENCES maps(id),
    current_map_type VARCHAR(50) DEFAULT 'world',
    weather VARCHAR(50),
    active_combat_id UUID REFERENCES combats(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_game_states_campaign_id ON game_states(campaign_id);

-- Messages table
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    player_id VARCHAR(255),
    tool_calls JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_messages_campaign_id ON messages(campaign_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_maps_updated_at BEFORE UPDATE ON maps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_game_states_updated_at BEFORE UPDATE ON game_states
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
