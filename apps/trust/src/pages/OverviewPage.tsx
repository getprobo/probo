// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import {
  getTrustCenterUrl,
  groupBy,
  objectEntries,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Card, IconChevronRight } from "@probo/ui";
import { Fragment } from "react";
import { useFragment } from "react-relay";
import { Link, useOutletContext } from "react-router";
import { graphql } from "relay-runtime";

import { AuditRow } from "#/components/AuditRow";
import { DocumentRow } from "#/components/DocumentRow";
import { RowHeader } from "#/components/RowHeader";
import { Rows } from "#/components/Rows";
import { SubprocessorRow } from "#/components/SubprocessorRow";
import { TrustCenterFileRow } from "#/components/TrustCenterFileRow";
import { documentTypeLabel } from "#/helpers/documents";
import type { TrustGraphCurrentQuery$data } from "#/queries/__generated__/TrustGraphCurrentQuery.graphql";

import type {
  OverviewPageFragment$data,
  OverviewPageFragment$key,
} from "./__generated__/OverviewPageFragment.graphql";

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
    subprocessors(first: 3) {
      edges {
        node {
          id
          countries
          ...SubprocessorRowFragment
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
    trustCenterFiles(first: 5) {
      edges {
        node {
          id
          category
          ...TrustCenterFileRowFragment
        }
      }
    }
  }
`;

export function OverviewPage() {
  const { trustCenter } = useOutletContext<{
    trustCenter: OverviewPageFragment$key
      & TrustGraphCurrentQuery$data["currentTrustCenter"];
  }>();
  const fragment = useFragment(overviewFragment, trustCenter);
  return (
    <div>
      <References
        references={fragment.references.edges.map(edge => edge.node)}
      />
      <Documents
        audits={trustCenter.audits.edges}
        documents={fragment.documents.edges}
        files={fragment.trustCenterFiles.edges}
        url={getTrustCenterUrl("documents")}
      />
      <Subprocessors
        organizationName={trustCenter.organization.name}
        subprocessors={fragment.subprocessors.edges}
        url={getTrustCenterUrl("subprocessors")}
      />
    </div>
  );
}

function Documents({
  documents,
  files,
  audits,
  url,
}: {
  documents: OverviewPageFragment$data["documents"]["edges"];
  files: OverviewPageFragment$data["trustCenterFiles"]["edges"];
  audits: NonNullable<
    TrustGraphCurrentQuery$data["currentTrustCenter"]
  >["audits"]["edges"];
  url: string;
}) {
  const { __ } = useTranslate();
  const documentsPerType = groupBy(
    documents.map(edge => edge.node),
    node => documentTypeLabel(node.documentType, __),
  );
  const filesPerCategory = groupBy(
    files.map(edge => edge.node),
    node => node.category,
  );
  const hasAudits = audits.length > 0;
  const hasDocuments = hasAudits || documents.length > 0 || files.length > 0;

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
            {audits.map(audit => (
              <AuditRow key={audit.node.id} audit={audit.node} />
            ))}
          </>
        )}
        {objectEntries(documentsPerType).map(([label, documents]) => (
          <Fragment key={label}>
            <RowHeader>{label}</RowHeader>
            {documents.map(document => (
              <DocumentRow key={document.id} document={document} />
            ))}
          </Fragment>
        ))}
        {objectEntries(filesPerCategory).map(([category, files]) => (
          <Fragment key={category}>
            <RowHeader>{category}</RowHeader>
            {files.map(file => (
              <TrustCenterFileRow key={file.id} file={file} />
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
  subprocessors,
  url,
  organizationName,
}: {
  subprocessors: OverviewPageFragment$data["subprocessors"]["edges"];
  url: string;
  organizationName: string;
}) {
  const { __ } = useTranslate();
  if (subprocessors.length === 0) {
    return null;
  }

  const hasAnyCountries = subprocessors.some((subprocessor) => {
    const subprocessorData = subprocessor.node;
    return subprocessorData.countries && subprocessorData.countries.length > 0;
  });

  return (
    <div>
      <h2 className="font-medium mb-1">{__("Subprocessors")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {sprintf(
          __("Third-party subprocessors %s work with:"),
          organizationName,
        )}
      </p>
      <Rows className="mb-8 *:py-5">
        {subprocessors.map(subprocessor => (
          <SubprocessorRow
            key={subprocessor.node.id}
            subprocessor={subprocessor.node}
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
        {references.map(reference => (
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
