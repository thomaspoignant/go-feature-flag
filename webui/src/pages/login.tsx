import { useLogin } from "@refinedev/core";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

export function LoginPage() {
  const { mutate: login, isPending } = useLogin();

  return (
    <div
      className={cn(
        "flex",
        "min-h-svh",
        "items-center",
        "justify-center",
        "px-6",
        "py-8"
      )}
    >
      <Card className={cn("sm:w-[420px]", "p-8")}>
        <CardHeader className="px-0">
          <CardTitle className="text-2xl font-semibold">
            Sign in to GO Feature Flag
          </CardTitle>
          <CardDescription>
            You will be redirected to your identity provider.
          </CardDescription>
        </CardHeader>
        <CardContent className="px-0">
          <Button
            type="button"
            size="lg"
            className="w-full"
            disabled={isPending}
            onClick={() => login({})}
          >
            Sign in with SSO
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

LoginPage.displayName = "LoginPage";
