import { useForm } from "@refinedev/react-hook-form";
import type { BaseRecord, HttpError as RefineHttpError } from "@refinedev/core";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  CreateView,
  CreateViewHeader,
} from "@/components/refine-ui/views/create-view";
import { Button } from "@/components/ui/button";
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
import { HttpError } from "@/providers/http";

export const teamCreateSchema = z.object({
  name: z.string().min(1, "Name is required").max(120),
  description: z.string().max(500),
});

export type TeamCreateValues = z.infer<typeof teamCreateSchema>;

export function TeamCreatePage() {
  const form = useForm<BaseRecord, RefineHttpError, TeamCreateValues>({
    resolver: zodResolver(teamCreateSchema),
    defaultValues: { name: "", description: "" },
    refineCoreProps: {
      resource: "teams",
      action: "create",
      redirect: "list",
      successNotification: () => ({
        message: "Team created",
        type: "success",
      }),
      errorNotification: (error) => ({
        message:
          (error as HttpError | undefined)?.message ?? "Failed to create team",
        type: "error",
      }),
    },
  });

  const {
    refineCore: { onFinish, formLoading },
    handleSubmit,
    setError,
    control,
  } = form;

  const onSubmit = handleSubmit(async (values) => {
    try {
      await onFinish(values);
    } catch (err) {
      if (err instanceof HttpError && err.errors) {
        for (const [field, messages] of Object.entries(err.errors)) {
          setError(field as keyof TeamCreateValues, {
            type: "server",
            message: messages.join(", "),
          });
        }
      }
    }
  });

  return (
    <CreateView>
      <CreateViewHeader title="Create team" />
      <Form {...form}>
        <form onSubmit={onSubmit} className="flex flex-col gap-4 max-w-xl">
          <FormField
            control={control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input placeholder="my-team" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={control}
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
          <div className="flex justify-end">
            <Button type="submit" disabled={formLoading}>
              {formLoading ? "Creating..." : "Create"}
            </Button>
          </div>
        </form>
      </Form>
    </CreateView>
  );
}

TeamCreatePage.displayName = "TeamCreatePage";
