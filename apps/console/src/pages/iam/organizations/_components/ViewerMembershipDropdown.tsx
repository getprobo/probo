import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  DropdownSeparator,
  IconArrowBoxLeft,
  IconCircleQuestionmark,
  IconKey,
  IconPageTextLine,
  UserDropdown,
  UserDropdownItem,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { ViewerMembershipDropdownFragment$key } from "#/__generated__/iam/ViewerMembershipDropdownFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const fragment = graphql`
  fragment ViewerMembershipDropdownFragment on Organization {
    viewerMembership @required(action: THROW) {
      identity @required(action: THROW) {
        email
        canListAPIKeys: permission(action: "iam:personal-api-key:list")
      }
      profile @required(action: THROW) {
        fullName
      }
    }
  }
`;

const signOutMutation = graphql`
  mutation ViewerMembershipDropdownSignOutMutation {
    signOut {
      success
    }
  }
`;

export function ViewerMembershipDropdown(props: {
  fKey: ViewerMembershipDropdownFragment$key;
}) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { toast } = useToast();

  const {
    viewerMembership: {
      identity: { canListAPIKeys, email },
      profile: { fullName },
    },
  } = useFragment<ViewerMembershipDropdownFragment$key>(fragment, fKey);
  const [signOut] = useMutation(signOutMutation);

  const handleLogout: React.MouseEventHandler<HTMLAnchorElement> = (e) => {
    e.preventDefault();

    signOut({
      variables: {},
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Request failed"),
            description: formatError(__("Cannot sign out"), e),
            variant: "error",
          });
          return;
        }
        window.location.href = "/auth/login";
      },
      onError: (e) => {
        toast({
          title: __("Error"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <UserDropdown fullName={fullName} email={email}>
      {canListAPIKeys && (
        <UserDropdownItem
          to="/me/api-keys"
          icon={IconKey}
          label={__("API Keys")}
        />
      )}
      <UserDropdownItem
        to={`/organizations/${organizationId}/employee`}
        icon={IconPageTextLine}
        label={__("My Signatures")}
      />
      <UserDropdownItem
        to="mailto:support@getprobo.com"
        icon={IconCircleQuestionmark}
        label={__("Help")}
      />
      <DropdownSeparator />
      <UserDropdownItem
        variant="danger"
        to="/logout"
        icon={IconArrowBoxLeft}
        label="Logout"
        onClick={handleLogout}
      />
    </UserDropdown>
  );
}
