# OFREP Response Files

This directory contains JSON response files for OFREP (OpenFeature Remote Evaluation Protocol) flag evaluations.

## File Structure

Each file is named after the flag key and contains the expected response for that flag evaluation.

### Error Responses

- `fail_500.json` - Internal Server Error (500)
- `api_key_missing.json` - API Key Missing (400)
- `invalid_api_key.json` - Invalid API Key (401)
- `flag_not_found.json` - Flag Not Found (404)

### Flag Evaluation Responses

- `bool_targeting_match.json` - Boolean flag with targeting match
- `disabled.json` - Disabled boolean flag
- `disabled_double.json` - Disabled double/float flag
- `disabled_integer.json` - Disabled integer flag
- `disabled_object.json` - Disabled object flag
- `disabled_string.json` - Disabled string flag
- `double_key.json` - Double/float flag with targeting match
- `integer_key.json` - Integer flag with targeting match
- `list_key.json` - List flag with targeting match
- `object_key.json` - Object flag with targeting match
- `string_key.json` - String flag with targeting match
- `unknown_reason.json` - Flag with custom reason
- `does_not_exists.json` - Flag that doesn't exist in configuration
- `integer_with_metadata.json` - Integer flag with metadata

## Usage

The mock automatically loads these files when handling OFREP evaluation requests. If a flag key doesn't have a corresponding file, it returns a "Flag not found" error.
