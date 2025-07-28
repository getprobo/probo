import { useParams, Navigate } from "react-router";
import { usePreloadedQuery, useLazyLoadQuery, type PreloadedQuery } from "react-relay";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import { publicTrustCenterQuery } from "/hooks/graph/PublicTrustCenterGraph";
import type { PublicTrustCenterGraphQuery } from "/hooks/graph/__generated__/PublicTrustCenterGraphQuery.graphql";
import { PublicTrustCenterLayout } from "/layouts/PublicTrustCenterLayout";
import { PublicTrustCenterAudits } from "../components/trustCenter/PublicTrustCenterAudits";
import { PublicTrustCenterVendors } from "../components/trustCenter/PublicTrustCenterVendors";
import { PublicTrustCenterDocuments } from "../components/trustCenter/PublicTrustCenterDocuments";

type Props = {
  queryRef?: PreloadedQuery<PublicTrustCenterGraphQuery>;
};

export default function PublicTrustCenterPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const { slug } = useParams<{ slug: string }>();

  if (!slug) {
    return <Navigate to="/" replace />;
  }

  const data = queryRef
    ? usePreloadedQuery(publicTrustCenterQuery, queryRef)
    : useLazyLoadQuery(publicTrustCenterQuery, { slug });

  const { trustCenterBySlug } = data as any;

  if (!trustCenterBySlug) {
    return <Navigate to="/auth/login" replace />;
  }

  if (!trustCenterBySlug.active) {
    return <Navigate to="/auth/login" replace />;
  }

  const { organization } = trustCenterBySlug;

  usePageTitle(`${organization.name} - Trust Center`);

  const documents = trustCenterBySlug.publicDocuments.edges
    .map((e: any) => e.node)

  const audits = trustCenterBySlug.publicAudits.edges
    .map((e: any) => e.node);

  const vendors = trustCenterBySlug.publicVendors.edges
    .map((e: any) => e.node);

  return (
    <PublicTrustCenterLayout
      organizationName={organization.name}
      organizationLogo={organization.logoUrl}
    >
      <div className="space-y-8">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-txt-primary mb-4">
            {sprintf(__("%s Trust Center"), organization.name)}
          </h1>
          <p className="text-lg text-txt-secondary max-w-2xl mx-auto">
            {__("Explore our security practices, compliance certifications, and transparency reports.")}
          </p>
        </div>
        <PublicTrustCenterAudits
          audits={audits}
          organizationName={organization.name}
        />
        <PublicTrustCenterDocuments
          documents={documents}
        />
        <PublicTrustCenterVendors
          vendors={vendors}
          organizationName={organization.name}
        />
      </div>
    </PublicTrustCenterLayout>
  );
}
