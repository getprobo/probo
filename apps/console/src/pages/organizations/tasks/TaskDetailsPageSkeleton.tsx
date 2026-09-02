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
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { taskDetailsPageSkeleton } from "./variants";

export function TaskDetailsPageSkeleton() {
  const { root, header, titleRow, title, body, description } = taskDetailsPageSkeleton();

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={titleRow()}>
          <div className={title()}>
            <HeadingSkeleton size={6} className="w-64" />
          </div>
        </div>
      </div>
      <div className={body()}>
        <div className={description()}>
          <TextSkeleton size={2} className="w-full" />
          <TextSkeleton size={2} className="w-5/6" />
          <TextSkeleton size={2} className="w-2/3" />
        </div>
        <CardSkeleton size={2} />
      </div>
    </div>
  );
}
