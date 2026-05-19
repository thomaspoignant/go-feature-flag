import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Refine } from "@refinedev/core";
import routerProvider from "@refinedev/react-router";
import { MemoryRouter, Route, Routes } from "react-router";
import { http, HttpResponse } from "msw";

import { server } from "@/test/msw-server";
import { dataProvider } from "@/providers/data";
import { API_URL } from "@/providers/constants";
import { TeamCreatePage } from "./create";

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/teams/create"]}>
      <Refine
        dataProvider={dataProvider}
        routerProvider={routerProvider}
        resources={[
          { name: "teams", list: "/teams", create: "/teams/create" },
        ]}
        options={{ disableTelemetry: true }}
      >
        <Routes>
          <Route path="/teams" element={<div>teams-list</div>} />
          <Route path="/teams/create" element={<TeamCreatePage />} />
        </Routes>
      </Refine>
    </MemoryRouter>
  );
}

describe("TeamCreatePage", () => {
  it("shows validation error when name is empty", async () => {
    renderPage();
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /create/i }));
    expect(await screen.findByText(/name is required/i)).toBeInTheDocument();
  });

  it("submits POST /teams and navigates to list on success", async () => {
    let captured: unknown = null;
    server.use(
      http.post(`${API_URL}/teams`, async ({ request }) => {
        captured = await request.json();
        return HttpResponse.json(
          {
            success: true,
            data: {
              id: "t1",
              name: "qa",
              description: "",
              createdAt: "2026-01-01T00:00:00Z",
              updatedAt: "2026-01-01T00:00:00Z",
            },
          },
          { status: 201 }
        );
      })
    );

    renderPage();
    const user = userEvent.setup();
    await user.type(screen.getByLabelText(/^name$/i), "qa");
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(screen.getByText("teams-list")).toBeInTheDocument()
    );
    expect(captured).toMatchObject({ name: "qa" });
  });

  it("surfaces API error message on 400", async () => {
    server.use(
      http.post(`${API_URL}/teams`, () =>
        HttpResponse.json(
          { success: false, message: "name already exists" },
          { status: 400 }
        )
      )
    );
    renderPage();
    const user = userEvent.setup();
    await user.type(screen.getByLabelText(/^name$/i), "dup");
    await user.click(screen.getByRole("button", { name: /create/i }));
    // Form stays on page; navigation does not occur.
    await waitFor(() =>
      expect(screen.queryByText("teams-list")).not.toBeInTheDocument()
    );
  });
});
