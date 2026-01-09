import { createContext } from "react";

interface ViewerContextValue {
  fullName: string;
  email: string;
}

export const Viewer = createContext<ViewerContextValue | undefined | null>({
  email: "",
  fullName: "",
});
