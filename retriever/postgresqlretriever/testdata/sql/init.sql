-- ----------------------------------------------------------------------------
-- GO FEATURE FLAG POSTGRESQL RETRIEVER INITIALIZATION SCRIPT
-- This script is used to initialize the PostgreSQL database for the Go Feature Flag retriever.
-- This is a very minimal setup for the retriever, you can add more columns or constraints to the table if you need to.
-- ----------------------------------------------------------------------------

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the go_feature_flag table
CREATE TABLE IF NOT EXISTS go_feature_flag (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_name VARCHAR(255) NOT NULL,
    flagset VARCHAR(255) NOT NULL,
    config JSONB NOT NULL
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_go_feature_flag_flagset ON go_feature_flag(flagset);
CREATE INDEX IF NOT EXISTS idx_go_feature_flag_flag_name ON go_feature_flag(flag_name);
CREATE INDEX IF NOT EXISTS idx_go_feature_flag_flagset_flag_name ON go_feature_flag(flagset, flag_name);

-- Add a unique constraint to prevent duplicate flags in the same flagset
ALTER TABLE go_feature_flag ADD CONSTRAINT unique_flag_per_flagset UNIQUE (flag_name, flagset);