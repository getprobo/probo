// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

const LayoutContext = createContext<{ setDrawer: (v: boolean) => void }>({
  setDrawer: () => {},
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
      <div className="text-txt-primary bg-level-0 min-h-screen">
        <header className="fixed top-0 z-2 left-0 right-0 px-4 flex items-center border-b border-border-solid h-12 bg-level-0">
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
        <div className="flex min-h-screen" id="main">
          {sidebar && <Sidebar>{sidebar}</Sidebar>}
          <main
            className={clsx(
              "w-full mt-12 transition-all duration-300",
              hasDrawer && "pr-105",
            )}
          >
            <div className="py-12 px-8 max-w-[1200px] w-full mx-auto min-h-[calc(100vh-48px)]">
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
        "fixed pt-20 top-0 right-0 w-105 px-6 pb-8 border-border-solid border-l h-screen bg-level-0",
        className,
      )}
    >
      {children}
    </aside>,
    document.body,
  );
}
