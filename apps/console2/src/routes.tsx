import { createBrowserRouter, useRouteError } from "react-router";
import { MainLayout } from "./layouts/MainLayout";
import { UnAuthenticatedError } from "./providers/RelayProviders";
import { AuthLayout } from "@probo/ui";
import { LoginPage } from "./pages/auth/LoginPage";

function Demo() {
  const error = useRouteError();
  if (error instanceof UnAuthenticatedError) {
    return <div></div>;
  }
  console.log(error);
  return <div>error</div>;
}

export const routes = createBrowserRouter([
  {
    path: "/auth",
    Component: AuthLayout,
    children: [
      {
        path: "login",
        element: <LoginPage />,
      },
    ],
  },
  {
    path: "/",
    Component: MainLayout,
    ErrorBoundary: Demo,
    children: [
      {
        path: "/",
        element: <div>Home</div>,
      },
    ],
  },
]);
