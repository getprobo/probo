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

import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { homeSection } from "./variants";

interface HomeSectionProps {
  title: ReactNode;
  // Optional trailing element in the header row (e.g. a "View all" link).
  action?: ReactNode;
  children: ReactNode;
}

// Shared frame for the home page body sections: a labelled header above the
// section content.
export function HomeSection({ title, action, children }: HomeSectionProps) {
  const slots = homeSection();

  return (
    <section className={slots.root()}>
      <div className={slots.header()}>
        <Text size={2} weight="medium" color="neutral" highContrast>
          {title}
        </Text>
        {action}
      </div>
      {children}
    </section>
  );
}
