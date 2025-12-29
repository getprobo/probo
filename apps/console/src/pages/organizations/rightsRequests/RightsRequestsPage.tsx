import {
  Button,
  IconPlusLarge,
  PageHeader,
  Card,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Table,
  useConfirm,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  ConnectionHandler,
  graphql,
  usePaginationFragment,
  usePreloadedQuery,
  useMutation,
  type PreloadedQuery,
} from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { CreateRightsRequestDialog } from "./dialogs/CreateRightsRequestDialog";
import { deleteRightsRequestMutation, RightsRequestsConnectionKey, rightsRequestsQuery } from "../../../hooks/graph/RightsRequestGraph";
import {
  sprintf,
  promisifyMutation,
  formatDate,
  getRightsRequestTypeLabel,
  getRightsRequestStateVariant,
  getRightsRequestStateLabel,
} from "@probo/helpers";
import type { NodeOf } from "/types";
import type {
  RightsRequestsPageFragment$key,
  RightsRequestsPageFragment$data,
} from "./__generated__/RightsRequestsPageFragment.graphql";
import { use } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";
import type { RightsRequestGraphListQuery } from "/hooks/graph/__generated__/RightsRequestGraphListQuery.graphql";

interface RightsRequestsPageProps {
  queryRef: PreloadedQuery<RightsRequestGraphListQuery>;
}

const rightsRequestsPageFragment = graphql`
  fragment RightsRequestsPageFragment on Organization
  @refetchable(queryName: "RightsRequestsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
  ) {
    id
    rightsRequests(
      first: $first
      after: $after
    )
      @connection(key: "RightsRequestsPage_rightsRequests") {
      __id
      totalCount
      edges {
        node {
          id
          requestType
          requestState
          dataSubject
          contact
          details
          deadline
          actionTaken
          createdAt
          updatedAt
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export default function RightsRequestsPage({ queryRef }: RightsRequestsPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { isAuthorized } = use(PermissionsContext);

  usePageTitle(__("Rights Requests"));

  const organization = usePreloadedQuery(
    rightsRequestsQuery,
    queryRef
  );

  const {
    data,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    RightsRequestGraphListQuery,
    RightsRequestsPageFragment$key
  >(rightsRequestsPageFragment, organization.node);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    RightsRequestsConnectionKey
  );
  const requests = data?.rightsRequests?.edges?.map((edge) => edge.node) ?? [];

  const hasAnyAction = (
    isAuthorized("RightsRequest", "updateRightsRequest") ||
    isAuthorized("RightsRequest", "deleteRightsRequest")
  );

  return (
    <div className="space-y-6">
      <PageHeader title={__("Rights Requests")} description={__("Manage data subject rights requests.")}>
        {isAuthorized("Organization", "createRightsRequest") && (
          <CreateRightsRequestDialog
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {__("Add rights request")}
            </Button>
          </CreateRightsRequestDialog>
        )}
      </PageHeader>

      {requests.length > 0 ? (
        <Card>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Type")}</Th>
                <Th>{__("State")}</Th>
                <Th>{__("Data Subject")}</Th>
                <Th>{__("Contact")}</Th>
                <Th>{__("Deadline")}</Th>
                {hasAnyAction && <Th>{__("Actions")}</Th>}
              </Tr>
            </Thead>
            <Tbody>
              {requests.map((request) => (
                <RequestRow
                  key={request.id}
                  request={request}
                  connectionId={connectionId}
                  hasAnyAction={hasAnyAction}
                />
              ))}
            </Tbody>
          </Table>

          {hasNext && (
            <div className="p-4 border-t">
              <Button
                variant="secondary"
                onClick={() => loadNext(10)}
                disabled={isLoadingNext}
              >
                {isLoadingNext ? __("Loading...") : __("Load more")}
              </Button>
            </div>
          )}
        </Card>
      ) : (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No rights requests yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first rights request to get started.")}
            </p>
          </div>
        </Card>
      )}
    </div>
  );
}

function RequestRow({
  request,
  connectionId,
  hasAnyAction,
}: {
  request: NodeOf<NonNullable<RightsRequestsPageFragment$data['rightsRequests']>>;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [deleteRequest] = useMutation(deleteRightsRequestMutation);
  const confirm = useConfirm();
  const { isAuthorized } = use(PermissionsContext);

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteRequest)({
          variables: {
            input: {
              rightsRequestId: request.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the rights request. This action cannot be undone."
          )
        ),
      }
    );
  };


  const detailsUrl = `/organizations/${organizationId}/rights-requests/${request.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>
        <Badge variant="neutral">{getRightsRequestTypeLabel(__, request.requestType)}</Badge>
      </Td>
      <Td>
        <Badge variant={getRightsRequestStateVariant(request.requestState)}>
          {getRightsRequestStateLabel(__, request.requestState)}
        </Badge>
      </Td>
      <Td>{request.dataSubject || "-"}</Td>
      <Td>{request.contact || "-"}</Td>
      <Td>
        {request.deadline ? (
          <time dateTime={request.deadline}>
            {formatDate(request.deadline)}
          </time>
        ) : (
          <span className="text-txt-tertiary">{__("No deadline")}</span>
        )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {isAuthorized("RightsRequest", "deleteRightsRequest") && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
