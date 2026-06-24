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

import type { PropsWithChildren, ReactNode } from "react";
import { tv, type VariantProps } from "tailwind-variants";

import { Card } from "../Atoms/Card/Card";
import { IconChevronDown } from "../Atoms/Icons";
import { Logo } from "../Atoms/Logo/Logo";

const errorLayout = tv({
  slots: {
    root: "w-full flex items-center justify-center px-6 py-16 bg-level-0 text-txt-primary",
    card: "w-full max-w-md flex flex-col items-center text-center px-8 py-10",
    brand: "w-[110px] mb-8",
    title: "text-xl font-semibold tracking-tight",
    description: "text-sm text-txt-secondary mt-2 leading-relaxed max-w-xs",
    details: "mt-6 w-full border-t border-border-mid pt-6",
    actions: "mt-8 w-full flex justify-center",
  },
  variants: {
    fullPage: {
      true: {
        root: "min-h-screen",
      },
      false: {
        root: "py-12",
      },
    },
  },
  defaultVariants: {
    fullPage: true,
  },
});

const errorDetails = tv({
  slots: {
    root: "w-full text-start group",
    summary:
      "text-sm text-txt-tertiary cursor-pointer select-none list-none flex items-center justify-center gap-1 hover:text-txt-secondary transition-colors [&::-webkit-details-marker]:hidden",
    marker: "transition-transform group-open:rotate-180",
    panel: "mt-3 rounded-lg bg-subtle px-3 py-2.5",
    message: "text-xs font-mono text-txt-tertiary break-all leading-relaxed",
  },
});

type ErrorLayoutProps = PropsWithChildren<
  {
    title: string;
    description?: string;
    actions?: ReactNode;
    showLogo?: boolean;
  } & VariantProps<typeof errorLayout>
>;

export function ErrorLayout({
  title,
  description,
  fullPage,
  showLogo = false,
  actions,
  children,
}: ErrorLayoutProps) {
  const classNames = errorLayout({ fullPage });

  return (
    <div className={classNames.root()}>
      <Card className={classNames.card()}>
        {showLogo && (
          <Logo withPicto className={classNames.brand()} />
        )}
        <h1 className={classNames.title()}>{title}</h1>
        {description && (
          <p className={classNames.description()}>{description}</p>
        )}
        {children && <div className={classNames.details()}>{children}</div>}
        {actions && <div className={classNames.actions()}>{actions}</div>}
      </Card>
    </div>
  );
}

type ErrorDetailsProps = {
  summary: string;
  children: ReactNode;
};

export function ErrorDetails({ summary, children }: ErrorDetailsProps) {
  const classNames = errorDetails();

  return (
    <details className={classNames.root()}>
      <summary className={classNames.summary()}>
        {summary}
        <IconChevronDown size={14} className={classNames.marker()} />
      </summary>
      <div className={classNames.panel()}>{children}</div>
    </details>
  );
}

export function ErrorDetailMessage({ children }: PropsWithChildren) {
  const classNames = errorDetails();

  return <p className={classNames.message()}>{children}</p>;
}
