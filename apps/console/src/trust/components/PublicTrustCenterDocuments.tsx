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
  IconLock,
  useToast,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useExportDocumentPDF, type TrustCenterDocument } from "/hooks/useTrustCenterQueries";
import { PublicTrustCenterAccessRequestDialog } from "./PublicTrustCenterAccessRequestDialog";

type Props = {
  documents: TrustCenterDocument[];
  isAuthenticated: boolean;
  trustCenterId: string;
  organizationName: string;
};

export function PublicTrustCenterDocuments({
  documents,
  isAuthenticated,
  trustCenterId,
  organizationName
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const mutation = useExportDocumentPDF();

  const handleDownload = (document: TrustCenterDocument) => {
    mutation.mutate(document.id, {
      onSuccess: (data) => {
        if (data.exportDocumentPDF?.data) {
          const link = window.document.createElement("a");
          link.href = data.exportDocumentPDF.data;
          link.download = `${document.title}.pdf`;
          window.document.body.appendChild(link);
          link.click();
          window.document.body.removeChild(link);
        }
      },
      onError: () => {
        toast({
          title: __("Download Failed"),
          description: __("Unable to download the document"),
          variant: "error",
        });
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
                  {!isAuthenticated ? (
                    <PublicTrustCenterAccessRequestDialog
                      trigger={
                        <Button
                          variant="secondary"
                          icon={IconLock}
                        >
                          {__("Request Access")}
                        </Button>
                      }
                      trustCenterId={trustCenterId}
                      organizationName={organizationName}
                    />
                  ) : (
                    <Button
                      variant="secondary"
                      icon={IconArrowDown}
                      onClick={() => handleDownload(document)}
                      disabled={mutation.isPending}
                    >
                      {mutation.isPending ? __("Downloading...") : __("Download")}
                    </Button>
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
