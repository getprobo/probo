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

import type { PeopleGraphPaginatedFragment$key } from "#/__generated__/core/PeopleGraphPaginatedFragment.graphql";
import type { PeopleGraphPaginatedQuery } from "#/__generated__/core/PeopleGraphPaginatedQuery.graphql";
import type { SignatureDocumentsDialogMutation } from "#/__generated__/core/SignatureDocumentsDialogMutation.graphql";
import {
  paginatedPeopleFragment,
  paginatedPeopleQuery,
} from "#/hooks/graph/PeopleGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  documentIds: string[];
  children: ReactNode;
  onSave: () => void;
};

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
          return sprintf(__("%s signature requests sent"), actualRequestsCount);
        },
        errorMessage: __("Failed to send signature requests"),
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
            {__("Send signature requests")}
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
  const data = useLazyLoadQuery<PeopleGraphPaginatedQuery>(
    paginatedPeopleQuery,
    {
      organizationId,
    },
  );
  const {
    data: page,
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment(
    paginatedPeopleFragment,
    data.organization as PeopleGraphPaginatedFragment$key,
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
