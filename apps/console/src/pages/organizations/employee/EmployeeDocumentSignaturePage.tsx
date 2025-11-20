import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Spinner,
  IconCircleCheck,
} from "@probo/ui";
import clsx from "clsx";
import {
  usePreloadedQuery,
  useFragment,
  useMutation,
  type PreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import type { EmployeeDocumentSignaturePageQuery } from "./__generated__/EmployeeDocumentSignaturePageQuery.graphql";
import { usePageTitle } from "@probo/hooks";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { EmployeeDocumentSignaturePageSignMutation } from "./__generated__/EmployeeDocumentSignaturePageSignMutation.graphql";
import type { EmployeeDocumentSignaturePageExportSignablePDFMutation } from "./__generated__/EmployeeDocumentSignaturePageExportSignablePDFMutation.graphql";
import { useNavigate, useParams } from "react-router";
import { PDFPreview } from "/components/documents/PDFPreview";
import { useWindowSize } from "usehooks-ts";
import { useState, useEffect, useRef, useMemo } from "react";
import type { EmployeeDocumentSignaturePageDocumentFragment$key } from "./__generated__/EmployeeDocumentSignaturePageDocumentFragment.graphql";
import type { EmployeeDocumentSignaturePageVersionFragment$key } from "./__generated__/EmployeeDocumentSignaturePageVersionFragment.graphql";
import { useToast } from "@probo/ui";
import { formatError, type GraphQLError } from "@probo/helpers";

export const employeeDocumentSignatureQuery = graphql`
  query EmployeeDocumentSignaturePageQuery($documentId: ID!) {
    viewer {
      id
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
    versions(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
      edges {
        node {
          id
          ...EmployeeDocumentSignaturePageVersionFragment
        }
      }
    }
  }
`;

const versionFragment = graphql`
  fragment EmployeeDocumentSignaturePageVersionFragment on DocumentVersion {
    id
    version
    signed
    publishedAt
  }
`;

const signDocumentMutation = graphql`
  mutation EmployeeDocumentSignaturePageSignMutation($input: SignDocumentInput!) {
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

type Props = {
  queryRef: PreloadedQuery<EmployeeDocumentSignaturePageQuery>;
  organizationId?: string;
};

export default function EmployeeDocumentSignaturePage(props: Props) {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const data = usePreloadedQuery(employeeDocumentSignatureQuery, props.queryRef);
  const { width } = useWindowSize();
  const isMobile = width < 1100;
  const isDesktop = !isMobile;

  const document = data.viewer.signableDocument;
  const params = useParams<{ organizationId?: string }>();

  const organizationId = props.organizationId || params.organizationId;

  if (!document) {
    return (
      <div className="flex items-center justify-center h-full">
        <Spinner />
      </div>
    );
  }

  const documentData = useFragment<EmployeeDocumentSignaturePageDocumentFragment$key>(
    documentFragment,
    document
  );

  const versions = documentData.versions?.edges
    ?.map((edge) => edge?.node)
    .filter(Boolean) || [];

  const [selectedVersionId, setSelectedVersionId] = useState<string | undefined>(
    versions[0]?.id
  );

  const selectedVersion = useMemo(() => {
    return versions.find((v) => v?.id === selectedVersionId);
  }, [versions, selectedVersionId]);

  usePageTitle(__("Sign Document"));
  const { toast } = useToast();

  const [signDocument, isSigning] = useMutationWithToasts<EmployeeDocumentSignaturePageSignMutation>(
    signDocumentMutation,
    {
      successMessage: __("Document signed successfully"),
      errorMessage: __("Failed to sign document"),
    }
  );

  const [exportSignableVersionDocumentPDF] = useMutation<EmployeeDocumentSignaturePageExportSignablePDFMutation>(
    exportSignableVersionDocumentPDFMutation
  );

  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const pdfUrlRef = useRef<string | null>(null);

  const handleSign = async () => {
    if (!selectedVersionData?.id) {
      console.error("Cannot sign: version.id is missing");
      return;
    }

    if (!organizationId) {
      console.error("Cannot sign: organizationId is missing");
      return;
    }

    await signDocument({
      variables: {
        input: {
          documentVersionId: selectedVersionData.id,
        },
      },
      onCompleted: () => {
        navigate(`/organizations/${organizationId}/employee`);
      },
      onError: (error) => {
        console.error("Error signing document:", error);
      },
    });
  };

  const selectedVersionData = selectedVersion
    ? useFragment<EmployeeDocumentSignaturePageVersionFragment$key>(
        versionFragment,
        selectedVersion
      )
    : null;

  const isSigned = selectedVersionData?.signed ?? false;

  useEffect(() => {
    if (!selectedVersionData?.id) return;

    exportSignableVersionDocumentPDF({
      variables: {
        input: {
          documentVersionId: selectedVersionData.id,
        },
      },
      onCompleted: (data, errors): void => {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to load PDF"), errors as GraphQLError[]),
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
          description: formatError(__("Failed to load PDF"), error as GraphQLError),
          variant: "error",
        });
      },
    });

    return () => {
      pdfUrlRef.current = null;
    };
  }, [selectedVersionData?.id, exportSignableVersionDocumentPDF, toast, __]);

  return (
    <div className="fixed bg-level-2 flex flex-col" style={{ top: '3rem', left: 0, right: 0, bottom: 0 }}>
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
                    version={version}
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
            {isSigned ? (
              <>
                <Button
                  onClick={() => navigate(`/organizations/${organizationId}/employee`)}
                  className="h-10 w-full"
                  variant="secondary"
                >
                  {__("Back to Documents")}
                </Button>
                <p className="text-xs text-txt-tertiary mt-2 h-5">
                  {/* Spacer to maintain consistent layout */}
                </p>
              </>
            ) : (
              <>
                <Button
                  onClick={handleSign}
                  className="h-10 w-full"
                  disabled={isSigning}
                  icon={isSigning ? Spinner : undefined}
                >
                  {__("I acknowledge and agree")}
                </Button>
                <p className="text-xs text-txt-tertiary mt-2 h-5">
                  {__(
                    "By clicking 'I acknowledge and agree', your digital signature will be recorded."
                  )}
                </p>
              </>
            )}
          </div>
        </div>

        {isDesktop && (
          <div className="bg-subtle h-full border-l border-border-solid min-h-0">
            {pdfUrl && <PDFPreview src={pdfUrl} name={documentData.title || ""} />}
          </div>
        )}
      </div>
    </div>
  );
}

function VersionRow({
  version,
  isSelected,
  onSelect,
}: {
  version: EmployeeDocumentSignaturePageVersionFragment$key;
  isSelected: boolean;
  onSelect: () => void;
}) {
  const { __ } = useTranslate();
  const versionData = useFragment<EmployeeDocumentSignaturePageVersionFragment$key>(
    versionFragment,
    version
  );
  const isVersionSigned = versionData.signed;

  return (
    <div
      onClick={onSelect}
      className={clsx(
        "flex items-center gap-3 py-3 px-4 transition-colors cursor-pointer",
        isSelected
          ? "bg-blue-50 border-l-4 border-blue-500"
          : "bg-transparent hover:bg-level-1"
      )}
    >
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-level-2 flex-shrink-0">
        {isVersionSigned ? (
          <IconCircleCheck
            size={20}
            className="text-txt-success"
          />
        ) : (
          <span className="text-sm font-semibold text-txt-tertiary">
            {versionData.version}
          </span>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <p
          className={clsx(
            "text-sm font-medium truncate",
            isVersionSigned
              ? "text-txt-tertiary"
              : "text-txt-primary"
          )}
        >
          {versionData.publishedAt
            ? `v${versionData.version} - ${(() => {
                const date = new Date(versionData.publishedAt);
                const day = String(date.getDate()).padStart(2, '0');
                const month = String(date.getMonth() + 1).padStart(2, '0');
                const year = date.getFullYear();
                return `${day}/${month}/${year}`;
              })()}`
            : `v${versionData.version}`}
        </p>
      </div>
      <div className="flex-shrink-0">
        <span
          className={clsx(
            "inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium",
            isVersionSigned
              ? "bg-green-100 text-green-800"
              : isSelected
              ? "bg-blue-100 text-blue-800"
              : "bg-gray-100 text-gray-700"
          )}
        >
          {isVersionSigned
            ? __("Signed")
            : isSelected
            ? __("In review")
            : __("Waiting signature")}
        </span>
      </div>
    </div>
  );
}
