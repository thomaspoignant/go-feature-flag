import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import { SuperAdminGuard } from "./super-admin-guard";

vi.mock("@refinedev/core", () => ({
  useGetIdentity: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn() },
}));

import { useGetIdentity } from "@refinedev/core";

function renderWithRoute(node: React.ReactNode) {
  return render(
    <MemoryRouter initialEntries={["/protected"]}>
      <Routes>
        <Route path="/" element={<div>home</div>} />
        <Route path="/protected" element={node} />
      </Routes>
    </MemoryRouter>
  );
}

describe("SuperAdminGuard", () => {
  it("renders children for super admin", async () => {
    (useGetIdentity as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { isSuperAdmin: true },
      isLoading: false,
    });
    renderWithRoute(
      <SuperAdminGuard>
        <div>secret</div>
      </SuperAdminGuard>
    );
    expect(await screen.findByText("secret")).toBeInTheDocument();
  });

  it("redirects non-super admin to /", async () => {
    (useGetIdentity as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { isSuperAdmin: false },
      isLoading: false,
    });
    renderWithRoute(
      <SuperAdminGuard>
        <div>secret</div>
      </SuperAdminGuard>
    );
    await waitFor(() =>
      expect(screen.getByText("home")).toBeInTheDocument()
    );
    expect(screen.queryByText("secret")).not.toBeInTheDocument();
  });

  it("renders nothing while loading", () => {
    (useGetIdentity as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: true,
    });
    const { container } = renderWithRoute(
      <SuperAdminGuard>
        <div>secret</div>
      </SuperAdminGuard>
    );
    expect(container.textContent).toBe("");
  });
});
