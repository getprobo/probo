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

import { CommitmentCard } from "#/components/CommitmentCard/CommitmentCard";

import { SECURITY_COMMITMENT_GROUPS } from "./securityCommitments";
import { securityCommitments } from "./variants";

// "Security Commitments" section.
//
// TODO: This section renders placeholder data from a local POJO
// (./securityCommitments.ts) because there is no backend / DB structure for it
// yet. Replace it with a relay-driven fragment (and i18n copy) once available.
export function SecurityCommitmentsSection() {
  const slots = securityCommitments();

  return (
    <section className={slots.root()}>
      {SECURITY_COMMITMENT_GROUPS.map(group => (
        <div key={group.title} className={slots.group()}>
          <div className={slots.groupHeader()}>
            {group.eyebrow != null && (
              <Text size={1} color="gold">
                {group.eyebrow}
              </Text>
            )}
            <Text size={2} weight="medium" color="neutral" highContrast>
              {group.title}
            </Text>
            <Text size={2} color="neutral">
              {group.description}
            </Text>
          </div>
          <div className={slots.grid()}>
            {group.items.map(item => (
              <CommitmentCard
                key={item.title}
                icon={<item.Icon size={32} weight="light" />}
                eyebrow={<Text size={1} color="gold">{item.eyebrow}</Text>}
                title={(
                  <Text size={4} weight="medium" color="neutral" highContrast>
                    {item.title}
                  </Text>
                )}
                description={<Text size={2} color="neutral">{item.description}</Text>}
              />
            ))}
          </div>
        </div>
      ))}
    </section>
  );
}
