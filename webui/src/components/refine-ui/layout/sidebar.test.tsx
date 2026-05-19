import { describe, it, expect, vi, beforeAll } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { Sidebar } from "./sidebar";
import { SidebarProvider } from "@/components/ui/sidebar";

beforeAll(() => {
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }),
  });
});

vi.mock("@refinedev/core", () => ({
  useGetIdentity: vi.fn(),
  useMenu: vi.fn(),
  useLink: () => (props: { to: string; children: React.ReactNode }) => (
    <a href={props.to}>{props.children}</a>
  ),
  useRefineOptions: () => ({ title: { icon: null, text: "GO Feature Flag" } }),
}));

import { useGetIdentity, useMenu } from "@refinedev/core";

const menuItems = [
  {
    key: "admin",
    name: "Admin",
    label: "Admin",
    meta: { label: "Admin", requiresSuperAdmin: true },
    children: [
      {
        key: "teams",
        name: "teams",
        route: "/teams",
        meta: { label: "Teams", parent: "Admin" },
        children: [],
      },
    ],
  },
];

function renderSidebar() {
  return render(
    <MemoryRouter>
      <SidebarProvider>
        <Sidebar />
      </SidebarProvider>
    </MemoryRouter>
  );
}

describe("Sidebar - Admin visibility", () => {
  it("shows Admin group for super admin", () => {
    (useMenu as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      menuItems,
      selectedKey: undefined,
    });
    (useGetIdentity as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { isSuperAdmin: true },
    });
    renderSidebar();
    expect(screen.getByText("Admin")).toBeInTheDocument();
  });

  it("hides Admin group for non-super admin", () => {
    (useMenu as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      menuItems,
      selectedKey: undefined,
    });
    (useGetIdentity as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { isSuperAdmin: false },
    });
    renderSidebar();
    expect(screen.queryByText("Admin")).not.toBeInTheDocument();
  });

  it("hides Admin group while identity is loading", () => {
    (useMenu as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      menuItems,
      selectedKey: undefined,
    });
    (useGetIdentity as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
    });
    renderSidebar();
    expect(screen.queryByText("Admin")).not.toBeInTheDocument();
  });
});
