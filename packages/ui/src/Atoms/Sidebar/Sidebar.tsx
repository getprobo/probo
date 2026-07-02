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
          "border-r border-border-solid pt-16 flex-none flex flex-col",
          open && "w-[280px]",
        )}
      >
        <div
          className={clsx(
            "flex-1 pb-2",
            open ? "px-4" : "px-2",
          )}
        >
          {children}
        </div>
        <div
          className={clsx(
            "sticky bottom-0 flex-none border-t border-border-solid bg-level-0 py-2",
            open ? "px-4" : "px-2",
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
