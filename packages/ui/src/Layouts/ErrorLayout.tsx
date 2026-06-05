// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
