import { useState } from "react";
import { useNavigate } from "react-router";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { apiFetch, HttpError } from "@/providers/http";
import { refreshIdentity } from "@/providers/auth";
import type { Team } from "@/types/api";
import { cn } from "@/lib/utils";

const schema = z.object({
  name: z.string().min(1, "Name is required").max(120),
  description: z.string().max(500),
});

type FormValues = z.infer<typeof schema>;

export function OnboardingCreateTeamPage() {
  const navigate = useNavigate();
  const [submitting, setSubmitting] = useState(false);
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", description: "" },
  });

  const onSubmit = form.handleSubmit(async (values) => {
    setSubmitting(true);
    try {
      await apiFetch<Team>("/onboarding/team", {
        method: "POST",
        body: JSON.stringify(values),
      });
      await refreshIdentity();
      toast.success("Team created");
      navigate("/teams", { replace: true });
    } catch (err) {
      if (err instanceof HttpError && err.errors) {
        for (const [field, messages] of Object.entries(err.errors)) {
          form.setError(field as keyof FormValues, {
            type: "server",
            message: messages.join(", "),
          });
        }
      }
      const msg =
        err instanceof HttpError ? err.message : "Failed to create team";
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  });

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
          <CardTitle>Create your team</CardTitle>
          <CardDescription>
            You will be added as the admin of this team.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={onSubmit} className="flex flex-col gap-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Team name</FormLabel>
                    <FormControl>
                      <Input placeholder="my-team" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="What is this team responsible for?"
                        rows={4}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className="flex justify-between gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => navigate("/onboarding")}
                  disabled={submitting}
                >
                  Back
                </Button>
                <Button type="submit" disabled={submitting}>
                  {submitting ? "Creating..." : "Create team"}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}

OnboardingCreateTeamPage.displayName = "OnboardingCreateTeamPage";
