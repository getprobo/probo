import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  DropdownSeparator,
  IconArrowBoxLeft,
  IconCircleQuestionmark,
  IconKey,
  UserDropdown,
  UserDropdownItem,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { ViewerDropdownFragment$key } from "#/__generated__/iam/ViewerDropdownFragment.graphql";

export const fragment = graphql`
  fragment ViewerDropdownFragment on Identity {
    canListAPIKeys: permission(action: "iam:personal-api-key:list")
    email
    fullName
  }
`;

const signOutMutation = graphql`
  mutation ViewerDropdownSignOutMutation {
    signOut {
      success
    }
  }
`;

export function ViewerDropdown(props: { fKey: ViewerDropdownFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const { canListAPIKeys, email, fullName }
    = useFragment<ViewerDropdownFragment$key>(fragment, fKey);
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
        window.location.reload();
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
