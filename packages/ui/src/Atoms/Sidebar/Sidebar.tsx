import {
    createContext,
    type PropsWithChildren,
    useContext,
    useState,
} from "react";
import { clsx } from "clsx";
import { IconCollapse, IconExpand } from "../Icons";
import { Button } from "../Button/Button";

const sidebarContext = createContext({ open: true });

export function Sidebar({ children }: PropsWithChildren) {
    const [open, setOpen] = useState(true);
    return (
        <sidebarContext.Provider value={{ open }}>
            <aside
                className={clsx(
                    "border-r border-border-solid relative pt-16 flex-none",
                    open ? "px-4 w-[330px]" : "px-2",
                )}
            >
                {children}
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
