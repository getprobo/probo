// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { faviconUrl, getCountryName } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { IconPin } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SubprocessorRowFragment$key } from "./__generated__/SubprocessorRowFragment.graphql";

const subprocessorRowFragment = graphql`
  fragment SubprocessorRowFragment on Subprocessor {
    name
    description
    websiteUrl
    countries
  }
`;

export function SubprocessorRow(props: { subprocessor: SubprocessorRowFragment$key; hasAnyCountries?: boolean }) {
  const subprocessor = useFragment(subprocessorRowFragment, props.subprocessor);
  const logo = faviconUrl(subprocessor.websiteUrl);
  const { __ } = useTranslate();

  return (
    <div className="flex text-sm leading-tight gap-6 items-center">
      {logo
        ? (
            <img
              src={logo}
              className="size-8 flex-none rounded-lg"
              alt=""
            />
          )
        : (
            <div className="size-8 flex-none rounded-lg" />
          )}
      <div className="flex flex-col gap-2 flex-1">
        <span className="text-sm">{subprocessor.name}</span>
        <div className="text-xs text-txt-secondary w-full">{subprocessor.description}</div>
        {props.hasAnyCountries
          && (
            <div className="text-xs flex gap-1 items-start text-txt-quaternary">
              {subprocessor.countries.length > 0 && (
                <>
                  <IconPin size={16} className="flex-none" />
                  <span>
                    {subprocessor.countries
                      .map(country => getCountryName(__, country))
                      .join(", ")}
                  </span>
                </>
              )}
            </div>
          )}
      </div>
    </div>
  );
}
