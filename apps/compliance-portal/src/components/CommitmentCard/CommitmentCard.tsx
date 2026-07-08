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

import type { ReactNode } from "react";

import { BackdropCard } from "#/components/BackdropCard/BackdropCard";

import { commitmentCard } from "./variants";

export interface CommitmentCardProps {
  // Leading icon shown at 32px over the dotted backdrop.
  icon: ReactNode;
  // Small accent eyebrow above the title.
  eyebrow: ReactNode;
  // Card heading.
  title: ReactNode;
  // Supporting body copy.
  description: ReactNode;
}

// Commitment card, composing the shared BackdropCard frame over its dotted
// backdrop. Region props are placed into the body; the consumer supplies
// typography (Text). Purely presentational, so it doubles as the layout for its
// skeleton.
export function CommitmentCard({ icon, eyebrow, title, description }: CommitmentCardProps) {
  const { icon: iconSlot } = commitmentCard();

  return (
    <BackdropCard media={<div className={iconSlot()}>{icon}</div>}>
      {eyebrow}
      {title}
      {description}
    </BackdropCard>
  );
}
