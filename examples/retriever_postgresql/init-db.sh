#!/bin/bash

set -e

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
EOSQL

