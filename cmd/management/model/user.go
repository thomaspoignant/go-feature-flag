package model

import "time"

type User struct {
	ID           string     `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	Name         string     `db:"name" json:"name"`
	OIDCSub      *string    `db:"oidc_sub" json:"-"`
	IsSuperAdmin bool       `db:"is_super_admin" json:"isSuperAdmin"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"lastLoginAt,omitempty"`
}

type MeResponse struct {
	User       User             `json:"user"`
	Membership []TeamMembership `json:"membership"`
}
