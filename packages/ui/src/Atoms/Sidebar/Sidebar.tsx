import { clsx } from "clsx";
import {
  createContext,
  type PropsWithChildren,
  useContext,
  useState,
} from "react";

import { Button } from "../Button/Button";
import { IconCollapse, IconExpand } from "../Icons";

const sidebarContext = createContext({ open: true });

function useSidebarState() {
  const [open, setOpenState] = useState<boolean>(() => {
    const stored = localStorage.getItem("sidebar-open");
    return stored !== null ? !!JSON.parse(stored) : true;
  });

  const setOpen = (value: boolean) => {
    setOpenState(value);

    localStorage.setItem("sidebar-open", JSON.stringify(value));
  };

  return [open, setOpen] as const;
}

export function Sidebar({ children }: PropsWithChildren) {
  const [open, setOpen] = useSidebarState();
  return (
    <sidebarContext.Provider value={{ open }}>
      <aside
        className={clsx(
          "border-r border-sidebar-border bg-sidebar-bg pt-16 flex-none flex flex-col h-screen",
          open ? "w-[260px]" : "",
        )}
      >
        <div
          className={clsx(
            "flex-1 min-h-0 overflow-y-auto",
            open ? "px-4" : "px-2",
          )}
        >
          {children}
        </div>
        <div
          className={clsx(
            "shrink-0 border-t border-sidebar-border bg-sidebar-bg flex flex-col gap-2 py-3",
            open ? "px-4" : "px-2 items-center",
          )}
        >
          <Button
            variant="tertiary"
            icon={open ? IconCollapse : IconExpand}
            onClick={() => setOpen(!open)}
          />
        </div>
      </aside>
    </sidebarContext.Provider>
  );
}

export function useSidebarCollapsed(): boolean {
  return !useContext(sidebarContext).open;
}
