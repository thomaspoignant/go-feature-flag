.PHONY: build

build:
	sam build

start-api:
	sam local start-api

call-api:
	curl -X 'POST' \
      'http://localhost:3000/v1/feature/flag-only-for-admin/eval' \
      -H 'accept: application/json' \
      -H 'Content-Type: application/json' \
      -d '{ \
      "defaultValue": "string",\
      "user": {\
        "anonymous": false,\
        "custom": {\
          "company": "GO Feature Flag",\
          "email": "contact@gofeatureflag.org",\
          "admin": true \
        },\
        "key": "08b5ffb7-7109-42f4-a6f2-b85560fbd20f"\
      }\
    }'
