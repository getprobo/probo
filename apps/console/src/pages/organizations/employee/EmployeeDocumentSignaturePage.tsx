import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Card, Spinner, useToast } from "@probo/ui";
import { useEffect, useRef, useState } from "react";
import {
  type PreloadedQuery,
  useFragment,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "react-relay";
import { useNavigate } from "react-router";
import { useWindowSize } from "usehooks-ts";

import type { EmployeeDocumentSignaturePageDocumentFragment$key } from "#/__generated__/core/EmployeeDocumentSignaturePageDocumentFragment.graphql";
import type { EmployeeDocumentSignaturePageExportSignablePDFMutation } from "#/__generated__/core/EmployeeDocumentSignaturePageExportSignablePDFMutation.graphql";
import type { EmployeeDocumentSignaturePageQuery } from "#/__generated__/core/EmployeeDocumentSignaturePageQuery.graphql";
import type { EmployeeDocumentSignaturePageSignMutation } from "#/__generated__/core/EmployeeDocumentSignaturePageSignMutation.graphql";
import { PDFPreview } from "#/components/documents/PDFPreview";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { VersionActions } from "./_components/VersionActions";
import { VersionRow } from "./_components/VersionRow";

export const employeeDocumentSignaturePageQuery = graphql`
  query EmployeeDocumentSignaturePageQuery($documentId: ID!) {
    viewer @required(action: THROW) {
      signableDocument(id: $documentId) {
        id
        ...EmployeeDocumentSignaturePageDocumentFragment
      }
    }
  }
`;

const documentFragment = graphql`
  fragment EmployeeDocumentSignaturePageDocumentFragment on SignableDocument {
    id
    title
    signed
    versions(first: 100, orderBy: { field: CREATED_AT, direction: DESC })
      @required(action: THROW) {
      edges @required(action: THROW) {
        node @required(action: THROW) {
          id
          ...VersionActionsFragment
          ...VersionRowFragment
        }
      }
    }
  }
`;

const signDocumentMutation = graphql`
  mutation EmployeeDocumentSignaturePageSignMutation(
    $input: SignDocumentInput!
  ) {
    signDocument(input: $input) {
      documentVersionSignature {
        id
        state
      }
    }
  }
`;

const exportSignableVersionDocumentPDFMutation = graphql`
  mutation EmployeeDocumentSignaturePageExportSignablePDFMutation(
    $input: ExportSignableDocumentVersionPDFInput!
  ) {
    exportSignableVersionDocumentPDF(input: $input) {
      data
    }
  }
`;

export function EmployeeDocumentSignaturePage(props: {
  queryRef: PreloadedQuery<EmployeeDocumentSignaturePageQuery>;
}) {
  const { queryRef } = props;
  const { viewer } = usePreloadedQuery<EmployeeDocumentSignaturePageQuery>(
    employeeDocumentSignaturePageQuery,
    queryRef,
  );
  const document = viewer.signableDocument;

  if (!document) {
    return (
      <div className="flex items-center justify-center h-full">
        <Spinner />
      </div>
    );
  }

  return <DocumentSignatureContent fKey={document} />;
}

function DocumentSignatureContent({
  fKey,
}: {
  fKey: EmployeeDocumentSignaturePageDocumentFragment$key;
}) {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { width } = useWindowSize();
  const isMobile = width < 1100;
  const isDesktop = !isMobile;
  const organizationId = useOrganizationId();

  const documentData
    = useFragment<EmployeeDocumentSignaturePageDocumentFragment$key>(
      documentFragment,
      fKey,
    );

  const versions = documentData.versions.edges.map(({ node }) => node);

  const [selectedVersionId, setSelectedVersionId] = useState<
    string | undefined
  >(() => versions[0]?.id);

  const selectedVersion = versions.find(v => v?.id === selectedVersionId);

  usePageTitle(__("Sign Document"));
  const { toast } = useToast();

  const [signDocument, isSigning]
    = useMutation<EmployeeDocumentSignaturePageSignMutation>(
      signDocumentMutation,
    );

  const [exportSignableVersionDocumentPDF]
    = useMutation<EmployeeDocumentSignaturePageExportSignablePDFMutation>(
      exportSignableVersionDocumentPDFMutation,
    );

  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const pdfUrlRef = useRef<string | null>(null);

  const handleSign = (versionId: string) => {
    signDocument({
      variables: {
        input: {
          documentVersionId: versionId,
        },
      },
      updater: (store) => {
        const signableDoc = store.get(documentData.id);
        if (signableDoc) {
          signableDoc.setValue(true, "signed");
        }
        store.invalidateStore();
      },
      onCompleted: () => {
        toast({
          title: __("Success"),
          description: __("Document signed successfully"),
          variant: "success",
        });
        void navigate(`/organizations/${organizationId}/employee`);
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to sign document"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  useEffect(() => {
    if (!selectedVersion?.id) return;

    exportSignableVersionDocumentPDF({
      variables: {
        input: {
          documentVersionId: selectedVersion.id,
        },
      },
      onCompleted: (data, errors): void => {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to load PDF"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        if (data.exportSignableVersionDocumentPDF?.data) {
          const dataUrl = data.exportSignableVersionDocumentPDF.data;
          pdfUrlRef.current = dataUrl;
          setPdfUrl(dataUrl);
        }
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to load PDF"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });

    return () => {
      pdfUrlRef.current = null;
    };
  }, [selectedVersion?.id, exportSignableVersionDocumentPDF, toast, __]);

  return (
    <div
      className="fixed bg-level-2 flex flex-col"
      style={{ top: "3rem", left: 0, right: 0, bottom: 0 }}
    >
      <div className="grid lg:grid-cols-2 min-h-0 h-full">
        <div className="w-full lg:w-[440px] mx-auto py-20 overflow-y-auto scrollbar-hide">
          <h1 className="text-2xl font-semibold mb-6">
            {documentData.title || ""}
          </h1>

          <Card className="mb-6 overflow-hidden">
            <div className="divide-y divide-border-solid">
              {versions.map((version) => {
                return (
                  <VersionRow
                    key={version.id}
                    fKey={version}
                    isSelected={version.id === selectedVersionId}
                    onSelect={() => setSelectedVersionId(version.id)}
                  />
                );
              })}
            </div>
          </Card>

          <p className="text-txt-secondary text-sm mb-6">
            {__("Please review the document carefully before signing.")}
          </p>

          <div className="min-h-[60px]">
            {selectedVersion
              ? (
                  <VersionActions
                    fKey={selectedVersion}
                    isSigning={isSigning}
                    onSign={handleSign}
                    onBack={() =>
                      void navigate(`/organizations/${organizationId}/employee`)}
                  />
                )
              : null}
          </div>
        </div>

        {isDesktop && (
          <div className="bg-subtle h-full border-l border-border-solid min-h-0">
            {pdfUrl && (
              <PDFPreview src={pdfUrl} name={documentData.title || ""} />
            )}
          </div>
        )}
      </div>
    </div>
  );
}
