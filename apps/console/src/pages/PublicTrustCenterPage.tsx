import { useParams, Navigate } from "react-router";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import { publicTrustCenterQuery } from "/hooks/graph/PublicTrustCenterGraph";
import { PublicTrustCenterLayout } from "/layouts/PublicTrustCenterLayout";
import { PublicTrustCenterAudits } from "../components/trustCenter/PublicTrustCenterAudits";
import { PublicTrustCenterVendors } from "../components/trustCenter/PublicTrustCenterVendors";
import { PublicTrustCenterDocuments } from "../components/trustCenter/PublicTrustCenterDocuments";
import { PageError } from "/components/PageError";
import { TrustRelayProvider } from "/providers/TrustRelayProvider";
import { useState, useEffect } from "react";
import { buildEndpoint } from "/providers/RelayProviders";
import { Spinner } from "@probo/ui";
interface GraphQLError {
  message: string;
  path?: string[];
  locations?: Array<{ line: number; column: number }>;
}

interface GraphQLResponse<T> {
  data?: T;
  errors?: GraphQLError[];
}

interface Organization {
  id: string;
  name: string;
  logoUrl?: string;
}

interface Framework {
  name: string;
}

interface DocumentVersion {
  id: string;
  status: string;
}

interface DocumentVersionConnection {
  edges: Array<{
    node: DocumentVersion;
  }>;
}

interface Document {
  id: string;
  title: string;
  documentType: string;
  versions: DocumentVersionConnection;
}

interface Audit {
  id: string;
  framework: Framework;
  validFrom: string;
  validUntil: string | null;
  state: string;
  createdAt: string;
  report: {
    id: string;
    filename: string;
    downloadUrl: string | null;
  } | null;
  reportUrl: string | null;
}

interface Vendor {
  id: string;
  name: string;
  category: string;
  description: string | null;
  createdAt: string;
  websiteUrl?: string | null;
  privacyPolicyUrl?: string | null;
}

interface Connection<T> {
  edges: Array<{
    node: T;
  }>;
}

interface TrustCenter {
  id: string;
  active: boolean;
  slug: string;
  organization: Organization;
  documents: Connection<Document>;
  audits: Connection<Audit>;
  vendors: Connection<Vendor>;
}

interface PublicTrustCenterData {
  trustCenterBySlug?: TrustCenter;
}

function PublicTrustCenterContent() {
  const { __ } = useTranslate();
  const { slug } = useParams<{ slug: string }>();
  const [data, setData] = useState<PublicTrustCenterData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  const organizationName = data?.trustCenterBySlug?.organization?.name;
  usePageTitle(
    organizationName ? `${organizationName} - Trust Center` : "Trust Center"
  );

  useEffect(() => {
    if (!slug) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setIsAuthenticated(false);

    fetch(buildEndpoint("/api/trust/v1/graphql"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        operationName: publicTrustCenterQuery.params.name,
        query: publicTrustCenterQuery.params.text,
        variables: { slug },
      }),
    })
      .then((response) => response.json())
      .then((result: GraphQLResponse<PublicTrustCenterData>) => {
        const accessDeniedErrors =
          result.errors?.filter((error: GraphQLError) =>
            error.message.includes("access denied")
          ) || [];

        const nonAuthErrors =
          result.errors?.filter(
            (error: GraphQLError) => !error.message.includes("access denied")
          ) || [];

        if (nonAuthErrors.length > 0) {
          throw new Error(nonAuthErrors[0].message);
        }

        setIsAuthenticated(accessDeniedErrors.length === 0);
        setData(result.data || null);
      })
      .catch(setError)
      .finally(() => setLoading(false));
  }, [slug]);

  if (!slug) {
    return <Navigate to="/" replace />;
  }

  if (loading) {
    return <Spinner />;
  }

  if (error) {
    return <PageError />;
  }

  const { trustCenterBySlug } = data || {};

  if (!trustCenterBySlug) {
    return <PageError />;
  }

  if (!trustCenterBySlug.active) {
    return <PageError />;
  }

  const { organization } = trustCenterBySlug;

  const documents = trustCenterBySlug.documents.edges.map((edge) => edge.node);

  const audits = trustCenterBySlug.audits.edges.map((edge) => edge.node);

  const vendors = trustCenterBySlug.vendors.edges.map((edge) => edge.node);

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
            {__(
              "Explore our security practices, compliance certifications, and transparency reports."
            )}
          </p>
        </div>
        <PublicTrustCenterAudits
          audits={audits}
          organizationName={organization.name}
          isAuthenticated={isAuthenticated}
        />
        <PublicTrustCenterDocuments
          documents={documents}
          isAuthenticated={isAuthenticated}
        />
        <PublicTrustCenterVendors
          vendors={vendors}
          organizationName={organization.name}
        />
      </div>
    </PublicTrustCenterLayout>
  );
}

export default function PublicTrustCenterPage() {
  return (
    <TrustRelayProvider>
      <PublicTrustCenterContent />
    </TrustRelayProvider>
  );
}
