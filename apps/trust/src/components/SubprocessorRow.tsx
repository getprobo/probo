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
