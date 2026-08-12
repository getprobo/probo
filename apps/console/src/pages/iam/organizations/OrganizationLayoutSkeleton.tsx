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

import { Logo } from "@probo/ui";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { navPanel, navRail, organizationLayout, topBar } from "./_components/shell/variants";
import { NAV_GROUPS } from "./_lib/navigation";

// Width per panel entry, sized to the English labels of the widest product
// (Governance) so the placeholder column reads like the real one.
// See .cursor/rules/skeleton-width-sync.mdc.
const PANEL_ITEMS = [
  { key: "frameworks", width: "w-24" },
  { key: "audits", width: "w-14" },
  { key: "findings", width: "w-16" },
  { key: "measures", width: "w-20" },
  { key: "documents", width: "w-22" },
  { key: "tasks", width: "w-12" },
  { key: "statementsOfApplicability", width: "w-44" },
] as const;

/**
 * Loading placeholder paired with OrganizationLayout. Shares its layout slots
 * and imports neither Relay nor Base UI, so it renders immediately.
 *
 * The rail shows the real icon count rather than a guess, since the number of
 * products does not depend on the query.
 */
export function OrganizationLayoutSkeleton() {
  const layoutSlots = organizationLayout();
  const topBarSlots = topBar();
  const railSlots = navRail();
  const panelSlots = navPanel();

  return (
    <div className={layoutSlots.root()}>
      <div className={topBarSlots.bar()}>
        <div className={topBarSlots.brand()}>
          <Logo className={topBarSlots.logo()} />
        </div>
        <span className={topBarSlots.separator()} aria-hidden>/</span>
        <TextSkeleton size={2} className="w-32" />
        <div className={topBarSlots.trailing()}>
          <TextSkeleton size={2} className="w-32" />
        </div>
      </div>

      <div className={layoutSlots.body()}>
        <div className={railSlots.rail()}>
          {NAV_GROUPS.map(group => (
            <div key={group.key} className="size-9 animate-pulse rounded-3 bg-sand-3" />
          ))}
        </div>

        <div className={panelSlots.panel()}>
          <TextSkeleton size={1} className={`w-20 ${panelSlots.title()}`} />
          <div className={panelSlots.list()}>
            {PANEL_ITEMS.map(item => (
              <TextSkeleton key={item.key} size={2} className={`my-1.5 ${item.width}`} />
            ))}
          </div>
        </div>

        <div className={layoutSlots.content()}>
          <div className={layoutSlots.contentInner()} />
        </div>
      </div>
    </div>
  );
}
