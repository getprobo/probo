import type { PropsWithChildren } from "react";
import { useState } from "react";
import { tv } from "tailwind-variants";
import { IconChevronDown, IconChevronRight } from "../Icons";
import { useSidebarCollapsed } from "./Sidebar.tsx";

const sidebarCategory = tv({
    base: "flex items-center gap-2 w-full py-2 px-3 rounded-full cursor-pointer",
    variants: {
        isCollapsed: {
            true: "px-[10px]",
            false: "px-3",
        },
    },
});

const categoryLabel = tv({
    base: "font-medium text-txt-primary text-sm",
});

const categoryContent = tv({
    base: "ml-4 space-y-[2px] overflow-hidden",
    variants: {
        expanded: {
            true: "block",
            false: "hidden",
        },
    },
});

type Props = PropsWithChildren<{
    label: string;
    defaultExpanded?: boolean;
}>;

export function SidebarCategory({ label, children, defaultExpanded = true }: Props) {
    const [isExpanded, setIsExpanded] = useState(defaultExpanded);
    const isCollapsed = useSidebarCollapsed();

    const handleToggle = () => {
        if (!isCollapsed) {
            setIsExpanded(!isExpanded);
        }
    };

    const ChevronIcon = isExpanded ? IconChevronDown : IconChevronRight;

    return (
        <li className="mb-2">
            <div
                className={sidebarCategory({ isCollapsed })}
                onClick={handleToggle}
            >
                {!isCollapsed && (
                    <>
                        <ChevronIcon size={16} className="text-txt-tertiary" />
                        <span className={categoryLabel()}>{label}</span>
                    </>
                )}
            </div>
            {!isCollapsed && (
                <ul className={categoryContent({ expanded: isExpanded })}>
                    {children}
                </ul>
            )}
            {isCollapsed && (
                <ul className="space-y-[2px]">
                    {children}
                </ul>
            )}
        </li>
    );
}