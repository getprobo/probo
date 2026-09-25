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

import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { BindingsPageQuery } from "#/__generated__/core/BindingsPageQuery.graphql";

import { BindingsList } from "./_components/BindingsList";
import { bindingsPage } from "./_components/variants";

export const bindingsPageQuery = graphql`
  query BindingsPageQuery @throwOnFieldError {
    viewer @required(action: THROW) {
      ...BindingsList_viewer
    }
  }
`;

interface BindingsPageProps {
  queryRef: PreloadedQuery<BindingsPageQuery>;
}

export function BindingsPage({ queryRef }: BindingsPageProps) {
  const slots = bindingsPage();
  const { viewer } = usePreloadedQuery<BindingsPageQuery>(
    bindingsPageQuery,
    queryRef,
  );

  return (
    <main className={slots.main()}>
      <BindingsList viewerKey={viewer} />
    </main>
  );
}
