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
  type ComponentPropsWithRef,
  type FC,
  type PropsWithChildren,
  type ReactNode,
} from "react";

import { Card } from "../Card/Card";
import { IconPlusLarge } from "../Icons";
import { type AsChildProps, Slot } from "../Slot";

export function DataTable({
  children,
  className,
  columns,
}: PropsWithChildren<{ className?: string; columns: number | string[] }>) {
  const style = () => {
    if (typeof columns === "number") {
      return {
        gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`,
      };
    }
    return {
      gridTemplateColumns: columns.join(" "),
    };
  };

  return (
    <div className="overflow-auto relative w-full p-1 -m-1">
      <Card
        className={clsx(
          className,
          "min-w-min text-left grid overflow-hidden",
        )}
        style={style()}
      >
        {children}
      </Card>
    </div>
  );
}

export function CellHead({
  children,
  className,
  ...props
}: ComponentPropsWithRef<"div">) {
  return (
    <div
      {...props}
      className={clsx(
        "text-xs text-txt-tertiary font-semibold ",
        "first:pl-6 p-3 whitespace-nowrap",
        className,
      )}
    >
      {children}
    </div>
  );
}

export function Row(props: ComponentPropsWithRef<"div">) {
  return <div {...props} className={clsx("contents", props.className)} />;
}

export function Cell({
  asChild,
  ...props
}: AsChildProps<ComponentPropsWithRef<"div">>) {
  const Component = asChild ? Slot : "div";
  return (
    <Component
      data-cell
      {...props}
      className={clsx(
        "first:pl-6 p-3 text-sm text-txt-primary bg-tertiary border-t border-border-low flex flex-col justify-center",
        props.className,
      )}
    />
  );
}

export function RowButton({
  icon = IconPlusLarge,
  children,
  ...props
}: {
  children: ReactNode;
  icon?: FC<{ size: number; className?: string }>;
} & ComponentPropsWithRef<"button">) {
  const IconComponent = icon;
  return (
    <button
      {...props}
      className={clsx(
        "py-2 bg-highlight hover:bg-highlight-hover active:bg-highlight-pressed cursor-pointer w-full flex gap-2 items-center justify-center text-sm text-txt-secondary",
        props.className,
      )}
      style={{
        gridColumnEnd: -1,
        gridColumnStart: 1,
        ...props.style,
      }}
    >
      <IconComponent size={16} />
      {children}
    </button>
  );
}
