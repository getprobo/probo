import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const cookieBannersQuery = graphql`
  query CookieBannerGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateCookieBanner: permission(
          action: "core:cookie-banner:create"
        )
        ...CookieBannersPageFragment
      }
    }
  }
`;

export const cookieBannerNodeQuery = graphql`
  query CookieBannerGraphNodeQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      ... on CookieBanner {
        id
        name
        domain
        state
        title
        description
        acceptAllLabel
        rejectAllLabel
        savePreferencesLabel
        privacyPolicyUrl
        consentExpiryDays
        version
        embedSnippet
        createdAt
        updatedAt
        canUpdate: permission(action: "core:cookie-banner:update")
        canDelete: permission(action: "core:cookie-banner:delete")
        canPublish: permission(action: "core:cookie-banner:update")
        ...CookieBannerOverviewTabFragment
        ...CookieBannerAppearanceTabFragment
        ...CookieBannerCategoriesTabFragment
        ...CookieBannerConsentRecordsTabFragment
      }
    }
  }
`;

export const createCookieBannerMutation = graphql`
  mutation CookieBannerGraphCreateMutation(
    $input: CreateCookieBannerInput!
    $connections: [ID!]!
  ) {
    createCookieBanner(input: $input) {
      cookieBannerEdge @appendEdge(connections: $connections) {
        node {
          id
          name
          domain
          state
          version
          createdAt
          updatedAt
          canUpdate: permission(action: "core:cookie-banner:update")
          canDelete: permission(action: "core:cookie-banner:delete")
        }
      }
    }
  }
`;

const deleteCookieBannerMutation = graphql`
  mutation CookieBannerGraphDeleteMutation(
    $input: DeleteCookieBannerInput!
    $connections: [ID!]!
  ) {
    deleteCookieBanner(input: $input) {
      deletedCookieBannerId @deleteEdge(connections: $connections)
    }
  }
`;

const publishCookieBannerMutation = graphql`
  mutation CookieBannerGraphPublishMutation(
    $input: PublishCookieBannerInput!
  ) {
    publishCookieBanner(input: $input) {
      cookieBanner {
        id
        state
        version
        updatedAt
      }
    }
  }
`;

const disableCookieBannerMutation = graphql`
  mutation CookieBannerGraphDisableMutation(
    $input: DisableCookieBannerInput!
  ) {
    disableCookieBanner(input: $input) {
      cookieBanner {
        id
        state
        updatedAt
      }
    }
  }
`;

export const updateCookieBannerMutation = graphql`
  mutation CookieBannerGraphUpdateMutation(
    $input: UpdateCookieBannerInput!
  ) {
    updateCookieBanner(input: $input) {
      cookieBanner {
        id
        name
        domain
        state
        title
        description
        acceptAllLabel
        rejectAllLabel
        savePreferencesLabel
        privacyPolicyUrl
        consentExpiryDays
        version
        updatedAt
        theme {
          primaryColor
          primaryTextColor
          secondaryColor
          secondaryTextColor
          backgroundColor
          textColor
          secondaryTextBodyColor
          borderColor
          fontFamily
          borderRadius
          position
          revisitPosition
        }
      }
    }
  }
`;

export function useCreateCookieBannerMutation() {
  const { __ } = useTranslate();
  return useMutationWithToasts(createCookieBannerMutation, {
    successMessage: __("Cookie banner created successfully."),
    errorMessage: __("Failed to create cookie banner"),
  });
}

export function useUpdateCookieBannerMutation() {
  const { __ } = useTranslate();
  return useMutationWithToasts(updateCookieBannerMutation, {
    successMessage: __("Cookie banner updated successfully."),
    errorMessage: __("Failed to update cookie banner"),
  });
}

export function usePublishCookieBannerMutation() {
  const { __ } = useTranslate();
  return useMutationWithToasts(publishCookieBannerMutation, {
    successMessage: __("Cookie banner published successfully."),
    errorMessage: __("Failed to publish cookie banner"),
  });
}

export function useDisableCookieBannerMutation() {
  const { __ } = useTranslate();
  return useMutationWithToasts(disableCookieBannerMutation, {
    successMessage: __("Cookie banner disabled successfully."),
    errorMessage: __("Failed to disable cookie banner"),
  });
}

export const useDeleteCookieBanner = (
  banner: { id?: string; name?: string },
  connectionId: string,
) => {
  // eslint-disable-next-line relay/generated-typescript-types
  const [mutate] = useMutation(deleteCookieBannerMutation);
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return () => {
    if (!banner.id || !banner.name) {
      return alert(__("Failed to delete cookie banner: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              id: banner.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete cookie banner \"%s\". This action cannot be undone.",
          ),
          banner.name,
        ),
      },
    );
  };
};
