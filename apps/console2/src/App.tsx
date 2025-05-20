import { RouterProvider } from "react-router";
import { router } from "./routes";
import { Toasts } from "@probo/ui";

export function App() {
  return (
    <>
      <RouterProvider router={router} />
      <Toasts />
    </>
  );
}
