import { clsx } from "clsx";
import {
  type PropsWithChildren,
  type ReactNode,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";

import { useSidebarCollapsed } from "./Sidebar";

type Props = PropsWithChildren<{
  label: string;
  defaultOpen?: boolean;
}>;

export function SidebarGroup(props: Props) {
  const { label, defaultOpen = true, children } = props;
  const isCollapsed = useSidebarCollapsed();
  const storageKey = `sidebar-group-${label}`;

  const [open, setOpen] = useState<boolean>(() => {
    const stored = localStorage.getItem(storageKey);
    return stored !== null ? !!JSON.parse(stored) : defaultOpen;
  });

  const contentRef = useRef<HTMLDivElement>(null);
  const [height, setHeight] = useState<number | undefined>(undefined);

  const updateHeight = useCallback(() => {
    if (contentRef.current) {
      setHeight(contentRef.current.scrollHeight);
    }
  }, []);

  useEffect(() => {
    updateHeight();
  }, [children, updateHeight]);

  useEffect(() => {
    if (!contentRef.current) return;
    const observer = new ResizeObserver(() => updateHeight());
    observer.observe(contentRef.current);
    return () => observer.disconnect();
  }, [updateHeight]);

  const toggle = () => {
    const next = !open;
    setOpen(next);
    localStorage.setItem(storageKey, JSON.stringify(next));
  };

  if (isCollapsed) {
    return <ul className="space-y-0.5">{children}</ul>;
  }

  return (
    <div className="mt-3 first:mt-0">
      <button
        type="button"
        onClick={toggle}
        className="flex items-center justify-between w-full px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-sidebar-text cursor-pointer select-none hover:text-sidebar-text-hover transition-colors duration-200"
      >
        <span>{label}</span>
        <SidebarGroupIcon open={open} />
      </button>
      <div
        className="overflow-hidden transition-[height,opacity] duration-300 ease-in-out"
        style={{
          height: open ? height : 0,
          opacity: open ? 1 : 0,
        }}
      >
        <div ref={contentRef}>
          <ul className="space-y-0.5 pl-1 border-l border-sidebar-border ml-2">
            {children}
          </ul>
        </div>
      </div>
    </div>
  );
}

function SidebarGroupIcon(props: { open: boolean }): ReactNode {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      className={clsx(
        "transition-transform duration-300 ease-in-out shrink-0",
        props.open ? "rotate-0" : "-rotate-90",
      )}
    >
      <path
        d="M4 6L8 10L12 6"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
