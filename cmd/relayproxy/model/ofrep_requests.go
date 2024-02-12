package model

type OFREPEvalFlagRequest struct {
	Context map[string]any `json:"context" xml:"context" form:"context" query:"context" swaggertype:"object,string" example:"targetingKey:4f433951-4c8c-42b3-9f18-8c9a5ed8e9eb,firstname:John,lastname:Doe,company:GO Feature Flag"` // nolint: lll
}

type OFREPBulkEvalFlagRequest struct {
	Context map[string]any `json:"context" xml:"context" form:"context" query:"context" swaggertype:"object,string" example:"targetingKey:4f433951-4c8c-42b3-9f18-8c9a5ed8e9eb,firstname:John,lastname:Doe,company:GO Feature Flag"` // nolint: lll
	Flags   []string       `json:"flags" xml:"flags" form:"flags" query:"flags" swaggertype:"array,string" example:"my-first-flag,another-flag"`                                                                                     // nolint: lll
}
