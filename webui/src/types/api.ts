export type Role = "admin" | "editor" | "viewer";

export interface APIResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
  errors?: Record<string, string[]>;
}

export interface Team {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface TeamMembership {
  teamId: string;
  teamName: string;
  role: Role;
}

export interface User {
  id: string;
  email: string;
  name: string;
  isSuperAdmin: boolean;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
}

export interface MeResponse {
  user: User;
  membership: TeamMembership[];
}

export interface CreateTeamRequest {
  name: string;
  description?: string;
}
