import type { PreloadedQuery } from "react-relay";
import type { PolicyGraphNodeQuery } from "/hooks/graph/__generated__/PolicyGraphNodeQuery.graphql";
import {
  ConnectionHandler,
  graphql,
  loadQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import {
  PoliciesConnectionKey,
  policyNodeQuery,
  useDeletePolicyMutation,
} from "/hooks/graph/PolicyGraph";
import { usePageTitle } from "@probo/hooks";
import type { PolicyPagePolicyFragment$key } from "./__generated__/PolicyPagePolicyFragment.graphql";
import { useTranslate } from "@probo/i18n";
import {
  PageHeader,
  Breadcrumb,
  IconCheckmark1,
  Markdown,
  PropertyRow,
  Drawer,
  Badge,
  Avatar,
  DropdownItem,
  ActionDropdown,
  IconTrashCan,
  IconPencil,
  IconClock,
  IconSignature,
  useConfirm,
} from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Button } from "@probo/ui";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { sprintf } from "@probo/helpers";
import { useNavigate } from "react-router";
import UpdateVersionDialog from "./dialogs/UpdateVersionDialog";
import { useRef } from "react";
import { PolicyVersionHistoryDialog } from "./dialogs/PolicyVersionHistoryDialog";
import { PolicySignaturesDialog } from "./dialogs/PolicySignaturesDialog";

type Props = {
  queryRef: PreloadedQuery<PolicyGraphNodeQuery>;
};

const policyFragment = graphql`
  fragment PolicyPagePolicyFragment on Policy {
    id
    title
    owner {
      id
      fullName
    }
    versions(first: 20) @connection(key: "PolicyPage_versions") {
      __id
      edges {
        node {
          id
          content
          status
          publishedAt
          version
          updatedAt
          signatures(first: 100) @connection(key: "PolicyPage_signatures") {
            __id
            edges {
              node {
                id
                state
                signedBy {
                  id
                }
                ...PolicySignaturesDialog_signature
              }
            }
          }
          ...PolicyVersionHistoryDialogFragment
          ...PolicySignaturesDialog_version
        }
      }
    }
  }
`;

const publishPolicyVersionMutation = graphql`
  mutation PolicyPagePublishMutation($input: PublishPolicyVersionInput!) {
    publishPolicyVersion(input: $input) {
      policy {
        id
      }
    }
  }
`;

export default function PolicyPage(props: Props) {
  const node = usePreloadedQuery(policyNodeQuery, props.queryRef).node;
  const policy = useFragment(
    policyFragment,
    node as PolicyPagePolicyFragment$key
  );
  const { __, dateFormat } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const lastVersion = policy.versions.edges[0].node;
  const isDraft = lastVersion.status === "DRAFT";
  const [publishPolicyVersion, isPublishing] = useMutationWithToasts(
    publishPolicyVersionMutation,
    {
      successMessage: __("Policy published successfully."),
      errorMessage: __("Failed to publish policy. Please try again."),
    }
  );
  const [deletePolicy, isDeleting] = useDeletePolicyMutation();
  const versionConnectionId = policy.versions.__id;

  usePageTitle(policy.title);

  const handlePublish = () => {
    publishPolicyVersion({
      variables: {
        input: { policyId: policy.id },
      },
      onSuccess: () => {
        // Refresh the whole query to get the new version
        loadQuery(
          props.queryRef.environment,
          policyNodeQuery,
          props.queryRef.variables,
          { fetchPolicy: "network-only" }
        );
      },
    });
  };

  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          const connectionId = ConnectionHandler.getConnectionID(
            organizationId,
            PoliciesConnectionKey
          );
          deletePolicy({
            variables: {
              input: { policyId: policy.id },
              connections: [connectionId],
            },
            onSuccess() {
              navigate(`/organizations/${organizationId}/policies`);
            },
            onError: () => resolve(),
          });
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the policy "%s". This action cannot be undone.'
          ),
          policy.title
        ),
      }
    );
  };

  const updateDialogRef = useRef<{ open: () => void }>(null);

  return (
    <>
      <UpdateVersionDialog
        ref={updateDialogRef}
        policy={policy}
        connectionId={versionConnectionId}
      />
      <div className="space-y-6">
        <div className="flex justify-between items-center mb-4">
          <Breadcrumb
            items={[
              {
                label: __("Policies"),
                to: `/organizations/${organizationId}/policies`,
              },
              {
                label: policy.title,
              },
            ]}
          />
          <div className="flex gap-2">
            {isDraft && (
              <Button
                onClick={handlePublish}
                icon={IconCheckmark1}
                disabled={isPublishing}
              >
                {__("Publish")}
              </Button>
            )}
            <PolicyVersionHistoryDialog policy={policy}>
              <Button icon={IconClock} variant="secondary">
                {__("Version history")}
              </Button>
            </PolicyVersionHistoryDialog>
            <PolicySignaturesDialog policy={policy}>
              <Button icon={IconSignature} variant="secondary">
                {__("Signatures history")}
              </Button>
            </PolicySignaturesDialog>

            <ActionDropdown variant="secondary">
              <DropdownItem
                onClick={() => updateDialogRef.current?.open()}
                icon={IconPencil}
              >
                {isDraft ? __("Edit draft policy") : __("Create new draft")}
              </DropdownItem>
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                disabled={isDeleting}
                onClick={handleDelete}
              >
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          </div>
        </div>
        <PageHeader title={policy.title} />
        <Markdown content={lastVersion.content} />
      </div>
      <Drawer>
        <div className="text-base text-txt-primary font-medium mb-4">
          {__("Properties")}
        </div>
        <PropertyRow label={__("Owner")}>
          <Badge variant="highlight" size="md" className="gap-2">
            <Avatar name={policy.owner?.fullName ?? ""} />
            {policy.owner?.fullName}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Status")}>
          <Badge
            variant={isDraft ? "highlight" : "success"}
            size="md"
            className="gap-2"
          >
            {isDraft ? __("Draft") : __("Published")}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Version")}>
          <div className="text-sm text-txt-secondary">
            {lastVersion.version}
          </div>
        </PropertyRow>
        <PropertyRow label={__("Last modified")}>
          <div className="text-sm text-txt-secondary">
            {dateFormat(lastVersion.updatedAt)}
          </div>
        </PropertyRow>
        {lastVersion.publishedAt && (
          <PropertyRow label={__("Published Date")}>
            <div className="text-sm text-txt-secondary">
              {dateFormat(lastVersion.publishedAt)}
            </div>
          </PropertyRow>
        )}
      </Drawer>
    </>
  );
}
