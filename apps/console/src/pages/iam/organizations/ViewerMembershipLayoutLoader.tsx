import { UnAuthenticatedError } from "@probo/relay";
import { Skeleton } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { graphql, useMutation, useQueryLoader } from "react-relay";
import { useNavigate } from "react-router";

import type { ViewerMembershipLayoutLoader_assumeMutation } from "#/__generated__/iam/ViewerMembershipLayoutLoader_assumeMutation.graphql";
import type { ViewerMembershipLayoutQuery } from "#/__generated__/iam/ViewerMembershipLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import {
  ViewerMembershipLayout,
  viewerMembershipLayoutQuery,
} from "./ViewerMembershipLayout";

const ensureAssumingMutation = graphql`
  mutation ViewerMembershipLayoutLoader_assumeMutation(
    $input: AssumeOrganizationSessionInput!
  ) {
    assumeOrganizationSession(input: $input) {
      result {
        __typename
        ... on OrganizationSessionCreated {
          membership {
            id
            lastSession {
              id
              expiresAt
            }
          }
        }
        ... on PasswordRequired {
          reason
        }
        ... on SAMLAuthenticationRequired {
          reason
          redirectUrl
        }
      }
    }
  }
`;

function ViewerMembershipLayoutQueryLoader() {
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const [assumeOrganizationSession]
    = useMutation<ViewerMembershipLayoutLoader_assumeMutation>(
      ensureAssumingMutation,
    );
  const [queryRef, loadQuery] = useQueryLoader<ViewerMembershipLayoutQuery>(
    viewerMembershipLayoutQuery,
  );

  useEffect(() => {
    assumeOrganizationSession({
      variables: {
        input: { organizationId },
      },
      onError: (error) => {
        if (error instanceof UnAuthenticatedError) {
          void navigate("/auth/login");
          return;
        }
      },
      onCompleted: ({ assumeOrganizationSession }) => {
        if (!assumeOrganizationSession) {
          throw new Error("complete mutation result is empty");
        }

        const { result } = assumeOrganizationSession;

        switch (result.__typename) {
          case "PasswordRequired":
            void navigate("/auth/login");
            break;
          case "SAMLAuthenticationRequired":
            window.location.href = result.redirectUrl;
            break;
          default:
            loadQuery({
              organizationId,
              hideSidebar: false,
            });
        }
      },
    });
  }, [navigate, assumeOrganizationSession, loadQuery, organizationId]);

  if (!queryRef) {
    return <Skeleton className="w-full h-screen" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <ViewerMembershipLayout queryRef={queryRef} />
    </Suspense>
  );
}

export default function ViewerMembershipLayoutLoader() {
  return (
    <IAMRelayProvider>
      <ViewerMembershipLayoutQueryLoader />
    </IAMRelayProvider>
  );
}
