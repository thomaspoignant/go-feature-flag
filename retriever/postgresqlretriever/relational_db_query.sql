WITH rule_percentages_agg AS (
    SELECT 
    rp.rule_id,
    jsonb_agg(jsonb_build_object(rv.name, rp.percentage)) AS aggregated_percentages
    FROM rule_percentages rp
    LEFT JOIN variations rv ON rp.variation_id = rv.id
    GROUP BY rp.rule_id
),
variations_agg AS (
    SELECT 
    ff.id AS feature_flag_id,
    jsonb_agg(
        jsonb_build_object(
            'name', v.name, 
            'value', 
            CASE 
                    WHEN ff.type = 'string' THEN v.value::text
                        -- todo add values casting
                    ELSE v.value
                END
            )
        ) AS variations
        FROM feature_flags ff
        LEFT JOIN variations v ON ff.id = v.feature_flag_id
        GROUP BY ff.id
    ),
    default_rule_agg AS (
        SELECT 
        ff.id AS feature_flag_id,
        jsonb_agg(
            jsonb_build_object('name', dr.id) ||
            CASE 
                WHEN drv IS NOT NULL THEN jsonb_build_object('variation', drv.name)
                WHEN pr IS NOT NULL THEN jsonb_build_object(
                    'initial', jsonb_build_object(
                        'variation', prvq.name, 'percentage', pr.initial_percentage, 'date', pr.initial_date
                    ),
                    'end', jsonb_build_object(
                        'variation', prvw.name, 'percentage', pr.end_percentage, 'date', pr.end_date
                    )
                )
                WHEN rpa.aggregated_percentages IS NOT NULL THEN jsonb_build_object(
                    'percentages', rpa.aggregated_percentages
                )
            ELSE '{}'::jsonb
            END
        ) AS default_rule
        FROM feature_flags ff
        LEFT JOIN rules dr ON ff.default_rule_id = dr.id
        LEFT JOIN variations drv ON dr.variation_result_id = drv.id
        LEFT JOIN progressive_rollouts pr ON dr.id = pr.rule_id
        LEFT JOIN variations prvq ON pr.initial_variation_id = prvq.id
        LEFT JOIN variations prvw ON pr.end_variation_id = prvw.id
        LEFT JOIN rule_percentages_agg rpa ON dr.id = rpa.rule_id
        GROUP BY ff.id
    ),
    targeting_agg AS (
        SELECT 
        ff.id AS feature_flag_id,
        jsonb_agg(
            jsonb_build_object('name', r.name, 'query', r.query) ||
            CASE 
                WHEN rvv IS NOT NULL THEN jsonb_build_object('variation', rvv.name)
                WHEN pr IS NOT NULL THEN jsonb_build_object(
                    'initial', jsonb_build_object(
                        'variation', prvq.name, 'percentage', pr.initial_percentage, 'date', pr.initial_date
                    ),
                    'end', jsonb_build_object(
                        'variation', prvw.name, 'percentage', pr.end_percentage, 'date', pr.end_date
                    )
                )
                WHEN rpa.aggregated_percentages IS NOT NULL THEN jsonb_build_object(
                    'percentages', rpa.aggregated_percentages
                )
            ELSE '{}'::jsonb
            END
        ) AS targeting
        FROM feature_flags ff
        LEFT JOIN rules r ON ff.id = r.feature_flag_id
        LEFT JOIN variations rvv ON r.variation_result_id = rvv.id
        LEFT JOIN progressive_rollouts pr ON r.id = pr.rule_id
        LEFT JOIN variations prvq ON pr.initial_variation_id = prvq.id
        LEFT JOIN variations prvw ON pr.end_variation_id = prvw.id
        LEFT JOIN rule_percentages_agg rpa ON r.id = rpa.rule_id 
        GROUP BY ff.id
    )
    SELECT jsonb_build_object(
        'id', ff.id,
        'flag', ff.name,
        'type', ff.type,
        'description', ff.description,
        'bucketing_key', ff.bucketing_key,
        'disable', ff.disable,
        'variations', va.variations,
        'defaultRule', dra.default_rule,
        'targeting', ta.targeting
    ) AS flag
    FROM feature_flags ff
    LEFT JOIN variations_agg va ON ff.id = va.feature_flag_id
    LEFT JOIN default_rule_agg dra ON ff.id = dra.feature_flag_id
    LEFT JOIN targeting_agg ta ON ff.id = ta.feature_flag_id
