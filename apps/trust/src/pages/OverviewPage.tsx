import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import type { OverviewFragment$key } from "./__generated__/OverviewFragment.graphql";
import { Link, useOutletContext } from "react-router";
import { groupBy, objectEntries, sprintf } from "@probo/helpers";
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
  fragment OverviewFragment on TrustCenter {
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
    trustCenter: OverviewFragment$key &
      TrustGraphQuery$data["trustCenterBySlug"];
  }>();
  const { __ } = useTranslate();
  const fragment = useFragment(overviewFragment, trustCenter);
  const documentsPerType = groupBy(
    fragment.documents.edges.map((edge) => edge.node),
    (node) => documentTypeLabel(node.documentType, __),
  );
  return (
    <div>
      <h2 className="font-medium mb-1">{__("Documents")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {__("Security and compliance documentation:")}
      </p>
      <Rows className="mb-8">
        <RowHeader>{__("Certifications")}</RowHeader>
        {trustCenter.audits.edges.map((edge) => (
          <AuditRow key={edge.node.id} audit={edge.node} />
        ))}
        {objectEntries(documentsPerType).map(([label, documents]) => (
          <Fragment key={label}>
            <RowHeader>{label}</RowHeader>
            {documents.map((document) => (
              <DocumentRow key={document.id} document={document} />
            ))}
          </Fragment>
        ))}
        <Link
          to={`/trust/${trustCenter.slug}/documents`}
          className="text-sm font-medium flex gap-2 items-center"
        >
          {__("See all documents")}
          <IconChevronRight size={16} />
        </Link>
      </Rows>

      <h2 className="font-medium mb-1">{__("Subprocessors")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {sprintf(
          __("Third-party subprocessors %s work with:"),
          trustCenter.organization.name,
        )}
      </p>
      <Rows className="mb-8 *:py-5">
        {fragment.vendors.edges.map((edge) => (
          <VendorRow key={edge.node.id} vendor={edge.node} />
        ))}
        <Link
          to={`/trust/${trustCenter.slug}/subprocessors`}
          className="text-sm font-medium flex gap-2 items-center"
        >
          {__("See all subprocessors")}
          <IconChevronRight size={16} />
        </Link>
      </Rows>

      <References
        references={fragment.references.edges.map((edge) => edge.node)}
      />
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
  return (
    <div>
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
