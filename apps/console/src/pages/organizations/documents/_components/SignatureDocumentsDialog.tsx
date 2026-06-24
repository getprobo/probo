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

import { sprintf } from "@probo/helpers";
import { useList } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Avatar,
  Breadcrumb,
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  IconChevronDown,
  Spinner,
  Table,
  Tbody,
  Td,
  Tr,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode, Suspense } from "react";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { SignatureDocumentsDialogMutation } from "#/__generated__/core/SignatureDocumentsDialogMutation.graphql";
import type { SignatureDocumentsDialogPeopleFragment$key } from "#/__generated__/core/SignatureDocumentsDialogPeopleFragment.graphql";
import type { SignatureDocumentsDialogPeopleQuery } from "#/__generated__/core/SignatureDocumentsDialogPeopleQuery.graphql";
import type { SignatureDocumentsDialogPeopleRefetchQuery } from "#/__generated__/core/SignatureDocumentsDialogPeopleRefetchQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  documentIds: string[];
  children: ReactNode;
  onSave: () => void;
};

const signatureDocumentsDialogPeopleQuery = graphql`
  query SignatureDocumentsDialogPeopleQuery(
    $organizationId: ID!
    $filter: ProfileFilter
  ) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...SignatureDocumentsDialogPeopleFragment
          @arguments(filter: $filter)
      }
    }
  }
`;

const signatureDocumentsDialogPeopleFragment = graphql`
  fragment SignatureDocumentsDialogPeopleFragment on Organization
  @refetchable(queryName: "SignatureDocumentsDialogPeopleRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "ProfileOrder"
      defaultValue: { direction: ASC, field: FULL_NAME }
    }
    filter: { type: "ProfileFilter", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    profiles(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "SignatureDocumentsDialog_profiles") {
      edges {
        node {
          id
          fullName
          emailAddress
        }
      }
    }
  }
`;

const documentsSignatureMutation = graphql`
  mutation SignatureDocumentsDialogMutation(
    $input: BulkRequestSignaturesInput!
  ) {
    bulkRequestSignatures(input: $input) {
      documentVersionSignatureEdges {
        node {
          id
          state
        }
      }
    }
  }
`;

export function SignatureDocumentsDialog({
  documentIds,
  children,
  onSave,
}: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const { list: selectedPeople, toggle } = useList<string>([]);

  const schema = z.object({});

  const [publishMutation]
    = useMutationWithToasts<SignatureDocumentsDialogMutation>(
      documentsSignatureMutation,
      {
        successMessage: (response) => {
          const actualRequestsCount
            = response.bulkRequestSignatures.documentVersionSignatureEdges.length;
          return sprintf(__("%s signature requests created"), actualRequestsCount);
        },
        errorMessage: __("Failed to create signature requests"),
      },
    );

  const {
    handleSubmit,
    formState: { isSubmitting },
  } = useFormWithSchema(schema, {});

  const onSubmit = async () => {
    await publishMutation({
      variables: {
        input: {
          documentIds,
          signatoryIds: selectedPeople,
        },
      },
      onSuccess: () => {
        dialogRef.current?.close();
        onSave();
      },
    });
  };

  return (
    <Dialog
      className="max-w-xl"
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Documents"), __("Signature requests")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent>
          <Suspense fallback={<Spinner />}>
            <PeopleList onChange={toggle} selectedPeople={selectedPeople} />
          </Suspense>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            disabled={selectedPeople.length === 0 || isSubmitting}
          >
            {__("Request signatures")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function PeopleList({
  onChange,
  selectedPeople,
}: {
  onChange: (id: string) => void;
  selectedPeople: string[];
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const data = useLazyLoadQuery<SignatureDocumentsDialogPeopleQuery>(
    signatureDocumentsDialogPeopleQuery,
    {
      organizationId,
      filter: { contractEnded: false, state: "ACTIVE" },
    },
  );
  const {
    data: page,
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<
    SignatureDocumentsDialogPeopleRefetchQuery,
    SignatureDocumentsDialogPeopleFragment$key
  >(
    signatureDocumentsDialogPeopleFragment,
    data.organization,
  );
  const profiles = page.profiles.edges.map(edge => edge.node);
  return (
    <>
      <Table className="border-none rounded-none">
        <Tbody>
          {profiles.map(person => (
            <Tr key={person.id}>
              <Td width={75}>
                <Checkbox
                  checked={selectedPeople.includes(person.id)}
                  onChange={() => onChange(person.id)}
                />
              </Td>
              <Td>
                <div className="flex gap-3 items-center">
                  <Avatar name={person.fullName} />
                  <div>
                    <div className="text-sm">{person.fullName}</div>
                    <div className="text-xs text-txt-tertiary">
                      {person.emailAddress}
                    </div>
                  </div>
                </div>
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
      {isLoadingNext && <Spinner className="mt-3 mx-auto" />}
      {hasNext && (
        <Button
          variant="tertiary"
          onClick={() => loadNext(20)}
          className="mx-auto"
          icon={IconChevronDown}
          type="button"
        >
          {sprintf(__("Show %s more"), profiles.length)}
        </Button>
      )}
    </>
  );
}
