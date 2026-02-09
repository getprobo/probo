import { UnAuthenticatedError } from "@probo/relay";
import { useEffect } from "react";
import { useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { useAssumeMutation } from "#/__generated__/iam/useAssumeMutation.graphql";

import { useOrganizationId } from "../useOrganizationId";

interface UseAssumeParameters {
  onSuccess: () => void;
}

const assumeMutation = graphql`
  mutation useAssumeMutation(
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

export function useAssume(params: UseAssumeParameters) {
  const { onSuccess } = params;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const [assumeOrganizationSession] = useMutation<useAssumeMutation>(assumeMutation);

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
            onSuccess();
        }
      },
    });
  }, [onSuccess, navigate, assumeOrganizationSession, organizationId]);

  return;
}
