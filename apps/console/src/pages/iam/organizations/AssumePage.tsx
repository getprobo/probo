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

import { UnAuthenticatedError } from "@probo/relay";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { AssumePageMutation } from "#/__generated__/iam/AssumePageMutation.graphql";
import type { AssumePageQuery } from "#/__generated__/iam/AssumePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useSafeContinueUrl } from "#/hooks/useSafeContinueUrl";

import AuthLayout from "../auth/AuthLayout";

const assumeMutation = graphql`
  mutation AssumePageMutation(
    $input: AssumeOrganizationSessionInput!
  ) {
    assumeOrganizationSession(input: $input) {
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

export const assumePageQuery = graphql`
  query AssumePageQuery {
    viewer @required(action: THROW) {
      __typename
      ... on Identity {
        ssoLoginURL
      }
    }
  }
`;

export function AssumePage(props: { queryRef: PreloadedQuery<AssumePageQuery> }) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { t } = useTranslation();

  const safeContinueUrl = useSafeContinueUrl(`/organizations/${organizationId}`);

  const { viewer } = usePreloadedQuery<AssumePageQuery>(assumePageQuery, queryRef);
  const [assumeOrganizationSession] = useMutation<AssumePageMutation>(assumeMutation);

  useEffect(() => {
    assumeOrganizationSession({
      variables: {
        input: { organizationId, continue: safeContinueUrl.toString() },
      },
      onError: (error) => {
        if (error instanceof UnAuthenticatedError) {
          const search = new URLSearchParams([
            ["organization-id", organizationId],
            ["continue", safeContinueUrl.toString()],
          ]);

          void navigate({ pathname: "/auth/login", search: "?" + search.toString() });
          return;
        }
      },
      onCompleted: ({ assumeOrganizationSession }) => {
        if (!assumeOrganizationSession) {
          throw new Error("complete mutation result is empty");
        }

        const { result } = assumeOrganizationSession;
        const search = new URLSearchParams();
        let samlSSOLoginURL: URL;

        switch (result.__typename) {
          case "PasswordRequired":
            search.set("organization-id", organizationId);
            search.set("continue", safeContinueUrl.toString());

            void navigate({ pathname: "/auth/login", search: "?" + search.toString() });
            break;
          case "SAMLAuthenticationRequired":
            if (!viewer.ssoLoginURL) {
              throw new Error("missing SSO login URL for user email");
            }
            samlSSOLoginURL = new URL(viewer.ssoLoginURL);
            samlSSOLoginURL.search = "?" + searchParams.toString();

            window.location.href = samlSSOLoginURL.toString();
            break;
          default:
            window.location.href = safeContinueUrl.toString();
        }
      },
    });
  }, [organizationId, navigate, assumeOrganizationSession, safeContinueUrl, searchParams, viewer.ssoLoginURL]);

  return (
    <AuthLayout>
      <div className="space-y-6 w-full max-w-md mx-auto pt-8">
        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold">{t("assumePage.title")}</h1>
          <p className="text-txt-tertiary">
            {t("assumePage.description")}
          </p>
        </div>
      </div>
    </AuthLayout>
  );
}
