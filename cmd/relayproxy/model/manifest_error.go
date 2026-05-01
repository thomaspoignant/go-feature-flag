package model

type ManifestError struct {
	ErrorDetails ManifestErrorDetails `json:"error"`
}

type ManifestErrorDetails struct {
	Message string `json:"message" example:"Authorization header required"`
	Status  int    `json:"status" example:"401"`
}
