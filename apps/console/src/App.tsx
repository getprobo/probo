import { RouterProvider } from "react-router";
import { router } from "./routes";
import { Toasts, ConfirmDialog } from "@probo/ui";

export function App() {
  return (
    <>
      <RouterProvider router={router} />
      <Toasts />
      <ConfirmDialog />
    </>
  );
}
