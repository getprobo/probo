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

import { BuildingsIcon, MapPinSimpleIcon } from "@phosphor-icons/react";
import { faviconUrl } from "@probo/helpers";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { useCountryLabel } from "../_lib/useCountryLabel";

import type { SubprocessorListItem_subprocessor$key } from "./__generated__/SubprocessorListItem_subprocessor.graphql";

const subprocessorListItemFragment = graphql`
  fragment SubprocessorListItem_subprocessor on Subprocessor {
    name
    description
    websiteUrl
    countries
  }
`;

interface SubprocessorListItemProps {
  subprocessorKey: SubprocessorListItem_subprocessor$key;
}

// A single subprocessor card: a favicon logo over a blurred backdrop, the name,
// description, and the hosting regions.
export function SubprocessorListItem({ subprocessorKey }: SubprocessorListItemProps) {
  const subprocessor = useFragment(subprocessorListItemFragment, subprocessorKey);
  const countryLabel = useCountryLabel();
  const logoUrl = faviconUrl(subprocessor.websiteUrl);
  const countries = subprocessor.countries.map(countryLabel).join(", ");

  return (
    <Card variant="soft" size={3} padding="none">
      <div className="relative flex items-center overflow-hidden p-8">
        {logoUrl != null && (
          <img
            src={logoUrl}
            alt=""
            aria-hidden
            className="pointer-events-none absolute inset-0 size-full scale-150 object-cover opacity-10 blur-lg"
          />
        )}
        <div className="pointer-events-none absolute inset-0 bg-linear-to-b from-sand-1/0 to-sand-1" />
        <div className="relative z-10 flex size-10 items-center justify-center overflow-hidden rounded-2 bg-sand-1">
          {logoUrl != null
            ? <img src={logoUrl} alt="" className="size-full object-cover" />
            : <BuildingsIcon size={24} weight="duotone" className="text-sand-9" />}
        </div>
      </div>
      <div className="flex flex-col gap-2 px-8 pb-8">
        <Text size={4} weight="medium" color="neutral" highContrast>
          {subprocessor.name}
        </Text>
        {subprocessor.description != null && subprocessor.description !== "" && (
          <Text size={2} color="neutral">
            {subprocessor.description}
          </Text>
        )}
        {countries !== "" && (
          <div className="flex items-start gap-1">
            <MapPinSimpleIcon className="mt-0.5 size-4 shrink-0 text-gold-11" />
            <Text size={1} color="gold">
              {countries}
            </Text>
          </div>
        )}
      </div>
    </Card>
  );
}
