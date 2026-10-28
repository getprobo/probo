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

import { ButtonSkeleton } from "@probo/ui/src/v2/Button/ButtonSkeleton";
import { TableSkeleton } from "@probo/ui/src/v2/Table/TableSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { internalControlsPage } from "./_components/variants";

export function InternalControlsPageSkeleton() {
  const { root, header, intro, title, actions } = internalControlsPage();

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <div className={title()}>
            <HeadingSkeleton size={6} className="w-48" />
            <TextSkeleton size={2} className="w-4" />
          </div>
          <TextSkeleton size={2} className="w-96" />
        </div>
        <div className={actions()}>
          <ButtonSkeleton size={2} className="w-44" />
          <ButtonSkeleton size={2} className="w-8" />
        </div>
      </div>
      <TableSkeleton size={2} variant="surface" columns={5} count={6} />
    </div>
  );
}
