-- ----------------------------------------------------------------------------
-- SAMPLE DATA INSERTION
-- Insert sample feature flags from configuration_flags.yaml
-- ----------------------------------------------------------------------------

-- Insert array-flag
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'array-flag', 
    'team-A', 
    '{
        "variations": {
            "variation_A": ["batmanDefault", "supermanDefault", "superherosDefault"],
            "variation_B": ["batmanFalse", "supermanFalse", "superherosFalse"],
            "variation_C": ["batmanTrue", "supermanTrue", "superherosTrue"]
        },
        "targeting": [
            {
                "name": "rule1",
                "query": "anonymous eq true",
                "percentage": {
                    "variation_A": 0,
                    "variation_B": 90,
                    "variation_C": 10
                }
            }
        ],
        "defaultRule": {
            "variation": "variation_A"
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

