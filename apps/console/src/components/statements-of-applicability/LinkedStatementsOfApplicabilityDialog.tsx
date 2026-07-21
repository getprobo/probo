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

import {
  Badge,
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Textarea,
} from "@probo/ui";
import type { ReactNode } from "react";
import { Suspense, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedStatementsOfApplicabilityDialogQuery } from "#/__generated__/core/LinkedStatementsOfApplicabilityDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const query = graphql`
    query LinkedStatementsOfApplicabilityDialogQuery($organizationId: ID!) {
        organization: node(id: $organizationId) {
            ... on Organization {
                statementsOfApplicability(first: 100) {
                    edges {
                        node {
                            id
                            name
                        }
                    }
                }
            }
        }
    }
`;

type LinkedSOAInfo = {
  statementOfApplicabilityId: string;
  controlId: string;
};

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedStatementsOfApplicability: readonly LinkedSOAInfo[];
  onLink: (
    statementOfApplicabilityId: string,
    applicability: boolean,
    justification: string | null,
  ) => void;
  onUnlink: (statementOfApplicabilityId: string, controlId: string) => void;
};

export function LinkedStatementsOfApplicabilityDialog({
  children,
  ...props
}: Props) {
  const { t } = useTranslation();
  const dialogRef = useRef<{ open: () => void; close: () => void }>(null);

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={t("linkedStatementsOfApplicabilityDialog.title")}
    >
      <Suspense fallback={<div>Loading...</div>}>
        <LinkedStatementsOfApplicabilityDialogContent
          {...props}
          onClose={() => dialogRef.current?.close()}
        />
      </Suspense>
    </Dialog>
  );
}

function LinkedStatementsOfApplicabilityDialogContent(
  props: Omit<Props, "children"> & { onClose: () => void },
) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const [selectedSOA, setSelectedSOA] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [applicability, setApplicability] = useState(true);
  const [justification, setJustification] = useState("");

  const data = useLazyLoadQuery<LinkedStatementsOfApplicabilityDialogQuery>(
    query,
    {
      organizationId,
    },
    { fetchPolicy: "network-only" },
  );

  const linkedSOAIds = new Set(
    props.linkedStatementsOfApplicability.map(
      soa => soa.statementOfApplicabilityId,
    ),
  );
  const linkedSOAMap = new Map(
    props.linkedStatementsOfApplicability.map(soa => [
      soa.statementOfApplicabilityId,
      soa,
    ]),
  );
  const statementsOfApplicability
    = data.organization?.statementsOfApplicability?.edges.map(
      edge => edge.node,
    ) ?? [];

  const handleSelectSOA = (soa: { id: string; name: string }) => {
    setSelectedSOA(soa);
    setApplicability(true);
    setJustification("");
  };

  const handleLink = () => {
    if (selectedSOA) {
      props.onLink(
        selectedSOA.id,
        applicability,
        justification.trim() || null,
      );
      props.onClose();
    }
  };

  const handleUnlink = (statementOfApplicabilityId: string) => {
    const linkedSOA = linkedSOAMap.get(statementOfApplicabilityId);
    if (linkedSOA) {
      props.onUnlink(
        linkedSOA.statementOfApplicabilityId,
        linkedSOA.controlId,
      );
    }
  };

  return (
    <>
      <DialogContent padded className="space-y-4">
        {statementsOfApplicability.length === 0
          ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <div className="text-txt-secondary text-base mb-2">
                  {t("linkedStatementsOfApplicabilityDialog.empty.title")}
                </div>
                <div className="text-txt-tertiary text-sm">
                  {t("linkedStatementsOfApplicabilityDialog.empty.description")}
                </div>
              </div>
            )
          : !selectedSOA
              ? (
                  <div className="space-y-2">
                    <div className="text-sm font-medium mb-2">
                      {t("linkedStatementsOfApplicabilityDialog.select")}
                    </div>
                    {statementsOfApplicability.map((soa) => {
                      const isLinked = linkedSOAIds.has(soa.id);
                      return (
                        <div
                          key={soa.id}
                          className={`border border-border-low rounded-lg p-3 flex items-center justify-between ${!isLinked ? "hover:bg-hover cursor-pointer" : ""}`}
                          onClick={() =>
                            !isLinked && handleSelectSOA(soa)}
                        >
                          <div className="font-medium">
                            {soa.name}
                          </div>
                          {isLinked
                            ? (
                                <div
                                  className="flex items-center gap-2"
                                  onClick={e => e.stopPropagation()}
                                >
                                  <Badge variant="success">
                                    {t("linkedStatementsOfApplicabilityDialog.linked")}
                                  </Badge>
                                  <Button
                                    variant="danger"
                                    onClick={() =>
                                      handleUnlink(soa.id)}
                                    disabled={props.disabled}
                                  >
                                    {t("linkedStatementsOfApplicabilityDialog.actions.unlink")}
                                  </Button>
                                </div>
                              )
                            : null}
                        </div>
                      );
                    })}
                  </div>
                )
              : (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-sm text-txt-secondary mb-1">
                          {t("linkedStatementsOfApplicabilityDialog.selected")}
                        </div>
                        <div className="text-lg font-medium">
                          {selectedSOA.name}
                        </div>
                      </div>
                      <Button
                        variant="tertiary"
                        onClick={() => setSelectedSOA(null)}
                      >
                        {t("linkedStatementsOfApplicabilityDialog.actions.change")}
                      </Button>
                    </div>

                    <div className="border-t border-border-low pt-4 space-y-3">
                      <label className="flex items-center gap-2 cursor-pointer">
                        <Checkbox
                          checked={applicability}
                          onChange={checked =>
                            setApplicability(checked)}
                        />
                        <span className="font-medium">
                          {t("linkedStatementsOfApplicabilityDialog.applicable")}
                        </span>
                      </label>

                      <div>
                        <label className="text-sm font-medium mb-1 block">
                          {t("linkedStatementsOfApplicabilityDialog.justification.label")}
                        </label>
                        <Textarea
                          placeholder={t("linkedStatementsOfApplicabilityDialog.justification.placeholder")}
                          value={justification}
                          onChange={e =>
                            setJustification(e.target.value)}
                          rows={3}
                        />
                      </div>
                    </div>
                  </div>
                )}
      </DialogContent>
      <DialogFooter exitLabel={t("linkedStatementsOfApplicabilityDialog.actions.close")}>
        {selectedSOA
          ? (
              <>
                <Button
                  variant="secondary"
                  onClick={() => setSelectedSOA(null)}
                >
                  {t("linkedStatementsOfApplicabilityDialog.actions.back")}
                </Button>
                <Button
                  variant="primary"
                  onClick={handleLink}
                  disabled={props.disabled}
                >
                  {t("linkedStatementsOfApplicabilityDialog.actions.link")}
                </Button>
              </>
            )
          : (
              <></>
            )}
      </DialogFooter>
    </>
  );
}
