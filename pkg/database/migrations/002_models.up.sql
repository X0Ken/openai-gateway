-- Migration: 002_models
-- Created: 2026-01-31
-- Description: Add models and model_channels tables for many-to-many model-channel mapping

-- Models table: stores logical model names
CREATE TABLE IF NOT EXISTS models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Model-Channel mapping table: links models to channels with backend-specific names
CREATE TABLE IF NOT EXISTS model_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    backend_model_name TEXT NOT NULL,
    weight INTEGER DEFAULT 10,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    UNIQUE(model_id, channel_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_models_name ON models(name);
CREATE INDEX IF NOT EXISTS idx_model_channels_model_id ON model_channels(model_id);
CREATE INDEX IF NOT EXISTS idx_model_channels_channel_id ON model_channels(channel_id);
