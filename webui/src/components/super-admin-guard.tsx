import { useGetIdentity } from "@refinedev/core";
import { Navigate } from "react-router";
import { toast } from "sonner";
import { useEffect } from "react";
import type { PropsWithChildren } from "react";
import type { User } from "@/types/api";

export function SuperAdminGuard({ children }: PropsWithChildren) {
  const { data: identity, isLoading } = useGetIdentity<User>();

  const forbidden = !isLoading && !identity?.isSuperAdmin;

  useEffect(() => {
    if (forbidden) {
      toast.error("Forbidden", {
        description: "Super admin access required.",
      });
    }
  }, [forbidden]);

  if (isLoading) return null;
  if (forbidden) return <Navigate to="/" replace />;
  return <>{children}</>;
}

SuperAdminGuard.displayName = "SuperAdminGuard";
