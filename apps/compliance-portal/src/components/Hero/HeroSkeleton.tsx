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

import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

import { hero, organizationContactInfo } from "./variants";

// Width per contact item, roughly sized to its typical content (hostname /
// email / address). See .cursor/rules/skeleton-width-sync.mdc.
const CONTACT_ITEMS = [
  { key: "website", width: "w-28" },
  { key: "email", width: "w-40" },
  { key: "location", width: "w-36" },
] as const;

// Loading placeholder paired with Hero: reuses the same layout slots with
// skeleton primitives. Imports no Relay, so it renders instantly.
export function HeroSkeleton() {
  const { content, section } = hero();
  const { root, item } = organizationContactInfo();

  return (
    <HeaderBand>
      <div className={content()}>
        <div className={section()}>
          <HeadingSkeleton size={8} className="w-80" />
          <TextSkeleton size={2} className="w-full max-w-2xl" />
        </div>
        <div className={root()}>
          {CONTACT_ITEMS.map(contact => (
            <div key={contact.key} className={item()}>
              <TextSkeleton size={2} className={contact.width} />
            </div>
          ))}
        </div>
      </div>
    </HeaderBand>
  );
}
