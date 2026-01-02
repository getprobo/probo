import { Role } from "@probo/helpers";
import { createContext } from "react";

type CurrentUserContextValue = {
  email: string;
  fullName: string;
  role: Role;
};

export const CurrentUser = createContext<CurrentUserContextValue>({
  email: "",
  fullName: "",
  role: Role.VIEWER,
});
