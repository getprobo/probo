// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { CaretRightIcon } from "@phosphor-icons/react";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router";

import { NotFoundError } from "#/lib/relay/errors";

import { pageHeader } from "./variants";

export interface PageHeaderProps {
  homeLabel: string;
  parent?: {
    label: string;
    to: string;
  };
  currentLabel: string;
  title: string;
  actions?: ReactNode;
}

export function PageHeader({
  homeLabel,
  parent,
  currentLabel,
  title,
  actions,
}: PageHeaderProps) {
  const { t } = useTranslation();
  const { organizationId } = useParams();
  const slots = pageHeader();

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  return (
    <div className={slots.root()}>
      <nav className={slots.crumbs()} aria-label={t("breadcrumb.nav")}>
        <Link
          to={`/${organizationId}`}
          size={2}
          color="neutral"
          underline={false}
          className={slots.crumb()}
        >
          {homeLabel}
        </Link>
        <CaretRightIcon className={slots.chevron()} />
        {parent === undefined
          ? null
          : (
              <>
                <Link
                  to={parent.to}
                  size={2}
                  color="neutral"
                  underline={false}
                  className={slots.crumb()}
                >
                  {parent.label}
                </Link>
                <CaretRightIcon className={slots.chevron()} />
              </>
            )}
        <Text size={2} weight="medium" color="current" className={slots.crumbCurrent()} aria-current="page">
          {currentLabel}
        </Text>
      </nav>
      <div className={slots.titleRow()}>
        <Heading level={1} size={7} weight="medium" highContrast className="min-w-0">
          {title}
        </Heading>
        {actions != null && (
          <div className={slots.actions()}>
            {actions}
          </div>
        )}
      </div>
    </div>
  );
}
