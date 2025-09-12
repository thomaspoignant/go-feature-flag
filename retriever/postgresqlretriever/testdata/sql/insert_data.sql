-- ----------------------------------------------------------------------------
-- SAMPLE DATA INSERTION
-- Insert sample feature flags from configuration_flags.yaml
-- ----------------------------------------------------------------------------

-- Insert array-flag
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'array-flag', 
    'default', 
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

-- Insert disable-flag
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'disable-flag', 
    'default', 
    '{
        "variations": {
            "variation_A": "value A",
            "variation_B": "value B",
            "variation_C": "value C"
        },
        "targeting": [
            {
                "name": "rule1",
                "query": "admin eq true",
                "percentage": {
                    "variation_A": 0,
                    "variation_B": 90,
                    "variation_C": 10
                }
            }
        ],
        "defaultRule": {
            "name": "defaultRule",
            "variation": "variation_A"
        },
        "disable": true
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

-- Insert flag-only-for-admin
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'flag-only-for-admin', 
    'default', 
    '{
        "variations": {
            "disabled": false,
            "enabled": true
        },
        "targeting": [
            {
                "name": "rule1",
                "query": "admin eq true",
                "percentage": {
                    "enabled": 0,
                    "disabled": 100
                }
            }
        ],
        "defaultRule": {
            "name": "defaultRule",
            "variation": "disabled"
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

-- Insert new-admin-access
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'new-admin-access', 
    'default', 
    '{
        "variations": {
            "disabled": false,
            "enabled": true
        },
        "defaultRule": {
            "name": "defaultRule",
            "percentage": {
                "enabled": 30,
                "disabled": 70
            }
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

-- Insert number-flag
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'number-flag', 
    'default', 
    '{
        "variations": {
            "variation_A": 1,
            "variation_B": 3,
            "variation_C": 2
        },
        "targeting": [
            {
                "name": "rule1",
                "query": "anonymous eq true",
                "percentage": {
                    "variation_B": 0,
                    "variation_C": 100
                }
            }
        ],
        "defaultRule": {
            "name": "defaultRule",
            "variation": "variation_A"
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

-- Insert targeting-key-rule
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'targeting-key-rule', 
    'default', 
    '{
        "variations": {
            "disabled": false,
            "enabled": true
        },
        "targeting": [
            {
                "query": "targetingKey eq \"specific-targeting-key\"",
                "variation": "enabled"
            }
        ],
        "defaultRule": {
            "variation": "disabled"
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

-- Insert test-flag-rule-apply
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'test-flag-rule-apply', 
    'default', 
    '{
        "variations": {
            "variation_A": {"test": "test"},
            "variation_B": {"test3": "test"},
            "variation_C": {"test2": "test"}
        },
        "targeting": [
            {
                "name": "rule1",
                "query": "key eq \"random-key\"",
                "percentage": {
                    "variation_B": 0,
                    "variation_C": 100
                }
            }
        ],
        "defaultRule": {
            "name": "defaultRule",
            "variation": "variation_A"
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

-- Insert test-flag-rule-apply-false
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'test-flag-rule-apply-false', 
    'default', 
    '{
        "variations": {
            "variation_A": {"test": "test"},
            "variation_B": {"test3": "test"},
            "variation_C": {"test2": "test"}
        },
        "targeting": [
            {
                "name": "rule1",
                "query": "anonymous eq true",
                "percentage": {
                    "variation_B": 90,
                    "variation_C": 10
                }
            }
        ],
        "defaultRule": {
            "name": "defaultRule",
            "variation": "variation_A"
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;

-- Insert test-flag-rule-not-apply
INSERT INTO go_feature_flag (flag_name, flagset, config) VALUES (
    'test-flag-rule-not-apply', 
    'default', 
    '{
        "variations": {
            "variation_A": {"test": "test"},
            "variation_B": {"test3": "test"},
            "variation_C": {"test2": "test"}
        },
        "targeting": [
            {
                "name": "rule1",
                "query": "key eq \"key\"",
                "percentage": {
                    "variation_B": 0,
                    "variation_C": 100
                }
            }
        ],
        "defaultRule": {
            "name": "defaultRule",
            "variation": "variation_A"
        }
    }'::jsonb
) ON CONFLICT (flag_name, flagset) DO NOTHING;
