import { useGetIdentity } from "@refinedev/core";
import { Badge } from "@/components/ui/badge";
import { getCachedMembership } from "@/providers/auth";
import type { User } from "@/types/api";

export function HomePage() {
  const { data: identity, isLoading } = useGetIdentity<User>();
  const memberships = getCachedMembership() ?? [];

  if (isLoading || !identity) return null;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-4">
        <h1 className="text-2xl font-semibold">
          Welcome {identity.name || identity.email}
        </h1>
        <div className="flex flex-wrap gap-2 justify-end">
          {memberships.map((m) => (
            <Badge key={m.teamId} variant="secondary">
              {m.teamName}
            </Badge>
          ))}
        </div>
      </div>
    </div>
  );
}

HomePage.displayName = "HomePage";
