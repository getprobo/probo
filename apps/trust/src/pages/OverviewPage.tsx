// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import {
  getCompliancePortalUrl,
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
import { CompliancePortalFileRow } from "#/components/CompliancePortalFileRow";
import { documentTypeLabel } from "#/helpers/documents";
import type { CompliancePortalGraphCurrentQuery$data } from "#/queries/__generated__/CompliancePortalGraphCurrentQuery.graphql";

import type {
  OverviewPageFragment$data,
  OverviewPageFragment$key,
} from "./__generated__/OverviewPageFragment.graphql";

const overviewFragment = graphql`
  fragment OverviewPageFragment on CompliancePortal {
    references(first: 14) {
      edges {
        node {
          id
          name
          logo {
            downloadUrl
          }
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
          documentType
          ...DocumentRowFragment
        }
      }
    }
    compliancePortalFiles(first: 5) {
      edges {
        node {
          id
          category
          ...CompliancePortalFileRowFragment
        }
      }
    }
  }
`;

export function OverviewPage() {
  const { compliancePortal } = useOutletContext<{
    compliancePortal: OverviewPageFragment$key
      & CompliancePortalGraphCurrentQuery$data["currentCompliancePortal"];
  }>();
  const fragment = useFragment(overviewFragment, compliancePortal);
  return (
    <div>
      <References
        references={fragment.references.edges.map(edge => edge.node)}
      />
      <Documents
        audits={compliancePortal.audits.edges}
        documents={fragment.documents.edges}
        files={fragment.compliancePortalFiles.edges}
        url={getCompliancePortalUrl("documents")}
      />
      <Subprocessors
        title={compliancePortal.title}
        subprocessors={fragment.subprocessors.edges}
        url={getCompliancePortalUrl("subprocessors")}
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
  files: OverviewPageFragment$data["compliancePortalFiles"]["edges"];
  audits: NonNullable<
    CompliancePortalGraphCurrentQuery$data["currentCompliancePortal"]
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
              <CompliancePortalFileRow key={file.id} file={file} />
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
  title,
}: {
  subprocessors: OverviewPageFragment$data["subprocessors"]["edges"];
  url: string;
  title: string;
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
          title,
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
  logo: { downloadUrl: string } | null;
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
              src={reference.logo?.downloadUrl}
              alt={reference.name}
              className="rounded-2xl size-12 block"
            />
            <span className="text-xs text-txt-secondary text-center">{reference.name}</span>
          </a>
        ))}
      </Card>
    </div>
  );
}
