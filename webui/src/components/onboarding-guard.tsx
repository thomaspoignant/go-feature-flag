import { useGetIdentity } from "@refinedev/core";
import { Navigate, useLocation } from "react-router";
import type { PropsWithChildren } from "react";
import { getCachedMembership } from "@/providers/auth";
import type { User } from "@/types/api";

export function OnboardingGuard({ children }: PropsWithChildren) {
  const { data: identity, isLoading } = useGetIdentity<User>();
  const location = useLocation();

  if (isLoading || !identity) return null;

  const membership = getCachedMembership() ?? [];
  const needsOnboarding = !identity.isSuperAdmin && membership.length === 0;

  if (needsOnboarding && !location.pathname.startsWith("/onboarding")) {
    return <Navigate to="/onboarding" replace />;
  }

  return <>{children}</>;
}

OnboardingGuard.displayName = "OnboardingGuard";
