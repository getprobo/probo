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

import { Suspense, useEffect, useRef } from "react";
import { useQueryLoader } from "react-relay";

import type { UsersPageQuery } from "#/__generated__/iam/UsersPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { USER_PAGE_SIZE } from "./_components/UsersList";
import { useUsersListFilters } from "./_lib/useUsersListFilters";
import { UsersPage, usersPageQuery } from "./UsersPage";
import { UsersPageSkeleton } from "./UsersPageSkeleton";

function UsersPageQueryLoader() {
  const organizationId = useOrganizationId();
  const { graphqlFilter } = useUsersListFilters();
  const filterRef = useRef(graphqlFilter);
  const [queryRef, loadQuery] = useQueryLoader<UsersPageQuery>(usersPageQuery);

  useEffect(() => {
    filterRef.current = graphqlFilter;
  }, [graphqlFilter]);

  useEffect(() => {
    loadQuery({
      organizationId,
      first: USER_PAGE_SIZE,
      filter: filterRef.current,
    }, { fetchPolicy: "network-only" });
  }, [loadQuery, organizationId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <UsersPageSkeleton />;
  }

  return (
    <Suspense fallback={<UsersPageSkeleton />}>
      <UsersPage key={organizationId} queryRef={currentQueryRef} />
    </Suspense>
  );
}

export default function UsersPageLoader() {
  return (
    <IAMRelayProvider>
      <UsersPageQueryLoader />
    </IAMRelayProvider>
  );
}
