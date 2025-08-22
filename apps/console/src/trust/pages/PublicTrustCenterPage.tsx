import { useParams, Navigate } from "react-router";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { PublicTrustCenterLayout } from "/layouts/PublicTrustCenterLayout";
import { PublicTrustCenterAudits } from "../components/PublicTrustCenterAudits";
import { PublicTrustCenterVendors } from "../components/PublicTrustCenterVendors";
import { PublicTrustCenterDocuments } from "../components/PublicTrustCenterDocuments";
import { TrustRelayProvider } from "/providers/TrustRelayProvider";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "react-relay";
import { Suspense } from "react";
import { Spinner } from "@probo/ui";
import type { PublicTrustCenterPageQuery } from "./__generated__/PublicTrustCenterPageQuery.graphql";

export interface TrustCenterDocument {
  id: string;
  title: string;
  documentType: string;
}

export interface TrustCenterAudit {
  id: string;
  framework: {
    name: string;
  };
  report: {
    id: string;
    filename: string;
    downloadUrl: string | null;
  } | null;
}

export interface TrustCenterVendor {
  id: string;
  name: string;
  category: string;
  privacyPolicyUrl?: string | null;
  websiteUrl?: string | null;
}

const PublicTrustCenterQuery = graphql`
  query PublicTrustCenterPageQuery($slug: String!) {
    trustCenterBySlug(slug: $slug) {
      id
      active
      slug
      isUserAuthenticated
      organization {
        id
        name
        logoUrl
      }
      documents(first: 100) {
        edges {
          node {
            id
            title
            documentType
          }
        }
      }
      audits(first: 100) {
        edges {
          node {
            id
            framework {
              name
            }
            report {
              id
              filename
              downloadUrl
            }
          }
        }
      }
      vendors(first: 100) {
        edges {
          node {
            id
            name
            category
            websiteUrl
            privacyPolicyUrl
          }
        }
      }
    }
  }
`;

function PublicTrustCenterContent() {
  const { __ } = useTranslate();
  const { slug } = useParams<{ slug: string }>();

  if (!slug) {
    return <Navigate to="/" replace />;
  }

  const data = useLazyLoadQuery<PublicTrustCenterPageQuery>(PublicTrustCenterQuery, { slug });

  const organization = data?.trustCenterBySlug?.organization;
  const organizationName = organization?.name || "";

  const isUserAuthenticated = data?.trustCenterBySlug?.isUserAuthenticated ?? false;

  usePageTitle(
    organizationName ? `${organizationName} - Trust Center` : "Trust Center"
  );

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

  const { documents, audits, vendors } = data.trustCenterBySlug;

  const trustCenterDocuments = documents.edges.map((edge) => edge.node) as TrustCenterDocument[];
  const trustCenterAudits = audits.edges.map((edge) => edge.node) as TrustCenterAudit[];
  const trustCenterVendors = vendors.edges.map((edge) => edge.node) as TrustCenterVendor[];

  return (
    <PublicTrustCenterLayout
      organizationName={organizationName}
      organizationLogo={organization?.logoUrl}
      isAuthenticated={isUserAuthenticated}
    >
      <div className="space-y-12">
        <PublicTrustCenterAudits
          audits={trustCenterAudits}
          organizationName={organizationName}
          isAuthenticated={isUserAuthenticated}
          trustCenterId={data.trustCenterBySlug.id}
        />
        <PublicTrustCenterDocuments
          documents={trustCenterDocuments}
          organizationName={organizationName}
          isAuthenticated={isUserAuthenticated}
          trustCenterId={data.trustCenterBySlug.id}
        />
        <PublicTrustCenterVendors
          vendors={trustCenterVendors}
          organizationName={organizationName}
        />
      </div>
    </PublicTrustCenterLayout>
  );
}

export default function PublicTrustCenterPage() {
  return (
    <TrustRelayProvider>
      <Suspense fallback={<Spinner />}>
        <PublicTrustCenterContent />
      </Suspense>
    </TrustRelayProvider>
  );
}
