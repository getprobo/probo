// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { APIKeysPageQuery } from "#/__generated__/iam/APIKeysPageQuery.graphql";

import { PersonalAPIKeyList } from "./_components/PersonalAPIKeyList";

export const apiKeysPageQuery = graphql`
  query APIKeysPageQuery {
    viewer {
      ...PersonalAPIKeyListFragment
    }
  }
`;

export function APIKeysPage(props: {
  queryRef: PreloadedQuery<APIKeysPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();

  const data = usePreloadedQuery<APIKeysPageQuery>(apiKeysPageQuery, queryRef);

  return (
    <div className="space-y-6 w-full py-6">
      <h1 className="text-3xl font-bold text-center">{__("API Keys")}</h1>
      {data.viewer && <PersonalAPIKeyList fKey={data.viewer} />}
    </div>
  );
}
