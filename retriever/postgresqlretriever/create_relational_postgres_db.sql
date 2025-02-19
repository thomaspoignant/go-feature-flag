  CREATE TYPE variationType AS ENUM ('string','boolean','integer','double','json');
  
  CREATE TABLE IF NOT EXISTS feature_flags
  (
    id                UUID NOT NULL PRIMARY KEY,
    name              TEXT NOT NULL UNIQUE CHECK (name != ''),
    description       TEXT,
    type              variationType NOT NULL,
    bucketing_key     TEXT,
    default_rule_id   UUID,
    track_events      BOOLEAN DEFAULT TRUE,
    disable           BOOLEAN DEFAULT FALSE,
    version           TEXT,
    created_date      TIMESTAMP NOT NULL,
    last_updated_date TIMESTAMP NOT NULL,
    last_modified_by  TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS metadata
  (
    id                  UUID PRIMARY KEY,
    feature_flag_id     UUID REFERENCES feature_flags (id) NOT NULL,
    name                TEXT NOT NULL,
    value               TEXT NOT NULL
  );
  
  CREATE TABLE IF NOT EXISTS variations 
  (
    id UUID PRIMARY KEY,
    feature_flag_id UUID REFERENCES feature_flags (id) ON DELETE CASCADE NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS rules
  (
    id                   UUID NOT NULL PRIMARY KEY,
    feature_flag_id      UUID REFERENCES feature_flags (id) ON DELETE SET NULL,
    name                 TEXT,
    query                TEXT,
    variation_result_id  UUID REFERENCES variations (id) ON DELETE SET NULL,
    disable              BOOLEAN DEFAULT FALSE NOT NULL,
    order_index          INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS progressive_rollouts
  (
    id                        UUID PRIMARY KEY,
    rule_id                   UUID REFERENCES rules (id) ON DELETE CASCADE NOT NULL,
    initial_variation_id      UUID REFERENCES variations (id) ON DELETE CASCADE,
    initial_percentage        FLOAT CHECK (initial_percentage >= 0 AND initial_percentage <= 100),
    initial_date              TIMESTAMP,
    end_variation_id          UUID REFERENCES variations (id) ON DELETE CASCADE,
    end_percentage            FLOAT CHECK (end_percentage >= 0 AND end_percentage <= 100),
    end_date                  TIMESTAMP
  );
  
  CREATE TABLE rule_percentages 
  (
    id UUID PRIMARY KEY,
    rule_id UUID REFERENCES rules (id) ON DELETE CASCADE NOT NULL,
    variation_id UUID REFERENCES variations (id) ON DELETE CASCADE NOT NULL,
    percentage INTEGER CHECK (percentage >= 0 AND percentage <= 100)
  );


  CREATE INDEX idx_feature_flags_name ON feature_flags (name);
  CREATE INDEX idx_rules_feature_flag_id ON rules (feature_flag_id);
  CREATE INDEX idx_variations_feature_flag_id ON variations (feature_flag_id);
  CREATE INDEX idx_percentages_rule_id ON rule_percentages (rule_id);
  CREATE INDEX idx_metadata_feature_flag_id ON metadata (feature_flag_id);
  CREATE INDEX idx_progressive_rollouts_rule_id ON progressive_rollouts (rule_id);

  CREATE INDEX idx_feature_flags_default_rule_id ON feature_flags (default_rule_id);
  CREATE INDEX idx_rules_variation_result_id ON rules (variation_result_id);
  CREATE INDEX idx_progressive_rollouts_initial_variation_id ON progressive_rollouts (initial_variation_id);
  CREATE INDEX idx_progressive_rollouts_end_variation_id ON progressive_rollouts (end_variation_id);
  CREATE INDEX idx_rule_percentages_variation_id ON rule_percentages (variation_id);

ALTER TABLE feature_flags
ADD CONSTRAINT fk_feature_flags_default_rule
FOREIGN KEY (default_rule_id) REFERENCES rules (id)
ON DELETE SET NULL;
