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

import { useEffect } from "react";
import { graphql, useLazyLoadQuery } from "react-relay";
import { useParams } from "react-router";

import type { AssumeOrganizationSessionMutation } from "#/__generated__/iam/AssumeOrganizationSessionMutation.graphql";
import type { AssumeOrganizationSessionQuery } from "#/__generated__/iam/AssumeOrganizationSessionQuery.graphql";
import { loginSearch, redirectToLogin } from "#/lib/auth/redirectToLogin";
import { useMutation } from "#/lib/relay/useMutation";
import { MainLayoutSkeleton } from "#/pages/iam/MainLayoutSkeleton";

const assumeOrganizationSessionQuery = graphql`
  query AssumeOrganizationSessionQuery {
    viewer @required(action: THROW) {
      ssoLoginURL
    }
  }
`;

const assumeMutation = graphql`
  mutation AssumeOrganizationSessionMutation(
    $input: AssumeEmployeePortalSessionInput!
  ) {
    assumeEmployeePortalSession(input: $input) {
      organizationId
      result {
        __typename
        ... on PasswordRequired {
          reason
        }
        ... on SAMLAuthenticationRequired {
          reason
        }
      }
    }
  }
`;

export function AssumeOrganizationSession() {
  const { employeePortalId } = useParams();
  const { viewer } = useLazyLoadQuery<AssumeOrganizationSessionQuery>(
    assumeOrganizationSessionQuery,
    {},
    { fetchPolicy: "network-only" },
  );
  const [assume] = useMutation<AssumeOrganizationSessionMutation>(assumeMutation);

  useEffect(() => {
    if (employeePortalId == null) {
      redirectToLogin();
      return;
    }

    void assume({
      variables: {
        input: {
          employeePortalId,
          continue: window.location.href,
        },
      },
    }).then((response) => {
      const payload = response.assumeEmployeePortalSession;
      const organizationId = payload?.organizationId;
      const result = payload?.result;
      if (result == null) {
        redirectToLogin({ organizationId });
        return;
      }

      switch (result.__typename) {
        case "PasswordRequired":
          redirectToLogin({ organizationId });
          return;
        case "SAMLAuthenticationRequired": {
          if (!viewer.ssoLoginURL) {
            throw new Error("missing SSO login URL for user email");
          }
          const samlURL = new URL(viewer.ssoLoginURL);
          const search = loginSearch({ organizationId });
          for (const [key, value] of search) {
            samlURL.searchParams.set(key, value);
          }
          window.location.replace(samlURL);
          return;
        }
        default:
          window.location.reload();
      }
    }).catch(() => {
      // UnAuthenticatedError is consumed by useMutation (full redirect).
    });
  }, [employeePortalId, assume, viewer.ssoLoginURL]);

  return <MainLayoutSkeleton />;
}
