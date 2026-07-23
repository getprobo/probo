// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { usePageTitle } from "@probo/hooks";
import {
  ActionDropdown,
  Button,
  ControlItem,
  DropdownItem,
  FrameworkLogo,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { Navigate, Outlet, useNavigate, useParams } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { FrameworkDetailPageExportFrameworkMutation } from "#/__generated__/core/FrameworkDetailPageExportFrameworkMutation.graphql";
import type { FrameworkDetailPageFragment$key } from "#/__generated__/core/FrameworkDetailPageFragment.graphql";
import type { FrameworkGraphNodeQuery } from "#/__generated__/core/FrameworkGraphNodeQuery.graphql";
import {
  connectionListKey,
  frameworkNodeQuery,
  useDeleteFrameworkMutation,
} from "#/hooks/graph/FrameworkGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";
import { FrameworkFormDialog } from "./dialogs/FrameworkFormDialog";

const frameworkDetailFragment = graphql`
    fragment FrameworkDetailPageFragment on Framework {
        id
        name
        # eslint-disable-next-line relay/unused-fields
        description
        lightLogo {
          downloadUrl
        }
        # eslint-disable-next-line relay/unused-fields
        darkLogo {
          downloadUrl
        }
        canExport: permission(action: "core:franework:export")
        canUpdate: permission(action: "core:framework:update")
        canDelete: permission(action: "core:framework:delete")
        canCreateControl: permission(action: "core:control:create")
        controls(
            first: 250
            orderBy: { field: SECTION_TITLE, direction: ASC }
        ) {
            __id
            edges {
                node {
                    id
                    sectionTitle
                    name
                }
            }
        }
    }
`;

const exportFrameworkMutation = graphql`
    mutation FrameworkDetailPageExportFrameworkMutation($frameworkId: ID!) {
        exportFramework(input: { frameworkId: $frameworkId }) {
            exportJobId
        }
    }
`;

type Props = {
  queryRef: PreloadedQuery<FrameworkGraphNodeQuery>;
};

export default function FrameworkDetailPage(props: Props) {
  const { queryRef } = props;

  const { t } = useTranslation();
  const { controlId } = useParams<{ controlId?: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<FrameworkGraphNodeQuery>(
    frameworkNodeQuery,
    queryRef,
  );
  const framework = useFragment<FrameworkDetailPageFragment$key>(
    frameworkDetailFragment,
    data.node,
  );
  const navigate = useNavigate();
  const controls = framework.controls.edges.map(edge => edge.node);
  const selectedControl = controlId
    ? controls.find(control => control.id === controlId)
    : controls[0] || null;
  const connectionId = framework.controls.__id;
  const deleteFramework = useDeleteFrameworkMutation(
    framework,
    ConnectionHandler.getConnectionID(organizationId, connectionListKey),
  );
  const [exportFramework]
    = useMutationWithToasts<FrameworkDetailPageExportFrameworkMutation>(
      exportFrameworkMutation,
      {
        errorMessage: t("frameworkDetailPage.errors.export"),
        successMessage: t("frameworkDetailPage.messages.exportStarted"),
      },
    );

  usePageTitle(`${framework.name} | ${selectedControl?.sectionTitle}`);
  const onDelete = () => {
    deleteFramework({
      onSuccess: () => {
        void navigate(`/organizations/${organizationId}/frameworks`);
      },
    });
  };

  if (!controlId && controls.length > 0) {
    return (
      <Navigate
        to={`/organizations/${organizationId}/frameworks/${framework.id}/controls/${controls[0].id}`}
      />
    );
  }

  const hasAnyAction = framework.canExport || framework.canDelete;

  return (
    <div className="space-y-6">
      <PageHeader
        title={(
          <>
            <FrameworkLogo
              name={framework.name}
              lightLogoURL={framework.lightLogo?.downloadUrl}
              darkLogoURL={framework.darkLogo?.downloadUrl}
            />
            {framework.name}
          </>
        )}
      >
        {framework.canUpdate && (
          <FrameworkFormDialog
            organizationId={organizationId}
            framework={framework}
          >
            <Button icon={IconPencil} variant="secondary">
              {t("frameworkDetailPage.actions.edit")}
            </Button>
          </FrameworkFormDialog>
        )}
        {hasAnyAction && (
          <ActionDropdown variant="secondary">
            {framework.canExport
              && (
                <DropdownItem
                  variant="primary"
                  onClick={() => {
                    void exportFramework({
                      variables: { frameworkId: framework.id },
                    });
                  }}
                >
                  {t("frameworkDetailPage.actions.export")}
                </DropdownItem>
              )}
            {framework.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onClick={onDelete}
              >
                {t("frameworkDetailPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        )}
      </PageHeader>
      <div className="text-lg font-semibold">
        {t("frameworkDetailPage.requirementCategories")}
      </div>
      <div className="divide-x divide-border-low grid grid-cols-[264px_1fr]">
        <div
          className="space-y-1 overflow-y-auto pr-6 mr-6 sticky top-0"
          style={{ maxHeight: "calc(100vh - 48px)" }}
        >
          {controls.map(control => (
            <ControlItem
              key={control.id}
              id={control.sectionTitle}
              description={control.name}
              to={`/organizations/${organizationId}/frameworks/${framework.id}/controls/${control.id}`}
              active={selectedControl?.id === control.id}
            />
          ))}
          {framework.canCreateControl && (
            <FrameworkControlDialog
              frameworkId={framework.id}
              connectionId={connectionId}
            >
              <button className="flex gap-[6px] flex-col w-full p-4 space-y-[6px] rounded-xl cursor-pointer text-start text-sm text-txt-tertiary hover:bg-tertiary-hover">
                <IconPlusLarge
                  size={20}
                  className="text-txt-primary"
                />
                {t("frameworkDetailPage.actions.addControl")}
              </button>
            </FrameworkControlDialog>
          )}
        </div>
        <Outlet context={{ framework }} />
      </div>
    </div>
  );
}
