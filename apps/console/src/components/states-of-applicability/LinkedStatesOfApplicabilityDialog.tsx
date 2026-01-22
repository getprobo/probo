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

import type { LinkedStatesOfApplicabilityDialogQuery } from "#/__generated__/core/LinkedStatesOfApplicabilityDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const query = graphql`
    query LinkedStatesOfApplicabilityDialogQuery($organizationId: ID!) {
        organization: node(id: $organizationId) {
            ... on Organization {
                statesOfApplicability(first: 100) {
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
  stateOfApplicabilityId: string;
  controlId: string;
};

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedStatesOfApplicability: readonly LinkedSOAInfo[];
  onLink: (
    stateOfApplicabilityId: string,
    applicability: boolean,
    justification: string | null,
  ) => void;
  onUnlink: (stateOfApplicabilityId: string, controlId: string) => void;
};

export function LinkedStatesOfApplicabilityDialog({
  children,
  ...props
}: Props) {
  const dialogRef = useRef<{ open: () => void; close: () => void }>(null);

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title="Link State of Applicability"
    >
      <Suspense fallback={<div>Loading...</div>}>
        <LinkedStatesOfApplicabilityDialogContent
          {...props}
          onClose={() => dialogRef.current?.close()}
        />
      </Suspense>
    </Dialog>
  );
}

function LinkedStatesOfApplicabilityDialogContent(
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

  const data = useLazyLoadQuery<LinkedStatesOfApplicabilityDialogQuery>(
    query,
    {
      organizationId,
    },
    { fetchPolicy: "network-only" },
  );

  const linkedSOAIds = new Set(
    props.linkedStatesOfApplicability.map(
      soa => soa.stateOfApplicabilityId,
    ),
  );
  const linkedSOAMap = new Map(
    props.linkedStatesOfApplicability.map(soa => [
      soa.stateOfApplicabilityId,
      soa,
    ]),
  );
  const statesOfApplicability
    = data.organization?.statesOfApplicability?.edges.map(
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

  const handleUnlink = (stateOfApplicabilityId: string) => {
    const linkedSOA = linkedSOAMap.get(stateOfApplicabilityId);
    if (linkedSOA) {
      props.onUnlink(
        linkedSOA.stateOfApplicabilityId,
        linkedSOA.controlId,
      );
    }
  };

  return (
    <>
      <DialogContent padded className="space-y-4">
        {statesOfApplicability.length === 0
          ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <div className="text-txt-secondary text-base mb-2">
                  {__("No states of applicability available")}
                </div>
                <div className="text-txt-tertiary text-sm">
                  {__(
                    "Create a state of applicability first to link it to this control",
                  )}
                </div>
              </div>
            )
          : !selectedSOA
              ? (
                  <div className="space-y-2">
                    <div className="text-sm font-medium mb-2">
                      {__("Select a state of applicability:")}
                    </div>
                    {statesOfApplicability.map((soa) => {
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
