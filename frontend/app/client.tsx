import { createRoot } from "react-dom/client";
import { RouterProvider, createRouter } from "@tanstack/react-router";
import { createRoute, createRootRoute } from "@tanstack/react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import RootLayout from "./__root";
import HomePage from "./index";
import ClanDetailPage from "./clans.$tag";
import PlayerDetailPage from "./players.$tag";
import LeaderboardsPage from "./leaderboards";
import SearchPage from "./search";
import GoldPassPage from "./gold-pass";
import AdminLayout from "./admin/__layout";
import WarMonitorPage from "./admin/war";
import CWLPage from "./admin/cwl";
import FindLayoutPage from "./find-layout";
import FindLayoutResultsPage from "./find-layout.results.$jobId";
import LayoutDetailPage from "./layouts.$layoutId";
import LeaguesPage from "./leagues";
import "../styles/globals.css";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
});

const rootRoute = createRootRoute({ component: RootLayout });

const indexRoute = createRoute({ getParentRoute: () => rootRoute, path: "/", component: HomePage });
const clanRoute = createRoute({ getParentRoute: () => rootRoute, path: "/clans/$tag", component: ClanDetailPage });
const playerRoute = createRoute({ getParentRoute: () => rootRoute, path: "/players/$tag", component: PlayerDetailPage });
const leaderboardsRoute = createRoute({ getParentRoute: () => rootRoute, path: "/leaderboards", component: LeaderboardsPage });
const searchRoute = createRoute({ getParentRoute: () => rootRoute, path: "/search", component: SearchPage });
const goldPassRoute = createRoute({ getParentRoute: () => rootRoute, path: "/gold-pass", component: GoldPassPage });
const adminRoute = createRoute({ getParentRoute: () => rootRoute, path: "/admin", component: AdminLayout });
const adminWarRoute = createRoute({ getParentRoute: () => adminRoute, path: "/war", component: WarMonitorPage });
const adminCwlRoute = createRoute({ getParentRoute: () => adminRoute, path: "/cwl", component: CWLPage });
const findLayoutRoute = createRoute({ getParentRoute: () => rootRoute, path: "/find-layout", component: FindLayoutPage });
const findLayoutResultsRoute = createRoute({ getParentRoute: () => rootRoute, path: "/find-layout/results/$jobId", component: FindLayoutResultsPage });
const layoutDetailRoute = createRoute({ getParentRoute: () => rootRoute, path: "/layouts/$layoutId", component: LayoutDetailPage });
const leaguesRoute = createRoute({ getParentRoute: () => rootRoute, path: "/leagues", component: LeaguesPage });

const routeTree = rootRoute.addChildren([
  indexRoute, clanRoute, playerRoute, leaderboardsRoute,
  searchRoute, goldPassRoute,
  findLayoutRoute, findLayoutResultsRoute, layoutDetailRoute,
  leaguesRoute,
  adminRoute.addChildren([adminWarRoute, adminCwlRoute]),
]);

const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register { router: typeof router }
}

createRoot(document.getElementById("root")!).render(
  <QueryClientProvider client={queryClient}>
    <RouterProvider router={router} />
  </QueryClientProvider>,
);
