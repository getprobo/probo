// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { FindingAuditsSection_finding$key } from "#/__generated__/core/FindingAuditsSection_finding.graphql";
import type { FindingAuditsSectionLinkMutation } from "#/__generated__/core/FindingAuditsSectionLinkMutation.graphql";
import type { FindingAuditsSectionPaginationQuery } from "#/__generated__/core/FindingAuditsSectionPaginationQuery.graphql";
import type { FindingAuditsSectionUnlinkMutation } from "#/__generated__/core/FindingAuditsSectionUnlinkMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { FindingAuditCombobox, type FindingAuditOption } from "./FindingAuditCombobox";
import { findingsPage } from "./variants";

const findingAuditsSectionFragment = graphql`
  fragment FindingAuditsSection_finding on Finding
  @refetchable(queryName: "FindingAuditsSectionPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    id
    canLinkAudit: permission(action: "core:finding:create-audit-mapping")
    canUnlinkAudit: permission(action: "core:finding:delete-audit-mapping")
    audits(first: $first, after: $after)
      @connection(key: "FindingAuditsSection_audits")
      @required(action: THROW) {
      __id
      edges {
        referenceId
        node {
          id
          name
        }
      }
    }
  }
`;

const linkAuditMutation = graphql`
  mutation FindingAuditsSectionLinkMutation(
    $input: CreateFindingAuditMappingInput!
    $connections: [ID!]!
  ) {
    createFindingAuditMapping(input: $input) {
      auditEdge @prependEdge(connections: $connections) {
        referenceId
        node {
          id
          name
        }
      }
    }
  }
`;

const unlinkAuditMutation = graphql`
  mutation FindingAuditsSectionUnlinkMutation(
    $input: DeleteFindingAuditMappingInput!
    $connections: [ID!]!
  ) {
    deleteFindingAuditMapping(input: $input) {
      deletedAuditId @deleteEdge(connections: $connections)
    }
  }
`;

interface FindingAuditsSectionProps {
  findingKey: FindingAuditsSection_finding$key;
}

export function FindingAuditsSection({ findingKey }: FindingAuditsSectionProps) {
  const { t } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const {
    data: finding,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    FindingAuditsSectionPaginationQuery,
    FindingAuditsSection_finding$key
  >(findingAuditsSectionFragment, findingKey);
  const [linkAudit, isLinking] = useMutation<FindingAuditsSectionLinkMutation>(linkAuditMutation, {
    errorToast: t("details.audits.linkError"),
  });
  const [unlinkAudit, isUnlinking] = useMutation<FindingAuditsSectionUnlinkMutation>(unlinkAuditMutation, {
    errorToast: t("details.audits.unlinkError"),
  });
  const [pending, setPending] = useState<FindingAuditOption | null>(null);
  const [referenceId, setReferenceId] = useState("");
  const { links, linkRow, linkText, section, pendingLink, pendingActions } = findingsPage();
  const audits = finding.audits.edges.map(edge => ({
    id: edge.node.id,
    name: edge.node.name?.trim() || t("details.audits.untitled"),
    referenceId: edge.referenceId,
  }));
  const linkedIds = new Set(audits.map(audit => audit.id));
  const busy = isLinking || isUnlinking;

  async function confirmLink() {
    const reference = referenceId.trim();
    if (pending == null || reference === "") {
      return;
    }
    try {
      await linkAudit({
        variables: {
          input: {
            findingId: finding.id,
            auditId: pending.id,
            referenceId: reference,
          },
          connections: [finding.audits.__id],
        },
      });
      setPending(null);
      setReferenceId("");
    } catch {
      // The mutation hook already reported the failure.
    }
  }

  return (
    <Card variant="soft" size={2}>
      <div className={section()}>
        <Heading level={2} size={2} weight="medium">
          {t("details.audits.title")}
        </Heading>
        {finding.canLinkAudit && pending == null && (
          <FindingAuditCombobox
            excludeIds={linkedIds}
            disabled={busy}
            onSelect={setPending}
          />
        )}
        {pending != null && (
          <div className={pendingLink()}>
            <Text size={2}>{pending.name}</Text>
            <Field label={t("details.audits.reference")}>
              <TextField
                value={referenceId}
                placeholder={t("details.audits.referencePlaceholder")}
                disabled={busy}
                onValueChange={setReferenceId}
              />
            </Field>
            <div className={pendingActions()}>
              <Button
                variant="ghost"
                color="neutral"
                disabled={busy}
                onClick={() => {
                  setPending(null);
                  setReferenceId("");
                }}
              >
                {t("details.audits.cancel")}
              </Button>
              <Button
                variant="solid"
                disabled={busy || referenceId.trim() === ""}
                loading={isLinking}
                onClick={() => {
                  void confirmLink();
                }}
              >
                {t("details.audits.link")}
              </Button>
            </div>
          </div>
        )}
        {audits.length === 0
          ? <Text size={2} color="faint">{t("details.audits.empty")}</Text>
          : (
              <div className={links()}>
                {audits.map(audit => (
                  <div key={audit.id} className={linkRow()}>
                    <div className={linkText()}>
                      <Link
                        to={`/organizations/${organizationId}/governance/audits/${audit.id}`}
                        size={2}
                      >
                        {audit.name}
                      </Link>
                      <Text size={1} color="faint">{audit.referenceId}</Text>
                    </div>
                    {finding.canUnlinkAudit && (
                      <IconButton
                        variant="ghost"
                        color="neutral"
                        aria-label={t("details.audits.unlink")}
                        disabled={busy}
                        onClick={() => {
                          void unlinkAudit({
                            variables: {
                              input: { findingId: finding.id, auditId: audit.id },
                              connections: [finding.audits.__id],
                            },
                          }).catch(() => {
                            // The mutation hook already reported the failure.
                          });
                        }}
                      >
                        <XIcon />
                      </IconButton>
                    )}
                  </div>
                ))}
              </div>
            )}
        {hasNext && (
          <Button
            variant="soft"
            color="neutral"
            disabled={isLoadingNext}
            onClick={() => loadNext(50)}
          >
            {t("details.audits.loadMore")}
          </Button>
        )}
      </div>
    </Card>
  );
}
