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

import { BuildingsIcon, MapPinSimpleIcon } from "@phosphor-icons/react";
import { faviconUrl } from "@probo/helpers";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { BackdropCard } from "#/components/BackdropCard/BackdropCard";

import { useCountryLabel } from "../_lib/useCountryLabel";

import type { SubprocessorListItem_subprocessor$key } from "./__generated__/SubprocessorListItem_subprocessor.graphql";
import { subprocessorListItem } from "./variants";

const subprocessorListItemFragment = graphql`
  fragment SubprocessorListItem_subprocessor on Subprocessor @throwOnFieldError {
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
  const slots = subprocessorListItem();

  return (
    <BackdropCard
      backdropSrc={logoUrl ?? undefined}
      media={(
        <div className={slots.logo()}>
          {logoUrl != null
            ? <img src={logoUrl} alt="" className={slots.logoImage()} />
            : <BuildingsIcon size={24} weight="duotone" className={slots.logoFallbackIcon()} />}
        </div>
      )}
    >
      <Text size={4} weight="medium" color="neutral" highContrast>
        {subprocessor.name}
      </Text>
      {subprocessor.description != null && subprocessor.description !== "" && (
        <Text size={2} color="neutral">
          {subprocessor.description}
        </Text>
      )}
      {countries !== "" && (
        <div className={slots.region()}>
          <MapPinSimpleIcon className={slots.regionIcon()} />
          <Text size={1} color="gold">
            {countries}
          </Text>
        </div>
      )}
    </BackdropCard>
  );
}
