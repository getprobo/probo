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

import type { PreloadedQuery } from "react-relay";
import { graphql, useFragment, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import type { MainLayout_organization$key } from "#/__generated__/iam/MainLayout_organization.graphql";
import type { MainLayoutQuery } from "#/__generated__/iam/MainLayoutQuery.graphql";
import type { TopBar_organization$key } from "#/__generated__/iam/TopBar_organization.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { RelayProvider } from "#/lib/relay/RelayProvider";
import { QueueTopBar } from "#/pages/_components/QueueTopBar";
import {
  DocumentQueueProvider,
  useDocumentQueueActive,
} from "#/pages/_lib/DocumentQueueContext";
import { TopBar } from "#/pages/iam/_components/TopBar/TopBar";
import { ViewerIdentityProvider } from "#/pages/iam/_lib/ViewerIdentityContext";

export const mainLayoutQuery = graphql`
  query MainLayoutQuery($organizationId: ID!) @throwOnFieldError {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...TopBar_organization
        ...MainLayout_organization
      }
    }
  }
`;

const mainLayoutFragment = graphql`
  fragment MainLayout_organization on Organization {
    viewer @required(action: THROW) {
      identity @required(action: THROW) {
        fullName
      }
    }
  }
`;

interface MainLayoutProps {
  queryRef: PreloadedQuery<MainLayoutQuery>;
}

export function MainLayout({ queryRef }: MainLayoutProps) {
  const data = usePreloadedQuery<MainLayoutQuery>(mainLayoutQuery, queryRef);

  if (data.organization == null || data.organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  const organization = useFragment<MainLayout_organization$key>(
    mainLayoutFragment,
    data.organization,
  );

  return (
    <ViewerIdentityProvider fullName={organization.viewer.identity.fullName}>
      <DocumentQueueProvider>
        <MainLayoutChrome organizationKey={data.organization} />
      </DocumentQueueProvider>
    </ViewerIdentityProvider>
  );
}

// Swaps the org top bar for the queue chrome when the current document is in
// the frozen signing / approval snapshot.
function MainLayoutChrome({
  organizationKey,
}: {
  organizationKey: TopBar_organization$key;
}) {
  const queueActive = useDocumentQueueActive();

  return (
    <div className="flex min-h-dvh flex-col bg-sand-2">
      {queueActive
        ? <QueueTopBar />
        : <TopBar organizationKey={organizationKey} />}
      <RelayProvider>
        <div className="flex min-h-0 flex-1 flex-col">
          <Outlet />
        </div>
      </RelayProvider>
    </div>
  );
}
