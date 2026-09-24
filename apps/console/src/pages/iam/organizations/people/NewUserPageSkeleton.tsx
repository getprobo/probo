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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextFieldSkeleton } from "@probo/ui/src/v2/form/TextFieldSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { newUserPageSkeleton } from "./variants";

export function NewUserPageSkeleton() {
  const { root, intro, form } = newUserPageSkeleton();

  return (
    <div className={root()}>
      <TextSkeleton size={2} className="w-16" />
      <div className={intro()}>
        <HeadingSkeleton size={6} className="w-40" />
        <TextSkeleton size={2} className="w-80" />
      </div>
      <Card variant="soft" size={2}>
        <div className={form()}>
          <TextFieldSkeleton size={2} className="w-full" />
          <TextFieldSkeleton size={2} className="w-full" />
          <TextFieldSkeleton size={2} className="w-full" />
          <TextFieldSkeleton size={2} className="w-full" />
          <TextFieldSkeleton size={2} className="w-full" />
          <ButtonSkeleton size={2} className="w-24" />
        </div>
      </Card>
    </div>
  );
}
