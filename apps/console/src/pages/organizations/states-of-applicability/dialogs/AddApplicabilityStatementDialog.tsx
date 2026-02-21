import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconCheckmark1,
  IconChevronDown,
  IconChevronUp,
  IconMagnifyingGlass,
  Input,
  Option,
  Select,
  Spinner,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, Suspense, useImperativeHandle, useMemo, useState } from "react";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { AddApplicabilityStatementDialogQuery } from "#/__generated__/core/AddApplicabilityStatementDialogQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const query = graphql`
    query AddApplicabilityStatementDialogQuery($stateOfApplicabilityId: ID!, $organizationId: ID!) {
        stateOfApplicability: node(id: $stateOfApplicabilityId) {
            ... on StateOfApplicability {
                id
                applicabilityStatements(first: 10000) {
                    edges {
                        node {
                            id
                            applicability
                            justification
                            control {
                                id
                            }
                        }
                    }
                }
            }
        }
        organization: node(id: $organizationId) {
            ... on Organization {
                id
                controls(first: 10000, orderBy: { direction: ASC, field: CREATED_AT }) {
                    edges {
                        node {
                            id
                            sectionTitle
                            name
                            framework {
                                id
                                name
                            }
                        }
                    }
                }
            }
        }
    }
`;

const createApplicabilityStatementMutation = graphql`
    mutation AddApplicabilityStatementDialogCreateMutation(
        $input: CreateApplicabilityStatementInput!
        $connections: [ID!]!
    ) {
        createApplicabilityStatement(input: $input) {
            applicabilityStatementEdge @appendEdge(connections: $connections) {
                node {
                    id
                    applicability
                    justification
                    control {
                        id
                        sectionTitle
                        name
                        bestPractice
                        regulatory
                        contractual
                        riskAssessment
                        framework {
                            id
                            name
                        }
                        organization {
                            id
                        }
                    }
                }
            }
        }
    }
`;

const deleteApplicabilityStatementMutation = graphql`
    mutation AddApplicabilityStatementDialogDeleteMutation(
        $input: DeleteApplicabilityStatementInput!
        $connections: [ID!]!
    ) {
        deleteApplicabilityStatement(input: $input) {
            deletedApplicabilityStatementId @deleteEdge(connections: $connections)
        }
    }
`;

const updateApplicabilityStatementMutation = graphql`
    mutation AddApplicabilityStatementDialogUpdateMutation(
        $input: UpdateApplicabilityStatementInput!
    ) {
        updateApplicabilityStatement(input: $input) {
            applicabilityStatement {
                id
                applicability
                justification
            }
        }
    }
