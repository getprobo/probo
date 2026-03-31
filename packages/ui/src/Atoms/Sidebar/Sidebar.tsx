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
          "border-r border-border-solid relative pt-16 flex-none flex flex-col",
          open ? "px-4 w-[260px]" : "px-2",
        )}
      >
        <div className="flex-1 overflow-y-auto pb-14">{children}</div>
        <Button
          variant="tertiary"
          icon={open ? IconCollapse : IconExpand}
          onClick={() => setOpen(!open)}
          className="absolute bottom-4 left-4"
        />
      </aside>
    </sidebarContext.Provider>
  );
}

export function useSidebarCollapsed(): boolean {
  return !useContext(sidebarContext).open;
}
