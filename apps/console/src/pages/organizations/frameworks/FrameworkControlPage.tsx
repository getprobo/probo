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

import {
  formatError,
  promisifyMutation,
} from "@probo/helpers";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  useConfirm,
} from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  useMutation,
  type UseMutationConfig,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate, useOutletContext } from "react-router";
import { graphql, type MutationParameters } from "relay-runtime";

import type { FrameworkDetailPageFragment$data } from "#/__generated__/core/FrameworkDetailPageFragment.graphql";
import type { FrameworkGraphControlNodeQuery } from "#/__generated__/core/FrameworkGraphControlNodeQuery.graphql";
import { LinkedAuditsCard } from "#/components/audits/LinkedAuditsCard";
import { LinkedDocumentsCard } from "#/components/documents/LinkedDocumentsCard";
import { LinkedInternalControlsCard } from "#/components/internal-controls/LinkedInternalControlsCard";
import { LinkedObligationsCard } from "#/components/obligations/LinkedObligationsCard";
import { frameworkControlNodeQuery } from "#/hooks/graph/FrameworkGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";

const attachInternalControlMutation = graphql`
  mutation FrameworkControlPageAttachMutation(
      $input: CreateControlInternalControlMappingInput!
      $connections: [ID!]!
  ) {
      createControlInternalControlMapping(input: $input) {
          internalControlEdge @prependEdge(connections: $connections) {
              node {
                  id
                  ...LinkedInternalControlsCardFragment
              }
          }
      }
  }
`;

const detachInternalControlMutation = graphql`
  mutation FrameworkControlPageDetachMutation(
      $input: DeleteControlInternalControlMappingInput!
      $connections: [ID!]!
  ) {
      deleteControlInternalControlMapping(input: $input) {
          deletedInternalControlId @deleteEdge(connections: $connections)
      }
  }
`;

const attachDocumentMutation = graphql`
  mutation FrameworkControlPageAttachDocumentMutation(
      $input: CreateControlDocumentMappingInput!
      $connections: [ID!]!
  ) {
      createControlDocumentMapping(input: $input) {
          documentEdge @prependEdge(connections: $connections) {
              node {
                  id
                  ...LinkedDocumentsCardFragment
              }
          }
      }
  }
`;

const detachDocumentMutation = graphql`
  mutation FrameworkControlPageDetachDocumentMutation(
      $input: DeleteControlDocumentMappingInput!
      $connections: [ID!]!
  ) {
      deleteControlDocumentMapping(input: $input) {
          deletedDocumentId @deleteEdge(connections: $connections)
      }
  }
`;

const attachAuditMutation = graphql`
  mutation FrameworkControlPageAttachAuditMutation(
      $input: CreateControlAuditMappingInput!
      $connections: [ID!]!
  ) {
      createControlAuditMapping(input: $input) {
          auditEdge @prependEdge(connections: $connections) {
              node {
                  id
                  ...LinkedAuditsCardFragment
              }
          }
      }
  }
`;

const detachAuditMutation = graphql`
  mutation FrameworkControlPageDetachAuditMutation(
      $input: DeleteControlAuditMappingInput!
      $connections: [ID!]!
  ) {
      deleteControlAuditMapping(input: $input) {
          deletedAuditId @deleteEdge(connections: $connections)
      }
  }
`;

const attachObligationMutation = graphql`
  mutation FrameworkControlPageAttachObligationMutation(
      $input: CreateControlObligationMappingInput!
      $connections: [ID!]!
  ) {
      createControlObligationMapping(input: $input) {
          obligationEdge @prependEdge(connections: $connections) {
              node {
                  id
                  ...LinkedObligationsCardFragment
              }
          }
      }
  }
`;

const detachObligationMutation = graphql`
  mutation FrameworkControlPageDetachObligationMutation(
      $input: DeleteControlObligationMappingInput!
      $connections: [ID!]!
  ) {
      deleteControlObligationMapping(input: $input) {
          deletedObligationId @deleteEdge(connections: $connections)
      }
  }
`;

const deleteControlMutation = graphql`
  mutation FrameworkControlPageDeleteControlMutation(
      $input: DeleteControlInput!
      $connections: [ID!]!
  ) {
      deleteControl(input: $input) {
          deletedControlId @deleteEdge(connections: $connections)
      }
  }
`;

