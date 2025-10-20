import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import type {
  OverviewPageFragment$data,
  OverviewPageFragment$key,
} from "./__generated__/OverviewPageFragment.graphql";
import { Link, useOutletContext } from "react-router";
import {
  groupBy,
  objectEntries,
  sprintf,
  getTrustCenterUrl,
} from "@probo/helpers";
import type { TrustGraphQuery$data } from "/queries/__generated__/TrustGraphQuery.graphql";
import { useTranslate } from "@probo/i18n";
import { Card, IconChevronRight } from "@probo/ui";
import { AuditRow } from "/components/AuditRow";
import { documentTypeLabel } from "/helpers/documents";
import { Fragment } from "react";
import { DocumentRow } from "/components/DocumentRow";
import { VendorRow } from "/components/VendorRow";
import { RowHeader } from "/components/RowHeader.tsx";
import { Rows } from "/components/Rows.tsx";

const overviewFragment = graphql`
  fragment OverviewPageFragment on TrustCenter {
    references(first: 14) {
      edges {
        node {
          id
          name
          logoUrl
          websiteUrl
        }
      }
    }
    vendors(first: 3) {
      edges {
        node {
          id
          countries
          ...VendorRowFragment
        }
      }
    }
    documents(first: 5) {
      edges {
        node {
          id
          ...DocumentRowFragment
          documentType
        }
      }
    }
  }
`;

export function OverviewPage() {
  const { trustCenter } = useOutletContext<{
    trustCenter: OverviewPageFragment$key &
      TrustGraphQuery$data["trustCenterBySlug"];
  }>();
  const fragment = useFragment(overviewFragment, trustCenter);
  return (
    <div>
      <References
        references={fragment.references.edges.map((edge) => edge.node)}
      />
      <Documents
        audits={trustCenter.audits.edges}
        documents={fragment.documents.edges}
        url={getTrustCenterUrl("documents")}
      />
      <Subprocessors
        organizationName={trustCenter.organization.name}
        vendors={fragment.vendors.edges}
        url={getTrustCenterUrl("subprocessors")}
      />
    </div>
  );
}

function Documents({
  documents,
  audits,
  url,
}: {
  documents: OverviewPageFragment$data["documents"]["edges"];
  audits: NonNullable<
    TrustGraphQuery$data["trustCenterBySlug"]
  >["audits"]["edges"];
  url: string;
}) {
  const { __ } = useTranslate();
  const documentsPerType = groupBy(
    documents.map((edge) => edge.node),
    (node) => documentTypeLabel(node.documentType, __)
  );
  const hasAudits = audits.length > 0;
  const hasDocuments = hasAudits || documents.length > 0;

  if (!hasDocuments) {
    return null;
  }

  return (
    <div>
      <h2 className="font-medium mb-1">{__("Documents")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {__("Security and compliance documentation:")}
      </p>
      <Rows className="mb-8">
        {audits.length > 0 && (
          <>
            <RowHeader>{__("Compliance")}</RowHeader>
            {audits.map((audit) => (
              <AuditRow key={audit.node.id} audit={audit.node} />
            ))}
          </>
        )}
        {objectEntries(documentsPerType).map(([label, documents]) => (
          <Fragment key={label}>
            <RowHeader>{label}</RowHeader>
            {documents.map((document) => (
              <DocumentRow key={document.id} document={document} />
            ))}
          </Fragment>
        ))}
        <Link to={url} className="text-sm font-medium flex gap-2 items-center">
          {__("See all documents")}
          <IconChevronRight size={16} />
        </Link>
      </Rows>
    </div>
  );
}

function Subprocessors({
  vendors,
  url,
  organizationName,
}: {
  vendors: OverviewPageFragment$data["vendors"]["edges"];
  url: string;
  organizationName: string;
}) {
  const { __ } = useTranslate();
  if (vendors.length === 0) {
    return null;
  }

  const hasAnyCountries = vendors.some((vendor) => {
    const vendorData = vendor.node as any;
    return vendorData.countries && vendorData.countries.length > 0;
  });

  return (
    <div>
      <h2 className="font-medium mb-1">{__("Subprocessors")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {sprintf(
          __("Third-party subprocessors %s work with:"),
          organizationName
        )}
      </p>
      <Rows className="mb-8 *:py-5">
        {vendors.map((vendor) => (
          <VendorRow
            key={vendor.node.id}
            vendor={vendor.node}
            hasAnyCountries={hasAnyCountries}
          />
        ))}
        <Link to={url} className="text-sm font-medium flex gap-2 items-center">
          {__("See all subprocessors")}
          <IconChevronRight size={16} />
        </Link>
      </Rows>
    </div>
  );
}

type Reference = {
  name: string;
  logoUrl: string;
  websiteUrl: string;
  id: string;
};

function References({ references }: { references: Reference[] }) {
  const { __ } = useTranslate();

  if (references.length === 0) {
    return null;
  }

  return (
    <div className="mb-8">
      <h2 className="font-medium mb-4">{__("Trusted by")}</h2>
      <Card className="grid grid-cols-2 flex-wrap p-6 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-7">
        {references.map((reference) => (
          <a
            key={reference.id}
            href={reference.websiteUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="flex flex-col justify-center items-center gap-2"
          >
            <img
              src={reference.logoUrl}
              alt={reference.name}
              className="rounded-2xl size-12 block"
            />
            <span className="text-xs text-txt-secondary">{reference.name}</span>
          </a>
        ))}
      </Card>
    </div>
  );
}
