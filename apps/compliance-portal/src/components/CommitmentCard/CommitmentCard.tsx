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
