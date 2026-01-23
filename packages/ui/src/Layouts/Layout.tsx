import { clsx } from "clsx";
import {
  createContext,
  type PropsWithChildren,
  type ReactNode,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { createPortal } from "react-dom";
import { Link } from "react-router";

import { Logo } from "../Atoms/Logo/Logo";
import { Sidebar } from "../Atoms/Sidebar/Sidebar";
import { Toasts } from "../Atoms/Toasts/Toasts";
import { ConfirmDialog } from "../Molecules/Dialog/ConfirmDialog";

type Props = PropsWithChildren<{
  headerLeading?: ReactNode;
  headerTrailing: ReactNode;
  sidebar?: ReactNode;
}>;

const LayoutContext = createContext({
  setDrawer: (() => {}) as (v: boolean) => void,
});

export function Layout({
  headerLeading,
  headerTrailing,
  sidebar,
  children,
}: Props) {
  const [hasDrawer, setDrawer] = useState(false);
  const layoutContext = useMemo(
    () => ({
      setDrawer,
    }),
    [],
  );
  return (
    <LayoutContext value={layoutContext}>
      <div className="text-txt-primary bg-level-0">
        <header className="absolute z-2 left-0 right-0 px-4 flex items-center border-b border-border-solid h-12 bg-level-0">
          <Link to="/">
            <Logo className="w-12 h-5" />
          </Link>
          {headerLeading && (
            <>
              <svg
                className="mx-3 text-txt-tertiary"
                width="8"
                height="18"
                viewBox="0 0 8 18"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path d="M1 17L7 1" stroke="currentColor" />
              </svg>
              {headerLeading}
            </>
          )}
          <div className="ml-auto">{headerTrailing}</div>
        </header>
        <div className="flex h-screen" id="main">
          {sidebar && <Sidebar>{sidebar}</Sidebar>}
          <main
            className={clsx(
              "overflow-y-auto w-full mt-12 transition-all duration-300",
              hasDrawer && "pr-105",
            )}
          >
            <div className="py-12 px-8 max-w-[1200px] w-full mx-auto">
              {children}
            </div>
          </main>
        </div>
        <Toasts />
        <ConfirmDialog />
      </div>
    </LayoutContext>
  );
}

export function Drawer({
  children,
  className,
}: PropsWithChildren<{ className?: string }>) {
  const { setDrawer } = useContext(LayoutContext);
  useEffect(() => {
    setDrawer(true);
    return () => {
      setDrawer(false);
    };
  }, [setDrawer]);
  return createPortal(
    <aside
      className={clsx(
        "absolute pt-20 top-0 right-0 w-105 px-6 pb-8 border-border-solid border-l h-screen",
        className,
      )}
    >
      {children}
    </aside>,
    document.body,
  );
}
