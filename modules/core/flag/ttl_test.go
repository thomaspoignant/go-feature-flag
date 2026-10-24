package flag_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func TestParseTTL(t *testing.T) {
	tests := []struct {
		name    string
		ttl     string
		want    time.Duration
		wantErr bool
	}{
		{name: "days", ttl: "90d", want: 90 * 24 * time.Hour},
		{name: "go duration", ttl: "1h30m", want: 90 * time.Minute},
		{name: "spaces", ttl: " 2d ", want: 48 * time.Hour},
		{name: "empty", ttl: "", wantErr: true},
		{name: "garbage", ttl: "abc", wantErr: true},
		{name: "negative days", ttl: "-1d", wantErr: true},
		{name: "zero", ttl: "0h", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flag.ParseTTL(tt.ttl)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInternalFlag_IsExpired(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ttl := "10d"
	badTTL := "nope"
	expiresAt := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	defaultTTL := 30 * 24 * time.Hour

	tests := []struct {
		name       string
		flag       flag.InternalFlag
		now        time.Time
		defaultTTL time.Duration
		want       bool
	}{
		{name: "no createdAt never expires", flag: flag.InternalFlag{}, now: created.AddDate(10, 0, 0), defaultTTL: defaultTTL},
		{name: "default ttl not reached", flag: flag.InternalFlag{CreatedAt: &created}, now: created.AddDate(0, 0, 29), defaultTTL: defaultTTL},
		{name: "default ttl reached", flag: flag.InternalFlag{CreatedAt: &created}, now: created.AddDate(0, 0, 30), defaultTTL: defaultTTL, want: true},
		{name: "flag ttl overrides default", flag: flag.InternalFlag{CreatedAt: &created, TTL: &ttl}, now: created.AddDate(0, 0, 10), defaultTTL: defaultTTL, want: true},
		{name: "expiresAt overrides ttl", flag: flag.InternalFlag{CreatedAt: &created, TTL: &ttl, ExpiresAt: &expiresAt}, now: expiresAt, defaultTTL: defaultTTL, want: true},
		{name: "expiresAt without createdAt", flag: flag.InternalFlag{ExpiresAt: &expiresAt}, now: expiresAt.Add(-time.Second)},
		{name: "invalid ttl never expires", flag: flag.InternalFlag{CreatedAt: &created, TTL: &badTTL}, now: created.AddDate(1, 0, 0), defaultTTL: defaultTTL},
		{name: "disabled default", flag: flag.InternalFlag{CreatedAt: &created}, now: created.AddDate(1, 0, 0), defaultTTL: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.flag.IsExpired(tt.now, tt.defaultTTL))
		})
	}
}

func TestInternalFlag_ExpirationDate_IsUTC(t *testing.T) {
	paris := time.FixedZone("CEST", 2*60*60)
	created := time.Date(2026, 1, 1, 2, 0, 0, 0, paris)
	expiresAt := time.Date(2026, 3, 1, 12, 0, 0, 0, paris)
	ttl := "1d"

	got, ok := (&flag.InternalFlag{CreatedAt: &created, TTL: &ttl}).ExpirationDate(0)
	assert.True(t, ok)
	assert.Equal(t, time.UTC, got.Location())
	assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), got.Add(-24*time.Hour))

	got, ok = (&flag.InternalFlag{ExpiresAt: &expiresAt}).ExpirationDate(0)
	assert.True(t, ok)
	assert.Equal(t, time.UTC, got.Location())
	assert.Equal(t, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), got)
}

func TestInternalFlag_Value_Expired(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	vTrue, vFalse := any(true), any(false)
	f := flag.InternalFlag{
		Variations:  &map[string]*any{"on": &vTrue, "off": &vFalse},
		DefaultRule: &flag.Rule{VariationResult: testutilsStrPtr("on")},
		CreatedAt:   &created,
	}
	ctx := ffcontext.NewEvaluationContext("user")
	ctx.AddCustomAttribute("gofeatureflag", map[string]any{"currentDateTime": created.AddDate(0, 0, 31)})

	value, details := f.Value("my-flag", ctx, flag.Context{DefaultSdkValue: false, DefaultTTL: 30 * 24 * time.Hour})
	assert.Equal(t, false, value)
	assert.Equal(t, flag.ReasonError, details.Reason)
	assert.Equal(t, flag.ErrorCodeFlagExpired, details.ErrorCode)
	assert.Equal(t, `flag "my-flag" expired on 2026-01-31T00:00:00Z`, details.ErrorMessage)

	value, details = f.Value("my-flag", ctx, flag.Context{DefaultSdkValue: false, DefaultTTL: 90 * 24 * time.Hour})
	assert.Equal(t, true, value)
	assert.Empty(t, details.ErrorCode)
}

func testutilsStrPtr(s string) *string { return &s }

func TestInternalFlag_Value_ExpiredMessageIsUTC(t *testing.T) {
	paris := time.FixedZone("CEST", 2*60*60)
	expiresAt := time.Date(2026, 3, 1, 12, 0, 0, 0, paris)
	vTrue, vFalse := any(true), any(false)
	f := flag.InternalFlag{
		Variations:  &map[string]*any{"on": &vTrue, "off": &vFalse},
		DefaultRule: &flag.Rule{VariationResult: testutilsStrPtr("on")},
		ExpiresAt:   &expiresAt,
	}
	_, details := f.Value("my-flag", ffcontext.NewEvaluationContext("user"), flag.Context{DefaultSdkValue: false})
	assert.Equal(t, `flag "my-flag" expired on 2026-03-01T10:00:00Z`, details.ErrorMessage)
}
