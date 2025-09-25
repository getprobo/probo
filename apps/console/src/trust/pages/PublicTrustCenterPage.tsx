import { useParams, Navigate } from "react-router";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { PublicTrustCenterLayout } from "/layouts/PublicTrustCenterLayout";
import { PublicTrustCenterAudits } from "../components/PublicTrustCenterAudits";
import { PublicTrustCenterVendors } from "../components/PublicTrustCenterVendors";
import { PublicTrustCenterDocuments } from "../components/PublicTrustCenterDocuments";
import { PublicTrustCenterAccessRequestDialog } from "../components/PublicTrustCenterAccessRequestDialog";
import { NDAAcceptanceDialog } from "../components/NDAAcceptanceDialog";
import { Spinner, Button, IconLock } from "@probo/ui";
import { useTrustCenterQuery, type TrustCenterDocument, type TrustCenterAudit, type TrustCenterVendor } from "/hooks/useTrustCenterQueries";

export type { TrustCenterDocument, TrustCenterAudit, TrustCenterVendor };

export default function PublicTrustCenterPage() {
  const { __ } = useTranslate();
  const { slug } = useParams<{ slug: string }>();

  const { data, isLoading, error } = useTrustCenterQuery(slug || "");

  const organization = data?.trustCenterBySlug?.organization;
  const organizationName = organization?.name || "";

  usePageTitle(
    organizationName ? `${organizationName} - Trust Center` : "Trust Center"
  );

  if (!slug) {
    return <Navigate to="/" replace />;
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Spinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {__("Error Loading Trust Center")}
          </h1>
          <p className="text-gray-600">
            {__("There was an error loading the trust center. Please try again later.")}
          </p>
        </div>
      </div>
    );
  }

  if (!data?.trustCenterBySlug) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {__("Trust Center Not Found")}
          </h1>
          <p className="text-gray-600">
            {__("The trust center you're looking for doesn't exist.")}
          </p>
        </div>
      </div>
    );
  }

  const { trustCenterBySlug } = data;
  const { documents, audits, vendors, isUserAuthenticated, hasAcceptedNonDisclosureAgreement } = trustCenterBySlug;

  const trustCenterDocuments = documents.edges.map((edge) => edge.node) as TrustCenterDocument[];
  const trustCenterAudits = audits.edges.map((edge) => edge.node) as TrustCenterAudit[];
  const trustCenterVendors = vendors.edges.map((edge) => edge.node) as TrustCenterVendor[];

  const showNdaDialog = isUserAuthenticated && !hasAcceptedNonDisclosureAgreement;

  let showRequestAllButton = false;

  for (const doc of trustCenterDocuments) {
    if (!doc.isUserAuthorized && !doc.hasUserRequestedAccess) {
      showRequestAllButton = true;
      break;
    }
  }

  if (!showRequestAllButton) {
    const reportsWithData = trustCenterAudits.filter(audit => audit.report !== null);
    for (const audit of reportsWithData) {
      if (!audit.report!.isUserAuthorized && !audit.report!.hasUserRequestedAccess) {
        showRequestAllButton = true;
        break;
      }
    }
  }

  return (
    <>
      {showNdaDialog && (
        <NDAAcceptanceDialog
          trustCenterId={trustCenterBySlug.id}
          organizationName={organizationName}
          ndaFileName={trustCenterBySlug.ndaFileName}
          ndaFileUrl={trustCenterBySlug.ndaFileUrl}
        />
      )}

      <PublicTrustCenterLayout
        organizationName={organizationName}
        organizationLogo={organization?.logoUrl}
      >
        <div className="space-y-12">
          {showRequestAllButton && (
            <div className="flex justify-end">
              <PublicTrustCenterAccessRequestDialog
                trigger={
                  <Button
                    variant="primary"
                    icon={IconLock}
                  >
                    {__("Request All Access")}
                  </Button>
                }
                trustCenterId={trustCenterBySlug.id}
                organizationName={organizationName}
              />
            </div>
          )}

          <PublicTrustCenterAudits
            audits={trustCenterAudits}
            organizationName={organizationName}
            isAuthenticated={isUserAuthenticated}
            trustCenterId={trustCenterBySlug.id}
          />
          <PublicTrustCenterDocuments
            documents={trustCenterDocuments}
            organizationName={organizationName}
            isAuthenticated={isUserAuthenticated}
            trustCenterId={trustCenterBySlug.id}
          />
          <PublicTrustCenterVendors
            vendors={trustCenterVendors}
            organizationName={organizationName}
          />
        </div>
      </PublicTrustCenterLayout>
    </>
  );
}
