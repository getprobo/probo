import { Toasts } from "@probo/ui";
import { RouterProvider } from "react-router";

import { router } from "./routes";

export function App() {
  return (
    <>
      <RouterProvider router={router} />
      <Toasts />
    </>
  );
}
