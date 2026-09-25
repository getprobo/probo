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

import { type ReactNode, Suspense } from "react";
import type { PreloadedQuery } from "react-relay";
import type { GraphQLTaggedNode, OperationType } from "relay-runtime";

import { useNavPanelQuery } from "#/pages/iam/organizations/_lib/useNavPanelQuery";

import { navPanel } from "./variants";

type NavPanelOperation = OperationType & { variables: { organizationId: string } };

export interface NavPanelQueryProps<TQuery extends NavPanelOperation> {
  query: GraphQLTaggedNode;
  children: (queryRef: PreloadedQuery<TQuery>) => ReactNode;
}

export function NavPanelQuery<TQuery extends NavPanelOperation>({
  query,
  children,
}: NavPanelQueryProps<TQuery>) {
  const queryRef = useNavPanelQuery<TQuery>(query);
  const slots = navPanel();
  const fallback = <span className={slots.groupFallback()} aria-hidden />;

  if (queryRef == null) {
    return fallback;
  }

  return <Suspense fallback={fallback}>{children(queryRef)}</Suspense>;
}
