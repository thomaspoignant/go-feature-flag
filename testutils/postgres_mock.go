package testutils

var PostgresFindResultString = `[{"flag":"test-flag","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},{"flag":"test-flag2","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var PostgresMissingFlagKey = `[{"flag":"test-flag","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var PostgresMissingFlagKeyResult = `{"test-flag":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}}`

var PostgresFindResultFlagNoStr = `[{"flag":123456,"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},{"flag":"test-flag","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var PostgresFlagKeyNotStringResult = MissingFlagKeyResult

var PostgresQueryResult = `{"test-flag":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},"test-flag2":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}}`

var PostgresQueryProperFlagsRelational = `
    BEGIN;

    INSERT INTO feature_flags (id, name, description, "type", bucketing_key, track_events, "disable", version, created_date, last_updated_date, last_modified_by)
    VALUES ('11111111-1111-1111-1111-111111111111', 'Flag_111', 'This is flag number 1', 'boolean', 'bucket1', TRUE, FALSE, 'v1.0', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'admin');

    INSERT INTO variations (id, feature_flag_id, name, value)
    VALUES 
    ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'variation_1', TRUE),
    ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'variation_2', FALSE),
    ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'variation_3', TRUE);

    INSERT INTO rules (id, feature_flag_id, name, "query", "disable", order_index)
    VALUES ('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'Default_Rule_1', 'default_query_1', FALSE, 1);

    COMMIT;

    BEGIN;

    UPDATE feature_flags
    SET default_rule_id = '55555555-5555-5555-5555-555555555555'
    WHERE id = '11111111-1111-1111-1111-111111111111';

    UPDATE rules
    SET variation_result_id = '22222222-2222-2222-2222-222222222222'
    WHERE id = '55555555-5555-5555-5555-555555555555';

    COMMIT;
`

var PostgresQueryProperFlagsRelationalResult = `{"Flag_111":{"bucketing_key":"bucket1","defaultRule":[{"name":"55555555-5555-5555-5555-555555555555","variation":"variation_1"}],"description":"This is flag number 1","disable":false,"id":"11111111-1111-1111-1111-111111111111","targeting":[{"name":"Default_Rule_1","query":"default_query_1","variation":"variation_1"}],"type":"boolean","variations":[{"name":"variation_1","value":"true"},{"name":"variation_2","value":"false"},{"name":"variation_3","value":"true"}]}}`
