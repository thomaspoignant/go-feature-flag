import { Authenticated, Refine } from "@refinedev/core";
import { DevtoolsPanel, DevtoolsProvider } from "@refinedev/devtools";
import { RefineKbar, RefineKbarProvider } from "@refinedev/kbar";

import routerProvider, {
  DocumentTitleHandler,
  UnsavedChangesNotifier,
} from "@refinedev/react-router";
import { BrowserRouter, Outlet, Route, Routes } from "react-router";
import { UsersIcon } from "lucide-react";
import "./App.css";
import { Layout } from "./components/refine-ui/layout/layout";
import { Toaster } from "./components/refine-ui/notification/toaster";
import { useNotificationProvider } from "./components/refine-ui/notification/use-notification-provider";
import { ThemeProvider } from "./components/refine-ui/theme/theme-provider";
import { OnboardingGuard } from "./components/onboarding-guard";
import { SuperAdminGuard } from "./components/super-admin-guard";
import { authProvider } from "./providers/auth";
import { dataProvider } from "./providers/data";
import { HomeRoute } from "./pages/home/home-route";
import { LoginPage } from "./pages/login";
import { OnboardingChoicePage } from "./pages/onboarding/choice";
import { OnboardingCreateTeamPage } from "./pages/onboarding/create-team";
import { OnboardingJoinPage } from "./pages/onboarding/join";
import { TeamCreatePage } from "./pages/teams/create";
import { TeamListPage } from "./pages/teams/list";

function App() {
  return (
    <BrowserRouter>
      <RefineKbarProvider>
        <ThemeProvider>
          <DevtoolsProvider>
            <Refine
              dataProvider={dataProvider}
              authProvider={authProvider}
              notificationProvider={useNotificationProvider()}
              routerProvider={routerProvider}
              resources={[
                {
                  name: "Admin",
                  meta: { requiresSuperAdmin: true },
                },
                {
                  name: "teams",
                  list: "/teams",
                  create: "/teams/create",
                  meta: {
                    label: "Teams",
                    icon: <UsersIcon />,
                    canDelete: false,
                    parent: "Admin",
                  },
                },
              ]}
              options={{
                syncWithLocation: true,
                warnWhenUnsavedChanges: true,
                projectId: "woW2xm-RcGTKp-nEVscy",
              }}
            >
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route
                  path="/onboarding"
                  element={
                    <Authenticated
                      key="onboarding"
                      redirectOnFail="/login"
                      fallback={<LoginPage />}
                    >
                      <Outlet />
                    </Authenticated>
                  }
                >
                  <Route index element={<OnboardingChoicePage />} />
                  <Route
                    path="create-team"
                    element={<OnboardingCreateTeamPage />}
                  />
                  <Route path="join" element={<OnboardingJoinPage />} />
                </Route>
                <Route
                  element={
                    <Authenticated
                      key="protected"
                      redirectOnFail="/login"
                      fallback={<LoginPage />}
                    >
                      <OnboardingGuard>
                        <Layout>
                          <Outlet />
                        </Layout>
                      </OnboardingGuard>
                    </Authenticated>
                  }
                >
                  <Route index element={<HomeRoute />} />
                  <Route path="/teams">
                    <Route index element={<TeamListPage />} />
                    <Route
                      path="create"
                      element={
                        <SuperAdminGuard>
                          <TeamCreatePage />
                        </SuperAdminGuard>
                      }
                    />
                  </Route>
                </Route>
              </Routes>
              <Toaster />
              <RefineKbar />
              <UnsavedChangesNotifier />
              <DocumentTitleHandler />
            </Refine>
            <DevtoolsPanel />
          </DevtoolsProvider>
        </ThemeProvider>
      </RefineKbarProvider>
    </BrowserRouter>
  );
}

export default App;
