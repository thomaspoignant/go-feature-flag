import { describe, it, expect } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw-server";
import { dataProvider } from "./data";
import { API_URL } from "./constants";
import type { Team } from "@/types/api";

const sampleTeam: Team = {
  id: "t1",
  name: "alpha",
  description: "",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

describe("dataProvider", () => {
  it("getList unwraps APIResponse.data array", async () => {
    server.use(
      http.get(`${API_URL}/teams`, () =>
        HttpResponse.json({ success: true, data: [sampleTeam] })
      )
    );
    const res = await dataProvider.getList({ resource: "teams" });
    expect(res.total).toBe(1);
    expect(res.data[0]).toMatchObject({ id: "t1", name: "alpha" });
  });

  it("create POSTs and unwraps data", async () => {
    server.use(
      http.post(`${API_URL}/teams`, async ({ request }) => {
        const body = (await request.json()) as { name: string };
        return HttpResponse.json(
          {
            success: true,
            data: { ...sampleTeam, name: body.name },
          },
          { status: 201 }
        );
      })
    );
    const res = await dataProvider.create({
      resource: "teams",
      variables: { name: "beta" },
    });
    expect(res.data).toMatchObject({ name: "beta" });
  });

  it("surfaces message on error", async () => {
    server.use(
      http.post(`${API_URL}/teams`, () =>
        HttpResponse.json(
          { success: false, message: "name is required" },
          { status: 400 }
        )
      )
    );
    await expect(
      dataProvider.create({ resource: "teams", variables: {} })
    ).rejects.toThrow("name is required");
  });
});
