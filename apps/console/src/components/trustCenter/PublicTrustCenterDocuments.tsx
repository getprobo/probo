import {
  Card,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  DocumentTypeBadge,
  Button,
  IconArrowDown,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import type { PublicTrustCenterDocumentsExportPDFMutation } from "./__generated__/PublicTrustCenterDocumentsExportPDFMutation.graphql";

const exportDocumentVersionPDFMutation = graphql`
  mutation PublicTrustCenterDocumentsExportPDFMutation(
    $input: ExportDocumentVersionPDFInput!
  ) {
    exportDocumentVersionPDF(input: $input) {
      data
    }
  }
`;

type Document = {
  id: string;
  title: string;
  documentType: string;
  versions: {
    edges: Array<{
      node: {
        id: string;
        status: string;
      };
    }>;
  };
};

type Props = {
  documents: Document[];
};

export function PublicTrustCenterDocuments({ documents }: Props) {
  const { __ } = useTranslate();
  const [exportDocumentVersionPDF] = useMutation<PublicTrustCenterDocumentsExportPDFMutation>(exportDocumentVersionPDFMutation);

  const handleDownload = (document: Document) => {
    const latestVersion = document.versions.edges[0]?.node;
    if (!latestVersion) return;

    exportDocumentVersionPDF({
      variables: {
        input: { documentVersionId: latestVersion.id },
      },
      onCompleted: (data) => {
        if (data.exportDocumentVersionPDF?.data) {
          const link = window.document.createElement("a");
          link.href = data.exportDocumentVersionPDF.data;
          link.download = `${document.title}.pdf`;
          window.document.body.appendChild(link);
          link.click();
          window.document.body.removeChild(link);
        }
      },
    });
  };

  if (documents.length === 0) {
    return (
      <Card padded>
        <div className="text-center py-8">
          <h2 className="text-xl font-semibold text-txt-primary mb-2">
            {__("Documents")}
          </h2>
          <p className="text-txt-secondary">
            {__("No documents are currently available.")}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Card padded className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold text-txt-primary">
          {__("Documents")}
        </h2>
        <p className="text-sm text-txt-secondary mt-1">
          {__("Security and compliance documentation")}
        </p>
      </div>

      <Table>
        <Thead>
          <Tr>
            <Th className="w-1/2">{__("Document")}</Th>
            <Th className="w-1/4">{__("Type")}</Th>
            <Th className="w-1/4">{__("Download")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {documents.map((document) => {
            const latestVersion = document.versions.edges[0]?.node;

            return (
              <Tr key={document.id}>
                <Td>
                  <div className="font-medium">
                    {document.title}
                  </div>
                </Td>
                <Td>
                  <DocumentTypeBadge type={document.documentType} />
                </Td>
                <Td>
                  {latestVersion ? (
                    <Button
                      variant="secondary"
                      icon={IconArrowDown}
                      onClick={() => handleDownload(document)}
                    >
                      {__("Download")}
                    </Button>
                  ) : (
                    <span className="text-txt-tertiary text-sm">
                      {__("Not available")}
                    </span>
                  )}
                </Td>
              </Tr>
            );
          })}
        </Tbody>
      </Table>
    </Card>
  );
}
