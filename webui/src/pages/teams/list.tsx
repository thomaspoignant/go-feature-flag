import { useList, useGetIdentity, useNavigation } from "@refinedev/core";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Team, User } from "@/types/api";

export function TeamListPage() {
  const { result, query } = useList<Team>({ resource: "teams" });
  const isLoading = query.isLoading;
  const teams = result.data ?? [];
  const { data: identity } = useGetIdentity<User>();
  const { create } = useNavigation();

  return (
    <div className="flex flex-col gap-4 p-2">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Teams</h2>
        {identity?.isSuperAdmin && (
          <Button onClick={() => create("teams")}>New team</Button>
        )}
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Created</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading && (
            <TableRow>
              <TableCell colSpan={3}>Loading...</TableCell>
            </TableRow>
          )}
          {!isLoading && teams.length === 0 && (
            <TableRow>
              <TableCell colSpan={3}>No teams yet.</TableCell>
            </TableRow>
          )}
          {teams.map((t: Team) => (
            <TableRow key={t.id}>
              <TableCell>{t.name}</TableCell>
              <TableCell>{t.description}</TableCell>
              <TableCell>
                {t.createdAt ? new Date(t.createdAt).toLocaleString() : ""}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

TeamListPage.displayName = "TeamListPage";
