import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import type { OverviewFragment$key } from "./__generated__/OverviewFragment.graphql";
import { Link, useOutletContext } from "react-router";
import { groupBy, objectEntries, sprintf } from "@probo/helpers";
import type { TrustGraphQuery$data } from "/queries/__generated__/TrustGraphQuery.graphql";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  IconChevronRight,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { AuditRow } from "/components/AuditRow";
import { documentTypeLabel } from "/helpers/documents";
import { Fragment } from "react";
import { DocumentRow } from "/components/DocumentRow";
import { VendorRow } from "/components/VendorRow";

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

export function Overview() {
  const { trustCenter } = useOutletContext<{
    trustCenter: OverviewFragment$key &
      TrustGraphQuery$data["trustCenterBySlug"];
  }>();
  const { __ } = useTranslate();
  const fragment = useFragment(overviewFragment, trustCenter);
  const documentsPerType = groupBy(
    fragment.documents.edges.map((edge) => edge.node),
    (node) => documentTypeLabel(node.documentType, __)
  );
  return (
    <div>
      <h2 className="font-medium mb-1">{__("Documents")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {__("Security and compliance documentation:")}
      </p>
      <Table className="mb-8">
        <Thead>
          <Tr className="bg-subtle">
            <Th colSpan={2}>{__("Certifications")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {trustCenter.audits.edges.map((edge) => (
            <AuditRow key={edge.node.id} audit={edge.node} />
          ))}
        </Tbody>
        {objectEntries(documentsPerType).map(([label, documents]) => (
          <Fragment key={label}>
            <Thead>
              <Tr className="bg-subtle">
                <Th colSpan={2}>{label}</Th>
              </Tr>
            </Thead>
            <Tbody>
              {documents.map((document) => (
                <DocumentRow key={document.id} document={document} />
              ))}
            </Tbody>
          </Fragment>
        ))}
        <Tr>
          <Td colSpan={2}>
            <Link
              to={`/trust/${trustCenter.slug}/documents`}
              className="text-sm font-medium flex gap-2 items-center h-8"
            >
              {__("See all documents")}
              <IconChevronRight size={16} />
            </Link>
          </Td>
        </Tr>
      </Table>

      <h2 className="font-medium mb-1">{__("Subprocessors")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {sprintf(
          __("Third-party subprocessors %s work with:"),
          trustCenter.organization.name
        )}
      </p>
      <Table className="mb-8">
        {fragment.vendors.edges.map((edge) => (
          <VendorRow key={edge.node.id} vendor={edge.node} />
        ))}
        <Tr>
          <Td colSpan={3}>
            <Link
              to={`/trust/${trustCenter.slug}/subprocessors`}
              className="text-sm font-medium flex gap-2 items-center h-8"
            >
              {__("See all subprocessors")}
              <IconChevronRight size={16} />
            </Link>
          </Td>
        </Tr>
      </Table>

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
