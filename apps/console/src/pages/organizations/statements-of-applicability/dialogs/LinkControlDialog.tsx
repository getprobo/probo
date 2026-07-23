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
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { LinkControlDialogQuery } from "#/__generated__/core/LinkControlDialogQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const linkControlQuery = graphql`
  query LinkControlDialogQuery($statementOfApplicabilityId: ID!, $organizationId: ID!) {
    statementOfApplicability: node(id: $statementOfApplicabilityId) {
      ... on StatementOfApplicability {
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

const linkControlMutation = graphql`
  mutation LinkControlDialogLinkMutation($input: CreateApplicabilityStatementInput!) {
    createApplicabilityStatement(input: $input) {
      applicabilityStatementEdge {
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
`;

const unlinkControlMutation = graphql`
  mutation LinkControlDialogUnlinkMutation($input: DeleteApplicabilityStatementInput!) {
    deleteApplicabilityStatement(input: $input) {
      deletedApplicabilityStatementId
    }
  }
`;

export type LinkControlDialogRef = {
  open: (statementOfApplicabilityId: string, organizationId: string, onUpdate?: () => void) => void;
};

type Control = {
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
  statementOfApplicabilityId,
  isLinked,
  onUpdate,
}: {
  control: Control;
  statementOfApplicabilityId: string;
  isLinked: boolean;
  onUpdate?: () => void;
}) {
  const { t } = useTranslation();
  const [selectedState, setSelectedState] = useState<string>(() => {
    if (!isLinked) return "not-linked";
    return control.applicability ? "applicable" : "not-applicable";
  });
  const [justification, setJustification] = useState(
    control.justification || "",
  );
  const [showJustification, setShowJustification] = useState(false);

  const [linkMutate, isLinking] = useMutationWithToasts(linkControlMutation, {
    successMessage: t("linkControlDialog.messages.updated"),
    errorMessage: t("linkControlDialog.errors.update"),
  });

  const [unlinkMutate, isUnlinking] = useMutationWithToasts(unlinkControlMutation, {
    successMessage: t("linkControlDialog.messages.removed"),
    errorMessage: t("linkControlDialog.errors.remove"),
  });

  const handleStateChange = async (newState: string) => {
    setSelectedState(newState);

    if (newState === "not-linked") {
      if (!control.applicabilityStatementId) return;
      await unlinkMutate({
        variables: {
          input: {
            applicabilityStatementId: control.applicabilityStatementId,
          },
        },
        onSuccess: () => {
          setShowJustification(false);
          onUpdate?.();
        },
      });
    } else if (newState === "applicable") {
      setShowJustification(false);
      await linkMutate({
        variables: {
          input: {
            statementOfApplicabilityId,
            controlId: control.controlId,
            applicability: true,
            justification: null,
          },
        },
        onSuccess: () => {
          onUpdate?.();
        },
      });
    } else if (newState === "not-applicable") {
      setShowJustification(true);
      setJustification(control.justification || "");
    }
  };

  const handleSaveJustification = async () => {
    await linkMutate({
      variables: {
        input: {
          statementOfApplicabilityId,
          controlId: control.controlId,
          applicability: false,
          justification: justification || null,
        },
      },
      onSuccess: () => {
        setShowJustification(false);
        onUpdate?.();
      },
    });
  };

  return (
    <div className="p-4 border-b border-border-low">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Badge size="md">{control.sectionTitle}</Badge>
            <span className="text-sm font-medium text-txt-primary">{control.name}</span>
          </div>
          {isLinked && control.applicability !== null && !showJustification && control.justification && (
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
            disabled={isLinking || isUnlinking}
            className="w-48"
          >
            <Option value="not-linked">
              {t("linkControlDialog.applicability.notLinked")}
            </Option>
            <Option value="applicable">
              {t("linkControlDialog.applicability.applicable")}
            </Option>
            <Option value="not-applicable">
              {t("linkControlDialog.applicability.notApplicable")}
            </Option>
          </Select>
        </div>
      </div>
      {showJustification && (
        <div className="mt-3 flex items-start gap-2">
          <Textarea
            value={justification}
            onChange={e => setJustification(e.target.value)}
            placeholder={t("linkControlDialog.placeholders.justification")}
            className="flex-1"
            autogrow
          />
          <Button
            variant="primary"
            icon={IconCheckmark1}
            onClick={() => void handleSaveJustification()}
            disabled={isLinking}
            aria-label={t("linkControlDialog.actions.save")}
          />
        </div>
      )}
    </div>
  );
}

function LinkControlDialogContent({
  statementOfApplicabilityId,
  organizationId,
  onUpdate,
}: {
  statementOfApplicabilityId: string;
  organizationId: string;
  onUpdate?: () => void;
}) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const [collapsedFrameworks, setCollapsedFrameworks] = useState<Set<string>>(new Set());
  const data = useLazyLoadQuery<LinkControlDialogQuery>(
    linkControlQuery,
    { statementOfApplicabilityId, organizationId },
    { fetchPolicy: "store-or-network" },
  ) as {
    statementOfApplicability: {
      id: string;
      applicabilityStatements?: {
        edges: Array<{
          node: {
            id: string;
            applicability: boolean;
            justification: string | null;
            control: { id: string };
          };
        }>;
      };
    } | null;
    organization: {
      id: string;
      controls?: {
        edges: Array<{
          node: {
            id: string;
            sectionTitle: string;
            name: string;
            framework: {
              id: string;
              name: string;
            };
          };
        }>;
      };
    } | null;
  };

  // Build a map of control ID -> applicability statement
  const applicabilityMap = useMemo(() => {
    const map = new Map<string, { id: string; applicability: boolean; justification: string | null }>();
    data.statementOfApplicability?.applicabilityStatements?.edges.forEach((edge) => {
      map.set(edge.node.control.id, {
        id: edge.node.id,
        applicability: edge.node.applicability,
        justification: edge.node.justification,
      });
    });
    return map;
  }, [data.statementOfApplicability?.applicabilityStatements]);

  // Merge controls with applicability info
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
      };
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
    const groups: Record<string, Record<string, Control[]>> = {};
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
            placeholder={t("linkControlDialog.placeholders.search")}
            onValueChange={setSearch}
          />
        </div>
        <div className="max-h-[60vh] overflow-y-auto">
          {filteredControls.length === 0
            ? (
                <div className="p-8 text-center text-txt-secondary">
                  {t("linkControlDialog.empty")}
                </div>
              )
            : (
                Object.entries(groupedControls).map(([frameworkName, sections]) => {
                  const isCollapsed = collapsedFrameworks.has(frameworkName);
                  return (
                    <div key={frameworkName}>
                      <div className="sticky top-0 bg-level-1 px-4 py-2 border-b border-border-low z-10 flex items-center justify-between">
                        <h3 className="text-sm font-semibold text-txt-primary">{frameworkName}</h3>
                        <Button
                          variant="tertiary"
                          icon={isCollapsed ? IconChevronDown : IconChevronUp}
                          onClick={() => toggleFramework(frameworkName)}
                          aria-label={isCollapsed ? t("linkControlDialog.actions.expand") : t("linkControlDialog.actions.collapse")}
                        />
                      </div>
                      {!isCollapsed && Object.entries(sections).map(([sectionTitle, sectionControls]) => (
                        <div key={`${frameworkName}-${sectionTitle}`}>
                          {sectionControls.map(control => (
                            <ControlRow
                              key={control.controlId}
                              control={control}
                              statementOfApplicabilityId={statementOfApplicabilityId}
                              isLinked={control.applicabilityStatementId !== null}
                              onUpdate={onUpdate}
                            />
                          ))}
                        </div>
                      ))}
                    </div>
                  );
                })
              )}
        </div>
      </DialogContent>
      <DialogFooter exitLabel={t("linkControlDialog.actions.close")}></DialogFooter>
    </>
  );
}

export const LinkControlDialog = forwardRef<LinkControlDialogRef>((_props, ref) => {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [statementOfApplicabilityId, setStatementOfApplicabilityId] = useState<string | null>(null);
  const [organizationId, setOrganizationId] = useState<string | null>(null);
  const [onUpdateCallback, setOnUpdateCallback] = useState<(() => void) | undefined>(undefined);

  useImperativeHandle(ref, () => ({
    open: (soaId: string, orgId: string, callback?: () => void) => {
      setStatementOfApplicabilityId(soaId);
      setOrganizationId(orgId);
      setOnUpdateCallback(() => callback);
      dialogRef.current?.open();
    },
  }), [dialogRef]);

  const handleClose = () => {
    setStatementOfApplicabilityId(null);
    setOrganizationId(null);
    setOnUpdateCallback(undefined);
  };

  return (
    <Dialog
      ref={dialogRef}
      className="max-w-3xl"
      title={(
        <Breadcrumb
          items={[t("linkControlDialog.breadcrumb.statementsOfApplicability"), t("linkControlDialog.breadcrumb.addControls")]}
        />
      )}
      onClose={handleClose}
    >
      {statementOfApplicabilityId && organizationId
        ? (
            <Suspense
              fallback={(
                <DialogContent padded className="flex items-center justify-center py-8">
                  <Spinner />
                </DialogContent>
              )}
            >
              <LinkControlDialogContent
                statementOfApplicabilityId={statementOfApplicabilityId}
                organizationId={organizationId}
                onUpdate={onUpdateCallback}
              />
            </Suspense>
          )
        : null}
    </Dialog>
  );
});

LinkControlDialog.displayName = "LinkControlDialog";
