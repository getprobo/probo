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

import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { visitorPageSkeleton } from "./variants";

export function CompliancePortalVisitorPageSkeleton() {
  const { root, hero, profile, nda, section } = visitorPageSkeleton();

  return (
    <div className={root()}>
      <TextSkeleton size={2} className="w-20" />
      <div className={hero()}>
        <CardSkeleton size={3} className={profile()} />
        <CardSkeleton size={3} className={nda()} />
      </div>
      <div className={section()}>
        <HeadingSkeleton size={3} className="w-72" />
        <ListSkeleton count={4} />
      </div>
    </div>
  );
}
