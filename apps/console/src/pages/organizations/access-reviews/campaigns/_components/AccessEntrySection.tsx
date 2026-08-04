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

import { IconWarning, ThirdPartyLogo } from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { accessEntrySection } from "./variants";

interface AccessEntrySectionProps {
  title: string;
  count: number;
  provider?: string | null;
  error?: string | null;
  children: ReactNode;
}

// One connector group: header (logo + name + count) above a bordered list of entries.
export function AccessEntrySection({
  title,
  count,
  provider,
  error,
  children,
}: AccessEntrySectionProps) {
  const { t } = useTranslation();
  const { root, header, title: titleClass, count: countClass } = accessEntrySection();

  return (
    <section className={root()}>
      <div className={header()}>
        {provider
          ? <ThirdPartyLogo thirdParty={provider} className="size-5 shrink-0" />
          : null}
        <h2 className={titleClass()}>{title}</h2>
        <span className={countClass()}>{`(${count})`}</span>
      </div>
      {error
        ? (
            <div className="flex items-start gap-2 rounded-[10px] border border-border-danger bg-danger px-4 py-3 text-sm text-txt-danger">
              <IconWarning className="mt-0.5 size-4 shrink-0" />
              <div>
                <p className="font-medium">{t("campaignDetailPage.fetchFailed")}</p>
                <p>{error}</p>
              </div>
            </div>
          )
        : null}
      {children}
    </section>
  );
}
