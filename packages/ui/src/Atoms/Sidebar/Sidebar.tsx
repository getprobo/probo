import {
    createContext,
    type PropsWithChildren,
    useContext,
    useState,
} from "react";
import { clsx } from "clsx";
import { IconCollapse, IconExpand } from "../Icons";

const sidebarContext = createContext({ open: true });

export function Sidebar({ children }: PropsWithChildren) {
    const [open, setOpen] = useState(true);
    return (
        <sidebarContext.Provider value={{ open }}>
            <aside
                className={clsx(
                    "border-r border-border-solid relative pt-12 flex-none",
                    open ? "px-4 w-[260px]" : "px-2",
                )}
            >
                {children}
                <button
                    onClick={() => setOpen(!open)}
                    className="cursor-pointer absolute bottom-4 left-4"
                >
                    {open ? (
                        <IconCollapse size={16} />
                    ) : (
                        <IconExpand size={16} />
                    )}
                </button>
            </aside>
        </sidebarContext.Provider>
    );
}

export function useSidebarCollapsed(): boolean {
    return !useContext(sidebarContext).open;
}
