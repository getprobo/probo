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
import { DateRangeFieldSkeleton } from "@probo/ui/src/v2/form/DateRangeFieldSkeleton";
import { TextFieldSkeleton } from "@probo/ui/src/v2/form/TextFieldSkeleton";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { TableSkeleton } from "@probo/ui/src/v2/Table/TableSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { auditLogPageSkeleton } from "./_components/variants";

export function AuditLogPageSkeleton() {
  const {
    root, header, intro, toolbar, action, filter, range, title, description, exportAction, more,
  } = auditLogPageSkeleton();

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <HeadingSkeleton size={6} className={title()} />
          <TextSkeleton size={2} className={description()} />
        </div>
        <ButtonSkeleton size={2} className={exportAction()} />
      </div>
      <div className={toolbar()}>
        <TextFieldSkeleton className={action()} />
        <SelectSkeleton className={filter()} />
        <SelectSkeleton className={filter()} />
        <DateRangeFieldSkeleton className={range()} />
        <ButtonSkeleton size={2} className={more()} />
      </div>
      <TableSkeleton variant="surface" layout="fixed" columns={5} count={8} />
    </div>
  );
}
