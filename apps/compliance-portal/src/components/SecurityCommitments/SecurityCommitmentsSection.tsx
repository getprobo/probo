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
