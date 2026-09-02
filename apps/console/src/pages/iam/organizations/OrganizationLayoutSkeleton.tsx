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

import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { navPanel, navRail, organizationLayout } from "./_components/shell/variants";
import { NAV_GROUPS } from "./_lib/navigation";

// Widths track the English labels of the widest product (Governance).
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

export function OrganizationLayoutSkeleton() {
  const layoutSlots = organizationLayout();
  const railSlots = navRail({ expandable: false });
  const panelSlots = navPanel();

  return (
    <div className={layoutSlots.root()}>
      <div className={layoutSlots.body()}>
        <div className={railSlots.root()}>
          <div className={railSlots.rail()}>
            <div className={railSlots.icon()}>
              <div className="size-8 animate-pulse rounded-2 bg-sand-3" />
            </div>
            <div className={railSlots.items()}>
              {NAV_GROUPS.map(group => (
                <div key={group.key} className="size-10 animate-pulse rounded-3 bg-sand-3" />
              ))}
            </div>
            <div className="size-10 animate-pulse rounded-3 bg-sand-3" />
            <div className="size-10 animate-pulse rounded-3 bg-sand-3" />
            <div className={railSlots.icon()}>
              <div className="size-8 animate-pulse rounded-full bg-sand-3" />
            </div>
          </div>
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
