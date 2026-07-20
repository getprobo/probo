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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import { useSafeContinueUrl } from "#/hooks/useSafeContinueUrl";

import type { ConnectPageQuery } from "./__generated__/ConnectPageQuery.graphql";

export const connectPageQuery = graphql`
  query ConnectPageQuery {
    currentCompliancePortal @required(action: THROW) {
      title
    }
  }
`;

export function ConnectPage(props: {
  queryRef: PreloadedQuery<ConnectPageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const safeContinueUrl = useSafeContinueUrl();

  const {
    currentCompliancePortal: { title },
  } = usePreloadedQuery<ConnectPageQuery>(connectPageQuery, queryRef);

  usePageTitle(__(`Connect to ${title}'s Compliance Page`));

  const initiateURL = new URL("/initiate", window.location.origin);
  initiateURL.searchParams.set("continue", safeContinueUrl.toString());

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">
          {__(`Connect to ${title}'s Compliance Page`)}
        </h1>
        <p className="text-txt-tertiary">
          {__(
            "Sign in to start requesting access to documents",
          )}
        </p>
      </div>

      <Button
        className="w-full h-10"
        onClick={() => {
          window.location.href = initiateURL.toString();
        }}
      >
        {__("Get Access")}
      </Button>
    </div>
  );
}
