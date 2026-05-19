import { useGetIdentity, useLogout } from "@refinedev/core";
import { useNavigate } from "react-router";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { User } from "@/types/api";
import { cn } from "@/lib/utils";

export function OnboardingJoinPage() {
  const navigate = useNavigate();
  const { data: identity } = useGetIdentity<User>();
  const { mutate: logout } = useLogout();

  return (
    <div
      className={cn(
        "flex",
        "min-h-svh",
        "items-center",
        "justify-center",
        "px-6",
        "py-12"
      )}
    >
      <Card className="w-full max-w-xl">
        <CardHeader>
          <CardTitle>Ask an admin to add you</CardTitle>
          <CardDescription>
            Only a team admin can add new members. Share the email below with an
            admin of the team you want to join.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {identity?.email && (
            <div className="rounded-md border bg-muted p-4 font-mono text-sm">
              {identity.email}
            </div>
          )}
          <div className="flex justify-between gap-2">
            <Button variant="ghost" onClick={() => navigate("/onboarding")}>
              Back
            </Button>
            <Button variant="outline" onClick={() => logout({})}>
              Sign out
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

OnboardingJoinPage.displayName = "OnboardingJoinPage";
