// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
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
  const dialogRef = useRef<{ open: () => void; close: () => void }>(null);

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title="Link Statement of Applicability"
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
  const { __ } = useTranslate();
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
                  {__("No statements of applicability available")}
                </div>
                <div className="text-txt-tertiary text-sm">
                  {__(
                    "Create a statement of applicability first to link it to this control",
                  )}
                </div>
              </div>
            )
          : !selectedSOA
              ? (
                  <div className="space-y-2">
                    <div className="text-sm font-medium mb-2">
                      {__("Select a statement of applicability:")}
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
                                    {__("Linked")}
                                  </Badge>
                                  <Button
                                    variant="danger"
                                    onClick={() =>
                                      handleUnlink(soa.id)}
                                    disabled={props.disabled}
                                  >
                                    {__("Unlink")}
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
                          {__("Selected:")}
                        </div>
                        <div className="text-lg font-medium">
                          {selectedSOA.name}
                        </div>
                      </div>
                      <Button
                        variant="tertiary"
                        onClick={() => setSelectedSOA(null)}
                      >
                        {__("Change")}
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
                          {__("Applicable")}
                        </span>
                      </label>

                      <div>
                        <label className="text-sm font-medium mb-1 block">
                          {__("Justification (optional)")}
                        </label>
                        <Textarea
                          placeholder={__("Add a justification...")}
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
      <DialogFooter exitLabel={__("Close")}>
        {selectedSOA
          ? (
              <>
                <Button
                  variant="secondary"
                  onClick={() => setSelectedSOA(null)}
                >
                  {__("Back")}
                </Button>
                <Button
                  variant="primary"
                  onClick={handleLink}
                  disabled={props.disabled}
                >
                  {__("Link")}
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
