// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
        iconStart={<CaretRightIcon />}
        aria-label={nextLabel}
        className={hasNext ? undefined : "invisible"}
        onClick={onNext}
      />
    </nav>
  );
}
