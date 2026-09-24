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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextFieldSkeleton } from "@probo/ui/src/v2/form/TextFieldSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { userPageSkeleton } from "./variants";

export function UserPageSkeleton() {
  const { root, header, heading, identity, person, fields, properties, intro, dates } = userPageSkeleton();

  return (
    <div className={root()}>
      <TextSkeleton size={2} className="w-16" />
      <div className={header()}>
        <div className={heading()}>
          <HeadingSkeleton size={6} className="w-56" />
          <TextSkeleton size={2} className="w-40" />
        </div>
        <TextFieldSkeleton size={2} className="w-40" />
      </div>
      <Card variant="soft" size={2}>
        <div className={identity()}>
          <div className={person()}>
            <div className="flex w-32 shrink-0 flex-col gap-2">
              <span className="size-32 animate-pulse rounded-4 bg-sand-3" aria-hidden />
              <TextSkeleton size={1} className="w-full" />
            </div>
            <div className={fields()}>
              <TextFieldSkeleton size={2} className="w-full" />
              <TextFieldSkeleton size={2} className="w-full" />
              <TextFieldSkeleton size={2} className="w-full" />
            </div>
          </div>
        </div>
      </Card>
      <section className={properties()}>
        <div className={intro()}>
          <HeadingSkeleton size={4} className="w-28" />
          <TextSkeleton size={2} className="w-72" />
        </div>
        <Card variant="soft" size={2}>
          <div className={fields()}>
            <TextFieldSkeleton size={2} className="w-full" />
            <TextFieldSkeleton size={2} className="w-full" />
            <div className={dates()}>
              <TextFieldSkeleton size={2} className="w-full" />
              <TextFieldSkeleton size={2} className="w-full" />
            </div>
          </div>
        </Card>
      </section>
    </div>
  );
}
