import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconPlusSmall,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Suspense, useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentApprovalList_addApproverMutation } from "#/__generated__/core/DocumentApprovalList_addApproverMutation.graphql";
import type { DocumentApprovalList_versionFragment$key } from "#/__generated__/core/DocumentApprovalList_versionFragment.graphql";
import { usePeople } from "#/hooks/graph/PeopleGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DocumentApprovalListItem } from "./DocumentApprovalListItem";

const versionFragment = graphql`
  fragment DocumentApprovalList_versionFragment on DocumentVersion {
    id
    canAddApprover: permission(action: "core:document-version:add-approver")
    approvalQuorums(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
      edges {
        node {
          decisions(first: 100, orderBy: { field: CREATED_AT, direction: ASC })
            @connection(key: "DocumentApprovalList_decisions") {
            __id
            edges {
              node {
                id
                approver {
                  id
                }
                ...DocumentApprovalListItemFragment
              }
            }
          }
        }
      }
    }
  }
`;

const addApproverMutation = graphql`
  mutation DocumentApprovalList_addApproverMutation(
    $input: AddDocumentVersionApproverInput!
    $connections: [ID!]!
  ) {
    addDocumentVersionApprover(input: $input) {
      approvalDecisionEdge @appendEdge(connections: $connections) {
        node {
          id
          approver {
            id
          }
          ...DocumentApprovalListItemFragment
        }
      }
    }
  }
`;

export function DocumentApprovalList(props: {
  versionFragmentRef: DocumentApprovalList_versionFragment$key;
  onRefetch: () => void;
}) {
  const { versionFragmentRef, onRefetch } = props;
  const { __ } = useTranslate();

  const version = useFragment(versionFragment, versionFragmentRef);
  const dialogRef = useDialogRef();
  const canManage = version.canAddApprover;

  const lastQuorum = version.approvalQuorums?.edges?.[0]?.node ?? null;
  const decisions = lastQuorum?.decisions;
  const edges = decisions?.edges ?? [];
  const existingApproverIds = edges.map(({ node }) => node.approver.id);

  return (
    <div>
      {canManage && (
        <div className="flex justify-end pb-3">
          <Button
            variant="secondary"
            icon={IconPlusSmall}
            onClick={() => dialogRef.current?.open()}
          >
            {__("Add approver")}
          </Button>
        </div>
      )}
      {edges.length === 0
        ? (
            <div className="text-sm text-txt-secondary text-center py-8">
              {__("No approval decisions yet.")}
            </div>
          )
        : (
            <div className="divide-y divide-border-solid">
              {edges.map(({ node }) => (
                <DocumentApprovalListItem
                  key={node.id}
                  fragmentRef={node}
                  canManage={canManage}
                  connectionId={decisions?.__id ?? ""}
                  onRefetch={onRefetch}
                />
              ))}
            </div>
          )}
      <Dialog ref={dialogRef} title={__("Add approver")}>
        <Suspense fallback={<DialogContent padded>{__("Loading...")}</DialogContent>}>
          <AddApproverDialogContent
            documentVersionId={version.id}
            existingApproverIds={existingApproverIds}
            connectionId={decisions?.__id ?? ""}
            onSuccess={() => {
              dialogRef.current?.close();
              onRefetch();
            }}
          />
        </Suspense>
      </Dialog>
    </div>
  );
}

function AddApproverDialogContent(props: {
  documentVersionId: string;
  existingApproverIds: string[];
  connectionId: string;
  onSuccess: () => void;
}) {
  const { documentVersionId, existingApproverIds, connectionId, onSuccess } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const allPeople = usePeople(organizationId, { excludeContractEnded: true });
  const people = allPeople.filter(p => !existingApproverIds.includes(p.id));
  const [selectedId, setSelectedId] = useState("");

  const [addApprover, isAdding] = useMutation<DocumentApprovalList_addApproverMutation>(
    addApproverMutation,
  );

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!selectedId) return;
        void addApprover({
          variables: {
            input: {
              documentVersionId,
              approverId: selectedId,
            },
            connections: [connectionId],
          },
          onCompleted: (_data, errors) => {
            if (errors?.length) {
              toast({
                title: __("Error"),
                description: errors[0].message,
                variant: "error",
              });
              return;
            }
            toast({
              title: __("Approver added"),
              description: __("The approver has been added successfully."),
              variant: "success",
            });
            onSuccess();
          },
          onError: (error) => {
            toast({
              title: __("Error"),
              description: error.message,
              variant: "error",
            });
          },
        });
      }}
    >
      <DialogContent padded>
        <label htmlFor="add-approver-select" className="block text-sm font-medium mb-1">{__("Approver")}</label>
        <select
          id="add-approver-select"
          className="w-full rounded-md border border-border-solid bg-bg-primary px-3 py-2 text-sm"
          value={selectedId}
          onChange={e => setSelectedId(e.target.value)}
        >
          <option value="">{__("Select a person...")}</option>
          {people.map(p => (
            <option key={p.id} value={p.id}>
              {p.fullName}
            </option>
          ))}
        </select>
      </DialogContent>
      <DialogFooter>
        <Button type="submit" disabled={!selectedId || isAdding}>
          {__("Add approver")}
        </Button>
      </DialogFooter>
    </form>
  );
}