type Props = {
  queryRef: PreloadedQuery<FrameworkGraphControlNodeQuery>;
};

/**
* Display the control detail on the right panel
*/
export default function FrameworkControlPage({ queryRef }: Props) {
  const { t } = useTranslation();
  const toast = useToast();
  const { framework } = useOutletContext<{
    framework: FrameworkDetailPageFragment$data;
  }>();
  const connectionId = framework.controls.__id;
  const control = usePreloadedQuery<FrameworkGraphControlNodeQuery>(frameworkControlNodeQuery, queryRef).node;
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const navigate = useNavigate();
  // eslint-disable-next-line relay/generated-typescript-types
  const [detachInternalControl, isDetachingInternalControl] = useMutation(
    detachInternalControlMutation,
  );
  // eslint-disable-next-line relay/generated-typescript-types
  const [attachInternalControl, isAttachingInternalControl] = useMutation(
    attachInternalControlMutation,
  );
  // eslint-disable-next-line relay/generated-typescript-types
  const [detachDocument, isDetachingDocument] = useMutation(
    detachDocumentMutation,
  );
  // eslint-disable-next-line relay/generated-typescript-types
  const [attachDocument, isAttachingDocument] = useMutation(
    attachDocumentMutation,
  );
  // eslint-disable-next-line relay/generated-typescript-types
  const [detachAudit, isDetachingAudit] = useMutation(detachAuditMutation);
  // eslint-disable-next-line relay/generated-typescript-types
  const [attachAudit, isAttachingAudit] = useMutation(attachAuditMutation);
  // eslint-disable-next-line relay/generated-typescript-types
  const [deleteControl] = useMutation(deleteControlMutation);

  // eslint-disable-next-line relay/generated-typescript-types
  const [attachObligation, isAttachingObligation] = useMutation(
    attachObligationMutation,
  );
  // eslint-disable-next-line relay/generated-typescript-types
  const [detachObligation, isDetachingObligation] = useMutation(
    detachObligationMutation,
  );

  const canLinkInternalControl = control.canCreateInternalControlMapping;
  const canUnlinkInternalControl = control.canDeleteInternalControlMapping;
  const internalControlsReadOnly = !canLinkInternalControl && !canUnlinkInternalControl;

  const canLinkDocument = control.canCreateDocumentMapping;
  const canUnlinkDocument = control.canDeleteDocumentMapping;
  const documentsReadOnly = !canLinkDocument && !canUnlinkDocument;

  const canLinkAudit = control.canCreateAuditMapping;
  const canUnlinkAudit = control.canDeleteAuditMapping;
  const auditsReadOnly = !canLinkAudit && !canUnlinkAudit;

  const canLinkObligation = control.canCreateObligationMapping;
  const canUnlinkObligation = control.canDeleteObligationMapping;
  const obligationsReadOnly = !canLinkObligation && !canUnlinkObligation;

  const withErrorHandling
    = <T extends MutationParameters>(
      mutationFn: (config: UseMutationConfig<T>) => void,
      errorMessage: string,
    ) =>
      (options: UseMutationConfig<T>) => {
        mutationFn({
          ...options,
          onCompleted: (response, error) => {
            if (error) {
              toast.add({
                title: t("frameworkControlPage.messages.error"),
                description: formatError(
                  errorMessage,
                  error,
                ),
                type: "error",
              });
            }
            options.onCompleted?.(response, error);
          },
          onError: (error) => {
            toast.add({
              title: t("frameworkControlPage.messages.error"),
              description: formatError(
                errorMessage,
                error,
              ),
              type: "error",
            });
            options.onError?.(error);
          },
        });
      };

  const onDelete = () => {
    confirm(
      () => {
        return promisifyMutation(deleteControl)({
          variables: {
            input: {
              controlId: control.id,
            },
            connections: [connectionId],
          },
          onCompleted: () => {
            void navigate(
              `/organizations/${organizationId}/governance/frameworks/${framework.id}`,
            );
          },
        });
      },
      {
        message: t("frameworkControlPage.deleteConfirmation"),
      },
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between">
        <div className="flex items-center gap-3">
          <div className="text-xl font-medium px-[6px] py-[2px] border border-border-low rounded-lg w-max bg-active mb-3">
            {control.sectionTitle}
          </div>
        </div>
        <div className="flex gap-2">
          {control.canUpdate && (
            <FrameworkControlDialog
              frameworkId={framework.id}
              connectionId={connectionId}
              control={control}
            >
              <Button icon={IconPencil} variant="secondary">
                {t("frameworkControlPage.actions.editControl")}
              </Button>
            </FrameworkControlDialog>
          )}
          {control.canDelete && (
            <ActionDropdown variant="secondary">
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onClick={onDelete}
              >
                {t("frameworkControlPage.actions.delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </div>
      </div>

      <div>
        <div className="text-base mb-1">{control.name}</div>
        {control.description && (
          <div className="text-sm text-txt-secondary mb-4 whitespace-pre-wrap">
            {control.description}
          </div>
        )}
        <Card padded className="mb-6 mt-6">
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="text-sm text-txt-secondary">{t("frameworkControlPage.fields.bestPractice")}</span>
              <Badge variant={control.bestPractice ? "success" : "neutral"} size="sm">
                {control.bestPractice ? t("frameworkControlPage.answers.yes") : t("frameworkControlPage.answers.no")}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-txt-secondary">{t("frameworkControlPage.fields.maturityLevel")}</span>
              <Badge variant="neutral" size="sm">
                {t(`frameworkControlPage.maturityLevels.${(control.maturityLevel ?? "NONE").toLowerCase()}`)}
              </Badge>
            </div>
            {control.maturityLevel === "NONE" && control.notImplementedJustification && (
              <div>
                <span className="text-xs text-txt-secondary">{t("frameworkControlPage.fields.notImplementedJustification")}</span>
                <div className="text-sm mt-0.5 whitespace-pre-wrap">{control.notImplementedJustification}</div>
              </div>
            )}
          </div>
        </Card>
        <div className="mb-4">
          <LinkedInternalControlsCard
            variant="card"
            internalControls={
              control.internalControls?.edges.map(edge => edge.node)
              ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.internalControls?.__id ?? ""}
            onAttach={withErrorHandling(
              attachInternalControl,
              t("frameworkControlPage.errors.linkInternalControl"),
            )}
            onDetach={withErrorHandling(
              detachInternalControl,
              t("frameworkControlPage.errors.unlinkInternalControl"),
            )}
            disabled={isAttachingInternalControl || isDetachingInternalControl}
            readOnly={internalControlsReadOnly}
          />
        </div>
        <div className="mb-4">
          <LinkedDocumentsCard
            variant="card"
            documents={
              control.documents?.edges.map(edge => edge.node)
              ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.documents?.__id ?? ""}
            onAttach={withErrorHandling(
              attachDocument,
              t("frameworkControlPage.errors.linkDocument"),
            )}
            onDetach={withErrorHandling(
              detachDocument,
              t("frameworkControlPage.errors.unlinkDocument"),
            )}
            disabled={isAttachingDocument || isDetachingDocument}
            readOnly={documentsReadOnly}
          />
        </div>
        <div className="mb-4">
          <LinkedAuditsCard
            variant="card"
            audits={
              control.audits?.edges.map(edge => edge.node) ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.audits?.__id ?? ""}
            onAttach={withErrorHandling(
              attachAudit,
              t("frameworkControlPage.errors.linkAudit"),
            )}
            onDetach={withErrorHandling(
              detachAudit,
              t("frameworkControlPage.errors.unlinkAudit"),
            )}
            disabled={isAttachingAudit || isDetachingAudit}
            readOnly={auditsReadOnly}
          />
        </div>
        <div className="mb-4">
          <LinkedObligationsCard
            variant="card"
            obligations={
              control.obligations?.edges.map(
                edge => edge.node,
              ) ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.obligations?.__id ?? ""}
            onAttach={withErrorHandling(
              attachObligation,
              t("frameworkControlPage.errors.linkObligation"),
            )}
            onDetach={withErrorHandling(
              detachObligation,
              t("frameworkControlPage.errors.unlinkObligation"),
            )}
            disabled={
              isAttachingObligation || isDetachingObligation
            }
            readOnly={obligationsReadOnly}
          />
        </div>
      </div>
    </div>
  );
}