`;

export type AddApplicabilityStatementDialogRef = {
  open: (stateOfApplicabilityId: string, organizationId: string, connectionId: string) => void;
};

type ControlWithStatement = {
  controlId: string;
  sectionTitle: string;
  name: string;
  frameworkId: string;
  frameworkName: string;
  applicabilityStatementId: string | null;
  applicability: boolean | null;
  justification: string | null;
};

function ControlRow({
  control,
  stateOfApplicabilityId,
  connectionId,
}: {
  control: ControlWithStatement;
  stateOfApplicabilityId: string;
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const isLinked = control.applicabilityStatementId !== null;
  const [selectedState, setSelectedState] = useState<string>(() => {
    if (!isLinked) return "not-linked";
    return control.applicability ? "applicable" : "not-applicable";
  });
  const [justification, setJustification] = useState(control.justification || "");
  const [showJustification, setShowJustification] = useState(false);

  const [createApplicabilityStatement, isCreating] = useMutationWithToasts(
    createApplicabilityStatementMutation,
    {
      successMessage: __("Statement created successfully."),
      errorMessage: __("Failed to create statement"),
    },
  );

  const [deleteApplicabilityStatement, isDeleting] = useMutationWithToasts(
    deleteApplicabilityStatementMutation,
    {
      successMessage: __("Statement removed successfully."),
      errorMessage: __("Failed to remove statement"),
    },
  );

  const [updateApplicabilityStatement, isUpdating] = useMutationWithToasts(
    updateApplicabilityStatementMutation,
    {
      successMessage: __("Statement updated successfully."),
      errorMessage: __("Failed to update statement"),
    },
  );

  const handleStateChange = async (newState: string) => {
    setSelectedState(newState);

    if (newState === "not-linked") {
      if (!control.applicabilityStatementId) return;
      await deleteApplicabilityStatement({
        variables: {
          input: {
            applicabilityStatementId: control.applicabilityStatementId,
          },
          connections: [connectionId],
        },
        onSuccess: () => {
          setShowJustification(false);
        },
      });
    } else if (newState === "applicable") {
      setShowJustification(false);
      if (control.applicabilityStatementId) {
        // Statement already exists, use update mutation
        await updateApplicabilityStatement({
          variables: {
            input: {
              applicabilityStatementId: control.applicabilityStatementId,
              applicability: true,
              justification: null,
            },
          },
        });
      } else {
        // No statement exists, use create mutation
        await createApplicabilityStatement({
          variables: {
            input: {
              stateOfApplicabilityId,
              controlId: control.controlId,
              applicability: true,
              justification: null,
            },
            connections: [connectionId],
          },
        });
      }
    } else if (newState === "not-applicable") {
      setShowJustification(true);
      setJustification(control.justification || "");
    }
  };

  const handleSaveJustification = async () => {
    if (control.applicabilityStatementId) {
      // Statement already exists, use update mutation
      await updateApplicabilityStatement({
        variables: {
          input: {
            applicabilityStatementId: control.applicabilityStatementId,
            applicability: false,
            justification: justification || null,
          },
        },
        onSuccess: () => {
          setShowJustification(false);
        },
      });
    } else {
      // No statement exists, use create mutation
      await createApplicabilityStatement({
        variables: {
          input: {
            stateOfApplicabilityId,
            controlId: control.controlId,
            applicability: false,
            justification: justification || null,
          },
          connections: [connectionId],
        },
        onSuccess: () => {
          setShowJustification(false);
        },
      });
    }
  };

  return (
    <div className="p-4 border-b border-border-low">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Badge size="md">{control.sectionTitle}</Badge>
            <span className="text-sm font-medium text-txt-primary">
              {control.name}
            </span>
          </div>
          {isLinked
            && control.applicability !== null
            && !showJustification
            && control.justification && (
            <div className="mt-2 text-sm text-txt-secondary">
              {control.justification}
            </div>
          )}
        </div>
        <div className="flex items-start gap-2">
          <Select
            variant="editor"
            value={selectedState}
            onValueChange={value => void handleStateChange(value)}
            disabled={isCreating || isDeleting || isUpdating}
            className="w-48"
          >
            <Option value="not-linked">{__("Not Linked")}</Option>
            <Option value="applicable">{__("Applicable")}</Option>
            <Option value="not-applicable">{__("Not Applicable")}</Option>
          </Select>
        </div>
      </div>
      {showJustification && (
        <div className="mt-3 flex items-start gap-2">
          <Textarea
            value={justification}
            onChange={e => setJustification(e.target.value)}
            placeholder={__("Reason for non-applicability")}
            className="flex-1"
            autogrow
          />
          <Button
            variant="primary"
            icon={IconCheckmark1}
            onClick={() => void handleSaveJustification()}
            disabled={isCreating || isUpdating}
            aria-label={__("Save")}
          />
        </div>
      )}
    </div>
  );
}

function AddApplicabilityStatementDialogContent({
  stateOfApplicabilityId,
  organizationId,
  connectionId,
}: {
  stateOfApplicabilityId: string;
  organizationId: string;
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const [collapsedFrameworks, setCollapsedFrameworks] = useState<Set<string>>(new Set());
  const data = useLazyLoadQuery<AddApplicabilityStatementDialogQuery>(
    query,
    { stateOfApplicabilityId, organizationId },
    { fetchPolicy: "store-or-network" },
  );

  const applicabilityMap = useMemo(() => {
    const map = new Map<
      string,
      { id: string; applicability: boolean; justification: string | null }
    >();
    data.stateOfApplicability?.applicabilityStatements?.edges.forEach((edge) => {
      map.set(edge.node.control.id, {
        id: edge.node.id,
        applicability: edge.node.applicability,
        justification: edge.node.justification,
      });
    });
    return map;
  }, [data.stateOfApplicability?.applicabilityStatements]);

  const allControls = useMemo(() => {
    return (data.organization?.controls?.edges || []).map((edge) => {
      const applicability = applicabilityMap.get(edge.node.id);
      return {
        controlId: edge.node.id,
        sectionTitle: edge.node.sectionTitle,
        name: edge.node.name,
        frameworkId: edge.node.framework.id,
        frameworkName: edge.node.framework.name,
        applicabilityStatementId: applicability?.id ?? null,
        applicability: applicability?.applicability ?? null,
        justification: applicability?.justification ?? null,
      } as ControlWithStatement;
    });
  }, [data.organization?.controls, applicabilityMap]);

  const filteredControls = useMemo(() => {
    if (!search) return allControls;
    const lowerSearch = search.toLowerCase();
    return allControls.filter(
      c =>
        c.name.toLowerCase().includes(lowerSearch)
        || c.sectionTitle.toLowerCase().includes(lowerSearch)
        || c.frameworkName.toLowerCase().includes(lowerSearch),
    );
  }, [allControls, search]);

  const groupedControls = useMemo(() => {
    const groups: Record<string, Record<string, ControlWithStatement[]>> = {};
    filteredControls.forEach((control) => {
      if (!groups[control.frameworkName]) {
        groups[control.frameworkName] = {};
      }
      if (!groups[control.frameworkName][control.sectionTitle]) {
        groups[control.frameworkName][control.sectionTitle] = [];
      }
      groups[control.frameworkName][control.sectionTitle].push(control);
    });
    return groups;
  }, [filteredControls]);

  const toggleFramework = (frameworkName: string) => {
    setCollapsedFrameworks((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(frameworkName)) {
        newSet.delete(frameworkName);
      } else {
        newSet.add(frameworkName);
      }
      return newSet;
    });
  };

  return (
    <>
      <DialogContent className="p-0">
        <div className="sticky top-0 bg-level-2 p-4 border-b border-border-low z-10">
          <Input
            icon={IconMagnifyingGlass}
            placeholder={__("Search controls...")}
            onValueChange={setSearch}
          />
        </div>
        <div className="max-h-[60vh] overflow-y-auto">
          {filteredControls.length === 0
            ? (
                <div className="p-8 text-center text-txt-secondary">
                  {__("No controls found")}
                </div>
              )
            : (
                Object.entries(groupedControls).map(([frameworkName, sections]) => {
                  const isCollapsed = collapsedFrameworks.has(frameworkName);
                  return (
                    <div key={frameworkName}>
                      <div className="sticky top-0 bg-level-1 px-4 py-2 border-b border-border-low z-10 flex items-center justify-between">
                        <h3 className="text-sm font-semibold text-txt-primary">
                          {frameworkName}
                        </h3>
                        <Button
                          variant="tertiary"
                          icon={isCollapsed ? IconChevronDown : IconChevronUp}
                          onClick={() => toggleFramework(frameworkName)}
                          aria-label={isCollapsed ? __("Expand") : __("Collapse")}
                        />
                      </div>
                      {!isCollapsed
                        && Object.entries(sections).map(
                          ([sectionTitle, sectionControls]) => (
                            <div key={`${frameworkName}-${sectionTitle}`}>
                              {sectionControls.map(control => (
                                <ControlRow
                                  key={control.controlId}
                                  control={control}
                                  stateOfApplicabilityId={stateOfApplicabilityId}
                                  connectionId={connectionId}
                                />
                              ))}
                            </div>
                          ),
                        )}
                    </div>
                  );
                })
              )}
        </div>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")}></DialogFooter>
    </>
  );
}

type AddApplicabilityStatementDialogProps = {
  onClose?: () => void;
};

export const AddApplicabilityStatementDialog = forwardRef<AddApplicabilityStatementDialogRef, AddApplicabilityStatementDialogProps>(
  ({ onClose }, ref) => {
    const { __ } = useTranslate();
    const dialogRef = useDialogRef();
    const [stateOfApplicabilityId, setStateOfApplicabilityId] = useState<string | null>(null);
    const [organizationId, setOrganizationId] = useState<string | null>(null);
    const [connectionId, setConnectionId] = useState<string | null>(null);

    useImperativeHandle(
      ref,
      () => ({
        open: (soaId: string, orgId: string, connId: string) => {
          setStateOfApplicabilityId(soaId);
          setOrganizationId(orgId);
          setConnectionId(connId);
          dialogRef.current?.open();
        },
      }),
      [dialogRef],
    );

    const handleClose = () => {
      setStateOfApplicabilityId(null);
      setOrganizationId(null);
      setConnectionId(null);
      onClose?.();
    };

    return (
      <Dialog
        ref={dialogRef}
        className="max-w-3xl"
        title={(
          <Breadcrumb
            items={[__("States of Applicability"), __("Add Statement")]}
          />
        )}
        onClose={handleClose}
      >
        {stateOfApplicabilityId && organizationId && connectionId
          ? (
              <Suspense
                fallback={(
                  <DialogContent
                    padded
                    className="flex items-center justify-center py-8"
                  >
                    <Spinner />
                  </DialogContent>
                )}
              >
                <AddApplicabilityStatementDialogContent
                  stateOfApplicabilityId={stateOfApplicabilityId}
                  organizationId={organizationId}
                  connectionId={connectionId}
                />
              </Suspense>
            )
          : null}
      </Dialog>
    );
  },
);

AddApplicabilityStatementDialog.displayName = "AddApplicabilityStatementDialog";
