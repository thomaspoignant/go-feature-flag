import { useNavigate } from "react-router";
import { PlusCircleIcon, UsersIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

export function OnboardingChoicePage() {
  const navigate = useNavigate();

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
      <div className="flex flex-col gap-8 max-w-4xl w-full">
        <div className="text-center">
          <h1 className="text-3xl font-semibold">Welcome to GO Feature Flag</h1>
          <p className="text-muted-foreground mt-2">
            To get started, create a new team or join an existing one.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card className="flex flex-col">
            <CardHeader>
              <PlusCircleIcon className="size-8 mb-2" />
              <CardTitle>Create a new GOFF team</CardTitle>
              <CardDescription>
                Start fresh. You will be the admin of the new team and can invite
                others.
              </CardDescription>
            </CardHeader>
            <CardContent className="mt-auto">
              <Button
                className="w-full"
                onClick={() => navigate("/onboarding/create-team")}
              >
                Create team
              </Button>
            </CardContent>
          </Card>
          <Card className="flex flex-col">
            <CardHeader>
              <UsersIcon className="size-8 mb-2" />
              <CardTitle>Join an existing team</CardTitle>
              <CardDescription>
                Your organization already uses GOFF. Ask an admin to add you.
              </CardDescription>
            </CardHeader>
            <CardContent className="mt-auto">
              <Button
                variant="outline"
                className="w-full"
                onClick={() => navigate("/onboarding/join")}
              >
                Join a team
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

OnboardingChoicePage.displayName = "OnboardingChoicePage";
