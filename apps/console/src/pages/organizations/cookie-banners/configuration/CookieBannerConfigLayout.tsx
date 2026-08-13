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

import { formatError } from "@probo/helpers";
import {
  ActionDropdown,
  Badge,
  Button,
  DropdownItem,
  IconSquareBehindSquare2,
  IconTrashCan,
  PageHeader,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Link, Outlet, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CookieBannerConfigLayoutActivateMutation } from "#/__generated__/core/CookieBannerConfigLayoutActivateMutation.graphql";
import type { CookieBannerConfigLayoutDeactivateMutation } from "#/__generated__/core/CookieBannerConfigLayoutDeactivateMutation.graphql";
import type { CookieBannerConfigLayoutDeleteMutation } from "#/__generated__/core/CookieBannerConfigLayoutDeleteMutation.graphql";
import type { CookieBannerConfigLayoutPublishMutation } from "#/__generated__/core/CookieBannerConfigLayoutPublishMutation.graphql";
import type { CookieBannerConfigLayoutQuery } from "#/__generated__/core/CookieBannerConfigLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const cookieBannerConfigLayoutQuery = graphql`
  query CookieBannerConfigLayoutQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        name
        origin
        state
        canDelete: permission(action: "core:cookie-banner:delete")
        latestVersion {
          id
          version
          state
        }
        policyDocument {
          id
        }
      }
    }
  }
`;

const activateMutation = graphql`
  mutation CookieBannerConfigLayoutActivateMutation($input: ActivateCookieBannerInput!) {
    activateCookieBanner(input: $input) {
      cookieBanner {
        id
        state
      }
    }
  }
