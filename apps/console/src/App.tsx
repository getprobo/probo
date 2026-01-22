import { RouterProvider } from "react-router";
import { Toasts, ConfirmDialog } from "@probo/ui";

import { router } from "./routes";

export function App() {
  return (
    <>
      <RouterProvider router={router} />
      <Toasts />
      <ConfirmDialog />
    </>
  );
}
