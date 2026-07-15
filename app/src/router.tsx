import {
  createBrowserRouter,
  RouterProvider,
  useLocation,
  useNavigate,
} from "react-router-dom";
import Home from "./routes/Home.tsx";
import NotFound from "./routes/NotFound.tsx";
import Auth from "./routes/Auth.tsx";
import React, { Suspense, useEffect } from "react";
import { lazyFactor } from "@/utils/loader.tsx";
import { useSelector } from "react-redux";
import { selectAdmin, selectAuthenticated, selectInit } from "@/store/auth.ts";
import Index from "@/routes/Index.tsx";
import Gallery from "@/routes/Gallery.tsx";

const Sharing = lazyFactor(() => import("@/routes/Sharing.tsx"));

const AdminPage = lazyFactor(() => import("@/routes/Admin.tsx"));
const AdminDashboard = lazyFactor(() => import("@/routes/admin/DashBoard.tsx"));
const AdminSystem = lazyFactor(() => import("@/routes/admin/System.tsx"));
const AdminLicense = lazyFactor(() => import("@/routes/admin/License.tsx"));
const AdminUsers = lazyFactor(() => import("@/routes/admin/Users.tsx"));
const AdminLogger = lazyFactor(() => import("@/routes/admin/Logger.tsx"));
const AdminGallery = lazyFactor(() => import("@/routes/admin/Gallery.tsx"));

const router = createBrowserRouter([
  {
    id: "index",
    path: "/",
    Component: Index,
    ErrorBoundary: NotFound,
    children: [
      {
        id: "not-found",
        path: "*",
        element: <NotFound />,
      },
      {
        id: "home",
        path: "",
        element: <Home />,
      },
      {
        id: "login",
        path: "/login",
        element: (
          <AuthForbidden>
            <Auth />
          </AuthForbidden>
        ),
        ErrorBoundary: NotFound,
      },
      {
        id: "gallery",
        path: "/gallery",
        element: <Gallery />,
        ErrorBoundary: NotFound,
      },
      {
        id: "admin",
        path: "/admin",
        element: (
          <AdminRequired>
            <Suspense>
              <AdminPage />
            </Suspense>
          </AdminRequired>
        ),
        children: [
          {
            id: "admin-dashboard",
            path: "",
            element: (
              <Suspense>
                <AdminDashboard />
              </Suspense>
            ),
          },
          {
            id: "admin-users",
            path: "users",
            element: (
              <Suspense>
                <AdminUsers />
              </Suspense>
            ),
          },
          {
            id: "admin-system",
            path: "system",
            element: (
              <Suspense>
                <AdminSystem />
              </Suspense>
            ),
          },
          {
            id: "admin-warm-up",
            path: "warmup",
            element: (
              <Suspense>
                <AdminLicense />
              </Suspense>
            ),
          },
          {
            id: "admin-license",
            path: "license",
            element: (
              <Suspense>
                <AdminLicense />
              </Suspense>
            ),
          },
          {
            id: "admin-logger",
            path: "logger",
            element: (
              <Suspense>
                <AdminLogger />
              </Suspense>
            ),
          },
          {
            id: "admin-gallery",
            path: "gallery",
            element: (
              <Suspense>
                <AdminGallery />
              </Suspense>
            ),
          },
        ],
        ErrorBoundary: NotFound,
      },
    ],
  },
  {
    id: "share",
    path: "/share/:hash",
    element: (
      <Suspense>
        <Sharing />
      </Suspense>
    ),
    ErrorBoundary: NotFound,
  },
]);

export function AuthRequired({ children }: { children: React.ReactNode }) {
  const init = useSelector(selectInit);
  const authenticated = useSelector(selectAuthenticated);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (init && !authenticated) {
      navigate("/login", { state: { from: location.pathname } });
    }
  }, [init, authenticated]);

  return <>{children}</>;
}

export function AuthForbidden({ children }: { children: React.ReactNode }) {
  const init = useSelector(selectInit);
  const authenticated = useSelector(selectAuthenticated);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (init && authenticated) {
      navigate("/", { state: { from: location.pathname } });
    }
  }, [init, authenticated]);

  return <>{children}</>;
}

export function AdminRequired({ children }: { children: React.ReactNode }) {
  const init = useSelector(selectInit);
  const admin = useSelector(selectAdmin);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (init && !admin) {
      navigate("/", { state: { from: location.pathname } });
    }
  }, [init, admin]);

  return <>{children}</>;
}

export function AppRouter() {
  return <RouterProvider router={router} />;
}

export default router;
