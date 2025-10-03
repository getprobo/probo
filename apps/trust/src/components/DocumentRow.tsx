import { graphql } from "relay-runtime";
import type { DocumentRowFragment$key } from "./__generated__/DocumentRowFragment.graphql";
import { useFragment } from "react-relay";
import {
  Button,
  IconArrowInbox,
  IconLock,
  IconPageTextLine,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { DocumentRowDownloadMutation } from "./__generated__/DocumentRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile } from "@probo/helpers";
import { RequestAccessDialog } from "/components/RequestAccessDialog.tsx";
import { useState } from "react";

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
    isUserAuthorized
    hasUserRequestedAccess
  }
`;

export function DocumentRow(props: { document: DocumentRowFragment$key }) {
  const document = useFragment(documentRowFragment, props.document);
  const { __ } = useTranslate();
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
  const [hasRequested, setHasRequested] = useState(
    document.hasUserRequestedAccess,
  );
  return (
    <div className="text-sm border-1 border-border-solid -mt-[1px] flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconPageTextLine size={16} className=" flex-none text-txt-tertiary" />
        {document.title}
      </div>
      {document.isUserAuthorized ? (
        <Button
          className="w-full md:w-max"
          variant="secondary"
          disabled={downloading}
          icon={downloading ? Spinner : IconArrowInbox}
          onClick={handleDownload}
        >
          {__("Download")}
        </Button>
      ) : (
        <RequestAccessDialog
          documentId={document.id}
          onSuccess={() => setHasRequested(true)}
        >
          <Button
            disabled={hasRequested}
            className="w-full md:w-max"
            variant="secondary"
            icon={IconLock}
          >
            {hasRequested ? __("Access requested") : __("Request access")}
          </Button>
        </RequestAccessDialog>
      )}
    </div>
  );
}
