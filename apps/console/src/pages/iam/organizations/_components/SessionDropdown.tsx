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
import { use } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";
import { graphql } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useFragment, useMutation } from "react-relay";
import { useTranslate } from "@probo/i18n";
import type { SessionDropdownFragment$key } from "./__generated__/SessionDropdownFragment.graphql";

export const fragment = graphql`
  fragment SessionDropdownFragment on Identity
  @argumentDefinitions(organizationId: { type: "ID!" }) {
    email
    profileFor(organizationId: $organizationId) {
      firstName
      lastName
    }
  }
`;

const signOutMutation = graphql`
  mutation SessionDropdownSignOutMutation {
    signOut {
      success
    }
  }
`;

export function SessionDropdown(props: { fKey: SessionDropdownFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { isAuthorized } = use(PermissionsContext);
  const { toast } = useToast();

  const { email, profileFor: profile } =
    useFragment<SessionDropdownFragment$key>(fragment, fKey);
  const [signOut] = useMutation(signOutMutation);

  const handleLogout: React.MouseEventHandler<HTMLAnchorElement> = (e) => {
    e.preventDefault();

    signOut({
      variables: {},
      onCompleted: () => {
        window.location.reload();
      },
      onError: (e) => {
        toast({
          title: __("Error"),
          description: e.message as string,
          variant: "error",
        });
      },
    });
  };

  // FIXME: user should always have a default profile
  const fullName = profile?.firstName
    ? `${profile.firstName} ${profile.lastName}`
    : email;

  return (
    <UserDropdown fullName={fullName} email={email}>
      {isAuthorized("Organization", "deleteOrganization") && (
        <UserDropdownItem
          to="/api-keys"
          icon={IconKey}
          label={__("API Keys")}
        />
      )}
      {isAuthorized("Organization", "listSignableDocuments") && (
        <UserDropdownItem
          to={`/organizations/${organizationId}/employee`}
          icon={IconPageTextLine}
          label={__("My Signatures")}
        />
      )}
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
