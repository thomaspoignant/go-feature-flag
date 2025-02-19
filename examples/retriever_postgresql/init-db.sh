#!/bin/bash

set -e

num_flags=5
count=1;


generate_uuid() {
  echo $(cat /proc/sys/kernel/random/uuid)
}


psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  CREATE TABLE IF NOT EXISTS flags (
    id SERIAL PRIMARY KEY,
    flag JSONB
  );

  INSERT INTO flags (flag) VALUES
  (
    '{
      "flag": "new-admin-access",
      "variations": {
          "default_var": false,
          "false_var": false,
          "true_var": true
      },
      "defaultRule": {
          "percentage": {
              "false_var": 70,
              "true_var": 30
          }
      }
    }'::jsonb
  );

  INSERT INTO flags (flag) VALUES
  (
    '{
      "flag": "flag-only-for-admin",
      "variations": {
          "default_var": false,
          "false_var": false,
          "true_var": true
      },
      "targeting": [
          {
              "query": "admin eq true",
              "percentage": {
                  "false_var": 0,
                  "true_var": 100
              }
          }
      ],
      "defaultRule": {
          "variation": "default_var"
      }
    }'::jsonb
  );

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


 
EOSQL

# Insert flags, rules, variations, percentages, and progressive rollouts
for i in $(seq 1 $num_flags); do
  feature_flag_uuid=$(generate_uuid)
  flag_name="Flag_${i}"
  bucketing_key="bucket${i}"
  version="v${i}.0"
  description="This is flag number ${i}"
  printf "number: %d\n" "$i"

  # Insert the flag with default_rule_id (set initially to NULL for the moment)
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    INSERT INTO feature_flags 
      (id, name, description, type, bucketing_key, default_rule_id, track_events, disable, version, created_date, last_updated_date, last_modified_by)
    VALUES 
      ('$feature_flag_uuid', '$flag_name', '$description', 'boolean', '$bucketing_key', NULL, true, false, '$version', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'admin');
EOSQL

  # Insert variations for each flag
  for variation_num in $(seq 1 3); do
    variation_uuid=$(generate_uuid)
    variation_name="variation_${variation_num}"
    variation_value=$((RANDOM % 2))  # Randomly assign true (1) or false (0)

    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
      INSERT INTO variations (id, feature_flag_id, name, value)
      VALUES ('$variation_uuid', '$feature_flag_uuid', '$variation_name', '$variation_value');
EOSQL
  done

  # Create the default rule and set it in the flag (this rule is not associated with feature_flag_id)
  default_rule_uuid=$(generate_uuid)
  default_rule_name="Default_Rule_${i}"
  default_rule_query="default_query_${i}"

  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    INSERT INTO rules 
      (id, feature_flag_id, name, query, disable, order_index)
    VALUES 
      ('$default_rule_uuid', NULL, '$default_rule_name', '$default_rule_query', true, 1);
EOSQL

  # Update the flag with the default_rule_id
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    UPDATE feature_flags 
    SET default_rule_id = '$default_rule_uuid'
    WHERE id = '$feature_flag_uuid';
EOSQL

  # Create between 1 and 5 rules for each flag (these will have feature_flag_id)
  num_rules=$((RANDOM % 5 + 1))  # Random number of rules (between 1 and 5)
  for j in $(seq 1 $num_rules); do
    rule_uuid=$(generate_uuid)
    rule_name="Rule_${i}_${j}"
    rule_query="query${i}_${j}"

    # Insert the rule with feature_flag_id (this will link the rule to the flag)
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
      INSERT INTO rules 
        (id, feature_flag_id, name, query, disable, order_index)
      VALUES 
        ('$rule_uuid', '$feature_flag_uuid', '$rule_name', '$rule_query', false, $j);
EOSQL

    # Randomly choose what the rule will have (variation_result_id, percentages, or progressive_rollouts)
    rule_type=$((RANDOM % 3))  # Randomly decide if the rule has percentages, progressive rollouts, or variation_result_id

    # If the rule should have percentage-based variations (rule_type == 0)
    if [[ $rule_type -eq 0 ]]; then
      # Use the variations created for the feature flag
      for variation_id in $(psql -t -A -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "SELECT id FROM variations WHERE feature_flag_id = '$feature_flag_uuid';"); do
        percentage=$((RANDOM % 100 + 1))  # Random percentage between 1-100

        psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
          INSERT INTO rule_percentages 
            (id, rule_id, variation_id, percentage)
          VALUES 
            ('$(generate_uuid)', '$rule_uuid', '$variation_id', $percentage);
EOSQL
      done

    # If the rule should have progressive rollouts (rule_type == 1)
    elif [[ $rule_type -eq 1 ]]; then
      initial_variation_uuid=$(psql -t -A -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "SELECT id FROM variations WHERE feature_flag_id = '$feature_flag_uuid' ORDER BY RANDOM() LIMIT 1;")
      end_variation_uuid=$(psql -t -A -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "SELECT id FROM variations WHERE feature_flag_id = '$feature_flag_uuid' ORDER BY RANDOM() LIMIT 1;")

      initial_percentage=$((RANDOM % 100 + 1))
      end_percentage=$((100 - initial_percentage))

      progressive_rollout_uuid=$(generate_uuid)
      
      psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        INSERT INTO progressive_rollouts 
          (id, rule_id, initial_variation_id, initial_percentage, initial_date, end_variation_id, end_percentage, end_date)
        VALUES 
          ('$progressive_rollout_uuid', '$rule_uuid', '$initial_variation_uuid', $initial_percentage, CURRENT_TIMESTAMP, '$end_variation_uuid', $end_percentage, CURRENT_TIMESTAMP + INTERVAL '1 day');
EOSQL

    # If the rule should have a variation_result_id (rule_type == 2)
    elif [[ $rule_type -eq 2 ]]; then
      # Select a random variation from the variations associated with this feature flag
      variation_result_uuid=$(psql -t -A -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "SELECT id FROM variations WHERE feature_flag_id = '$feature_flag_uuid' ORDER BY RANDOM() LIMIT 1;")

      # Update the rule with the randomly selected variation_result_id
      psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        UPDATE rules
        SET variation_result_id = '$variation_result_uuid'
        WHERE id = '$rule_uuid';
EOSQL
    fi
  done
done

