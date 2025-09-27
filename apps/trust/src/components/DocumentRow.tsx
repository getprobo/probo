import { graphql } from "relay-runtime";
import type { DocumentRowFragment$key } from "./__generated__/DocumentRowFragment.graphql";
import { useFragment } from "react-relay";
import {
  Button,
  IconArrowInbox,
  IconLock,
  IconPageTextLine,
  Spinner,
  Td,
  Tr,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useIsAuthenticated } from "/hooks/useIsAuthenticated";
import type { DocumentRowDownloadMutation } from "./__generated__/DocumentRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile } from "@probo/helpers";
import { RequestAccessDialog } from "/components/RequestAccessDialog.tsx";

const downloadMutation = graphql`
  mutation DocumentRowDownloadMutation($input: ExportDocumentPDFInput!) {
    exportDocumentPDF(input: $input) {
      data
    }
  }
`;

const documentRowFragment = graphql`
  fragment DocumentRowFragment on Document {
    id
    title
  }
`;

export function DocumentRow(props: { document: DocumentRowFragment$key }) {
  const document = useFragment(documentRowFragment, props.document);
  const { __ } = useTranslate();
  const isAuthenticated = useIsAuthenticated();
  const [commitDownload, downloading] =
    useMutationWithToasts<DocumentRowDownloadMutation>(downloadMutation);
  const handleDownload = () => {
    commitDownload({
      variables: {
        input: {
          documentId: document.id,
        },
      },
      onSuccess(response) {
        downloadFile(response.exportDocumentPDF.data, document.title);
      },
    });
  };
  return (
    <Tr className="text-sm *:border-border-solid *:border-b-1">
      <Td>
        <div className="flex items-center gap-2">
          <IconPageTextLine size={16} />
          {document.title}
        </div>
      </Td>
      <Td className="text-end">
        {isAuthenticated ? (
          <Button
            className="ml-auto"
            variant="secondary"
            disabled={downloading}
            icon={downloading ? Spinner : IconArrowInbox}
            onClick={handleDownload}
          >
            {__("Download")}
          </Button>
        ) : (
          <RequestAccessDialog>
            <Button className="ml-auto" variant="secondary" icon={IconLock}>
              {__("Request access")}
            </Button>
          </RequestAccessDialog>
        )}
      </Td>
    </Tr>
  );
}
