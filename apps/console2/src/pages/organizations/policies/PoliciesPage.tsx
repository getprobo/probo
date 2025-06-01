import { useTranslate } from "@probo/i18n";
import {
  PageHeader,
  Table,
  Tbody,
  Thead,
  Tr,
  Th,
  Td,
  Avatar,
  Badge,
  IconTrashCan,
  Button,
  IconPlusLarge,
  useConfirm,
  ActionDropdown,
  DropdownItem,
} from "@probo/ui";
import {
  useFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import type { PolicyGraphListQuery } from "/hooks/graph/__generated__/PolicyGraphListQuery.graphql";
import {
  policiesQuery,
  useDeletePolicyMutation,
} from "/hooks/graph/PolicyGraph";
import type { PoliciesPageListFragment$key } from "./__generated__/PoliciesPageListFragment.graphql";
import { usePageTitle } from "@probo/hooks";
import { sprintf } from "@probo/helpers";
import { CreatePolicyDialog } from "./dialogs/CreatePolicyDialog";
import type { PoliciesPageRowFragment$key } from "./__generated__/PoliciesPageRowFragment.graphql";

const policiesFragment = graphql`
  fragment PoliciesPageListFragment on Organization {
    policies(first: 100) @connection(key: "PoliciesPageFragment_policies") {
      __id
      edges {
        node {
          id
          ...PoliciesPageRowFragment
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<PolicyGraphListQuery>;
};

export default function PoliciesPage(props: Props) {
  const { __ } = useTranslate();

  const organization = usePreloadedQuery(
    policiesQuery,
    props.queryRef
  ).organization;
  const data = useFragment<PoliciesPageListFragment$key>(
    policiesFragment,
    organization
  );

  const policies = data.policies.edges.map((edge) => edge.node);
  const connectionId = data.policies.__id;
  usePageTitle(__("Policies"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Policies")}
        description={__("Manage your organization’s policies")}
      >
        <CreatePolicyDialog
          connection={connectionId}
          trigger={<Button icon={IconPlusLarge}>{__("New policy")}</Button>}
        />
      </PageHeader>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Status")}</Th>
            <Th>{__("Owner")}</Th>
            <Th>{__("Last update")}</Th>
            <Th>{__("Signatures")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {policies.map((policy) => (
            <PolicyRow
              key={policy.id}
              policy={policy}
              organizationId={organization.id}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}

const rowFragment = graphql`
  fragment PoliciesPageRowFragment on Policy {
    id
    title
    description
    updatedAt
    owner {
      id
      fullName
    }
    versions(first: 1) {
      edges {
        node {
          id
          status
          signatures(first: 100) {
            edges {
              node {
                id
                state
              }
            }
          }
        }
      }
    }
  }
`;

function PolicyRow({
  policy: policyKey,
  organizationId,
  connectionId,
}: {
  policy: PoliciesPageRowFragment$key;
  organizationId: string;
  connectionId: string;
}) {
  const policy = useFragment<PoliciesPageRowFragment$key>(
    rowFragment,
    policyKey
  );
  const lastVersion = policy.versions.edges[0].node;
  const isDraft = lastVersion.status === "DRAFT";
  const { __, dateFormat } = useTranslate();
  const signatures = lastVersion.signatures.edges.map((edge) => edge.node);
  const signedCount = signatures.filter(
    (signature) => signature.state === "SIGNED"
  ).length;
  const [deletePolicy, isDeleting] = useDeletePolicyMutation();
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        deletePolicy({
          variables: {
            input: { policyId: policy.id },
            connections: [connectionId],
          },
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

  return (
    <Tr to={`/organizations/${organizationId}/policies/${policy.id}`}>
      <Td>
        <div className="flex gap-4 items-center">
          <img
            src="/policy.png"
            alt=""
            width={28}
            height={36}
            className="border-4 border-highlight rounded box-content"
          />
          {policy.title}
        </div>
      </Td>
      <Td>
        <Badge variant={isDraft ? "neutral" : "success"}>
          {isDraft ? __("Draft") : __("Published")}
        </Badge>
      </Td>
      <Td>
        <div className="flex gap-2 items-center">
          <Avatar name={policy.owner.fullName} />
          {policy.owner.fullName}
        </div>
      </Td>
      <Td>
        {dateFormat(policy.updatedAt, {
          year: "numeric",
          month: "short",
          day: "numeric",
          weekday: "short",
        })}
      </Td>
      <Td>
        {signedCount}/{signatures.length}
      </Td>
      <Td noLink width={50} className="text-end">
        <ActionDropdown>
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            onClick={handleDelete}
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </Td>
    </Tr>
  );
}