`;

const deactivateMutation = graphql`
  mutation CookieBannerConfigLayoutDeactivateMutation($input: DeactivateCookieBannerInput!) {
    deactivateCookieBanner(input: $input) {
      cookieBanner {
        id
        state
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation CookieBannerConfigLayoutDeleteMutation(
    $input: DeleteCookieBannerInput!
    $connections: [ID!]!
  ) {
    deleteCookieBanner(input: $input) {
      deletedCookieBannerId @deleteEdge(connections: $connections)
    }
  }
`;

const publishMutation = graphql`
  mutation CookieBannerConfigLayoutPublishMutation($input: PublishCookieBannerVersionInput!) {
    publishCookieBannerVersion(input: $input) {
      cookieBannerVersion {
        id
        version
        state
      }
      cookieBanner {
        id
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface CookieBannerConfigLayoutProps {
  queryRef: PreloadedQuery<CookieBannerConfigLayoutQuery>;
}

export default function CookieBannerConfigLayout({ queryRef }: CookieBannerConfigLayoutProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const confirm = useConfirm();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery<CookieBannerConfigLayoutQuery>(cookieBannerConfigLayoutQuery, queryRef);
  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;

  const [activate, isActivating] = useMutation<CookieBannerConfigLayoutActivateMutation>(activateMutation);
  const [deactivate, isDeactivating] = useMutation<CookieBannerConfigLayoutDeactivateMutation>(
    deactivateMutation,
  );
  const [publish, isPublishing] = useMutation<CookieBannerConfigLayoutPublishMutation>(publishMutation);
  const [deleteCookieBanner] = useMutation<CookieBannerConfigLayoutDeleteMutation>(deleteMutation);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "CookieBannerSwitcherMenu_cookieBanners",
  );

  const handleToggleState = () => {
    if (banner.state === "ACTIVE") {
      deactivate({
        variables: { input: { cookieBannerId: banner.id } },
        onCompleted() {
          toast({ title: t("configLayout.messages.successTitle"), description: t("configLayout.messages.deactivated"), variant: "success" });
        },
        onError(error) {
          toast({ title: t("configLayout.errors.title"), description: formatError(t("configLayout.errors.deactivate"), error), variant: "error" });
        },
      });
    } else {
      activate({
        variables: { input: { cookieBannerId: banner.id } },
        onCompleted() {
          toast({ title: t("configLayout.messages.successTitle"), description: t("configLayout.messages.activated"), variant: "success" });
        },
        onError(error) {
          toast({ title: t("configLayout.errors.title"), description: formatError(t("configLayout.errors.activate"), error), variant: "error" });
        },
      });
    }
  };

  const handlePublish = () => {
    publish({
      variables: { input: { cookieBannerId: banner.id } },
      onCompleted() {
        toast({ title: t("configLayout.messages.successTitle"), description: t("configLayout.messages.published"), variant: "success" });
      },
      onError(error) {
        toast({ title: t("configLayout.errors.title"), description: formatError(t("configLayout.errors.publish"), error), variant: "error" });
      },
    });
  };

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          let nextPath = `/organizations/${organizationId}/privacy/cookie-banners/new`;
          deleteCookieBanner({
            variables: {
              input: { cookieBannerId: banner.id },
              connections: [connectionId],
            },
            updater(store) {
              const connection = store.get(connectionId);
              if (connection == null) {
                return;
              }
              const edges = connection.getLinkedRecords("edges") ?? [];
              for (const edge of edges) {
                const id = edge?.getLinkedRecord("node")?.getDataID();
                if (typeof id === "string" && id !== banner.id) {
                  nextPath = `/organizations/${organizationId}/privacy/cookie-banners/${id}/configure`;
                  return;
                }
              }
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: t("configLayout.errors.title"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: t("configLayout.messages.successTitle"),
                  description: t("configLayout.messages.deleted"),
                  variant: "success",
                });
                void navigate(nextPath);
              }
              resolve();
            },
            onError(error) {
              toast({
                title: t("configLayout.errors.title"),
                description: formatError(t("configLayout.errors.delete"), error),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("configLayout.deleteConfirmation", { name: banner.name }),
        variant: "danger",
        label: t("configLayout.actions.delete"),
      },
    );
  };

  const hasDraft = banner.latestVersion?.state === "DRAFT";

  return (
    <div className="space-y-6">
      <PageHeader
        title={(
          <div className="align-baseline">
            {banner.name}
            {banner.latestVersion?.version != null && (
              <span className="font-mono text-base text-txt-secondary ml-2">
                {t("configLayout.version", { version: banner.latestVersion.version })}
                {banner.latestVersion.state === "DRAFT" && (
                  <span className="text-xs font-sans">
                    {t("configLayout.draft")}
                  </span>
                )}
              </span>
            )}
          </div>
        )}
        description={(
          <span className="flex items-center gap-3 text-sm text-txt-secondary">
            <span>
              <span className="font-medium text-txt-primary">{t("configLayout.metadata.origin")}</span>
              {" "}
              {banner.origin}
            </span>
            <span className="text-border-primary">·</span>
            <span className="flex items-center gap-1">
              <span className="font-medium text-txt-primary">{t("configLayout.metadata.id")}</span>
              {" "}
              {banner.id}
              <button
                type="button"
                className="p-1 rounded hover:bg-bg-hover transition-colors cursor-pointer"
                onClick={() => {
                  void navigator.clipboard.writeText(banner.id);
                  toast({ title: t("configLayout.messages.copiedTitle"), description: t("configLayout.messages.idCopied"), variant: "success" });
                }}
              >
                <IconSquareBehindSquare2 size={16} />
              </button>
            </span>
            {banner.policyDocument && (
              <>
                <span className="text-border-primary">·</span>
                <Link
                  to={`/organizations/${organizationId}/governance/documents/${banner.policyDocument.id}`}
                  className="font-medium text-txt-primary underline"
                >
                  {t("configLayout.metadata.cookiePolicy")}
                </Link>
              </>
            )}
          </span>
        )}
      >
        <Badge variant={banner.state === "ACTIVE" ? "success" : "danger"}>
          {banner.state === "ACTIVE" ? t("configLayout.status.active") : t("configLayout.status.inactive")}
        </Badge>
        {hasDraft && (
          <Button onClick={handlePublish} disabled={isPublishing}>
            {isPublishing ? t("configLayout.actions.publishing") : t("configLayout.actions.publish")}
          </Button>
        )}
        <Button
          variant="secondary"
          onClick={handleToggleState}
          disabled={isActivating || isDeactivating}
        >
          {banner.state === "ACTIVE" ? t("configLayout.actions.deactivate") : t("configLayout.actions.activate")}
        </Button>
        {banner.canDelete && banner.state !== "ACTIVE" && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {t("configLayout.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <Outlet />
    </div>
  );
}
