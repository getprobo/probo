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

import { CaretLeftIcon, CaretRightIcon } from "@phosphor-icons/react";
import type { ComponentProps, ReactNode } from "react";

import { Button } from "../Button/Button";
import { Text } from "../typography/Text";

import { pagination } from "./variants";

export type PaginationProps = Omit<ComponentProps<"nav">, "onChange"> & {
  hasPrevious: boolean;
  hasNext: boolean;
  // Current-position label rendered between the arrows (e.g. "Page 2").
  label?: ReactNode;
  // Accessible labels for the arrow controls.
  previousLabel?: string;
  nextLabel?: string;
  onPrevious: () => void;
  onNext: () => void;
};

// Prev/Next pager for cursor-paginated lists. Page numbers are intentionally
// omitted because cursor pagination cannot compute a total page count; an
// optional current-position label sits between the arrows instead. Each arrow
// only renders when its page exists, but its slot is always reserved (the
// missing arrow is kept invisible) so a visible arrow sits in the exact same
// position whether or not the other is present. Renders nothing when neither
// page exists.
export function Pagination(props: PaginationProps) {
  const {
    hasPrevious, hasNext, label,
    previousLabel = "Previous page", nextLabel = "Next page",
    onPrevious, onNext, className, ...rest
  } = props;
  const { root, label: labelSlot } = pagination();

  if (!hasPrevious && !hasNext) {
    return null;
  }

  return (
    <nav className={root({ className })} {...rest}>
      <Button
        variant="ghost"
        color="neutral"
        size={2}
        iconStart={<CaretLeftIcon />}
        aria-label={previousLabel}
        className={hasPrevious ? undefined : "invisible"}
        onClick={onPrevious}
      />
      {label != null && (
        <Text size={2} color="faint" className={labelSlot()}>
          {label}
        </Text>
      )}
      <Button
        variant="ghost"
        color="neutral"
        size={2}
        iconEnd={<CaretRightIcon />}
        aria-label={nextLabel}
        className={hasNext ? undefined : "invisible"}
        onClick={onNext}
      />
    </nav>
  );
}
