package helper

// XAPIKeyHeader defines the header name for passing API keys.
// This is not a credential itself, but a standard HTTP header key.
const XAPIKeyHeader = "X-API-Key" // nolint: gosec
const ContentTypeHeader = "Content-Type"
const AuthorizationHeader = "Authorization"
const BearerPrefix = "Bearer "
const ContentTypeValueJSON = "application/json"
