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

import { formatDate, formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useList } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Breadcrumb,
  Button,
  Card,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconChevronDown,
  IconChevronRight,
  IconPlusLarge,
  IconRobot,
  Option,
  Select,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import * as Popover from "@radix-ui/react-popover";
import { useEffect, useMemo, useRef, useState } from "react";
import { type PreloadedQuery, useMutation, usePreloadedQuery, useRelayEnvironment } from "react-relay";
import { fetchQuery, graphql } from "relay-runtime";

import type { AccessEntryDecision, CampaignDetailPageBulkDecisionMutation } from "#/__generated__/core/CampaignDetailPageBulkDecisionMutation.graphql";
import type { AccessEntryFlag, CampaignDetailPageBulkFlagMutation } from "#/__generated__/core/CampaignDetailPageBulkFlagMutation.graphql";
import type { CampaignDetailPageCloseMutation } from "#/__generated__/core/CampaignDetailPageCloseMutation.graphql";
import type { CampaignDetailPageQuery } from "#/__generated__/core/CampaignDetailPageQuery.graphql";
import type { CampaignDetailPageStartMutation } from "#/__generated__/core/CampaignDetailPageStartMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  decisionBadgeVariant,
  decisionLabel,
  flagBadgeVariant,
  flagGroups,
  flagLabel,
  formatStatus,
  NotAvailable,
  statusBadgeVariant,
  statusLabel,
} from "../_components/accessReviewHelpers";
import { EntryDecisionActions } from "../_components/EntryDecisionActions";
import { EntryFlagSelect } from "../_components/EntryFlagSelect";
import { AddCampaignScopeSourceDialog } from "../dialogs/AddCampaignScopeSourceDialog";

const startCampaignMutation = graphql`
  mutation CampaignDetailPageStartMutation(
    $input: StartAccessReviewCampaignInput!
  ) {
    startAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
        startedAt
      }
    }
  }
`;

const closeCampaignMutation = graphql`
  mutation CampaignDetailPageCloseMutation(
    $input: CloseAccessReviewCampaignInput!
  ) {
    closeAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
        completedAt
      }
    }
  }
`;

const bulkDecisionMutation = graphql`
  mutation CampaignDetailPageBulkDecisionMutation(
    $input: RecordAccessEntryDecisionsInput!
  ) {
    recordAccessEntryDecisions(input: $input) {
      accessEntries {
        id
        decision
        decisionNote
      }
    }
  }
`;

const bulkFlagMutation = graphql`
  mutation CampaignDetailPageBulkFlagMutation(
    $input: FlagAccessEntryInput!
  ) {
    flagAccessEntry(input: $input) {
      accessEntry {
        id
        flags
        flagReasons
      }
    }
  }
`;

export const campaignDetailPageQuery = graphql`
  query CampaignDetailPageQuery($campaignId: ID!) {
    node(id: $campaignId) {
      __typename
      ... on AccessReviewCampaign {
        id
        name
        status
        createdAt
        startedAt
        completedAt
        pendingEntryCount
        scopeSources {
          id
          source {
            id
          }
          name
          fetchStatus
          fetchedAccountsCount
          entries(first: 500) {
            edges {
              node {
                id
                email
                fullName
                role
                isAdmin
                mfaStatus
                accountType
                lastLogin
                decision
                flags
              }
            }
            pageInfo {
              hasNextPage
            }
          }
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<CampaignDetailPageQuery>;
};

export default function CampaignDetailPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const environment = useRelayEnvironment();
  const data = usePreloadedQuery(campaignDetailPageQuery, queryRef);

  if (data.node.__typename !== "AccessReviewCampaign") {
    throw new Error("Campaign not found");
  }

  const campaign = data.node;
  const { toast } = useToast();
  const isInProgress = campaign.status === "IN_PROGRESS";
  const isDraft = campaign.status === "DRAFT";
  const isPendingActions = campaign.status === "PENDING_ACTIONS";

  const campaignIdRef = useRef(campaign.id);

  useEffect(() => {
    campaignIdRef.current = campaign.id;
  }, [campaign.id]);

  useEffect(() => {
    if (!isInProgress) return;
    const interval = setInterval(() => {
      if (document.hidden) return;
      fetchQuery<CampaignDetailPageQuery>(
        environment,
        campaignDetailPageQuery,
        { campaignId: campaignIdRef.current },
        { fetchPolicy: "network-only" },
      ).subscribe({});
    }, 3000);
    return () => clearInterval(interval);
  }, [isInProgress, environment]);
  const existingScopeSourceIds = useMemo(
    () => campaign.scopeSources.flatMap(s => s.source?.id ? [s.source.id] : []),
    [campaign.scopeSources],
  );

  const confirm = useConfirm();

  const [startCampaign, isStarting]
    = useMutation<CampaignDetailPageStartMutation>(startCampaignMutation);

  const [closeCampaign, isClosing]
    = useMutation<CampaignDetailPageCloseMutation>(closeCampaignMutation);

  const allDecided = campaign.scopeSources.length > 0
    && campaign.scopeSources.every(source =>
      source.entries
      && source.entries.edges.length > 0
      && source.entries.edges.every(edge => edge.node.decision !== "PENDING")
      && !source.entries.pageInfo.hasNextPage,
    );

  const handleStart = () => {
    startCampaign({
      variables: {
        input: {
          accessReviewCampaignId: campaign.id,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to start campaign"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Campaign started. Sources are being fetched."),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to start campaign"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleComplete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          closeCampaign({
            variables: {
              input: { accessReviewCampaignId: campaign.id },
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: formatError(
                    __("Failed to complete campaign"),
                    errors as GraphQLError[],
                  ),
                  variant: "error",
                });
                resolve();
                return;
              }
              toast({
                title: __("Success"),
                description: __("Campaign completed successfully."),
                variant: "success",
              });
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to complete campaign"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: __(
          "Are you sure you want to complete this campaign? This action cannot be undone. All decisions will be finalized.",
        ),
        label: __("Complete"),
        variant: "primary",
      },
    );
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Access Reviews"),
            to: `/organizations/${organizationId}/access-reviews`,
          },
          { label: campaign.name },
        ]}
      />

      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-semibold">{campaign.name}</h1>
        <Badge variant={statusBadgeVariant(campaign.status)}>
          {statusLabel(__, campaign.status)}
        </Badge>
        {isPendingActions && (
          <Button
            onClick={handleComplete}
            disabled={!allDecided || isClosing}
          >
            {isClosing ? __("Completing...") : __("Complete campaign")}
          </Button>
        )}
      </div>

      <div className="space-y-4">
        {isDraft && (
          <div className="flex items-center justify-end gap-2">
            <AddCampaignScopeSourceDialog
              organizationId={organizationId}
              campaignId={campaign.id}
              existingScopeSourceIds={existingScopeSourceIds}
            >
              <Button icon={IconPlusLarge} variant="secondary">
                {__("Add source")}
              </Button>
            </AddCampaignScopeSourceDialog>
            {campaign.scopeSources.length > 0 && (
              <Button
                onClick={handleStart}
                disabled={isStarting}
              >
                {isStarting ? __("Starting...") : __("Start campaign")}
              </Button>
            )}
          </div>
        )}

        {campaign.scopeSources.map(source => (
          <ScopeSourceCard
            key={source.id}
            source={source}
            isPendingActions={isPendingActions}
          />
        ))}

        {campaign.scopeSources.length === 0 && (
          <Card padded>
            <div className="text-center py-8">
              <p className="text-txt-tertiary">
                {__("No sources configured for this campaign.")}
              </p>
            </div>
          </Card>
        )}
      </div>
    </div>
  );
}

type ScopeSource = NonNullable<
  Extract<
    CampaignDetailPageQuery["response"]["node"],
    { readonly __typename: "AccessReviewCampaign" }
  >["scopeSources"]
>[number];

function ScopeSourceCard({ source, isPendingActions }: { source: ScopeSource; isPendingActions: boolean }) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [expanded, setExpanded] = useState(false);
  const { list: selection, toggle, clear, reset } = useList<string>([]);
  const [bulkPendingDecision, setBulkPendingDecision] = useState<AccessEntryDecision | null>(null);
  const [bulkNote, setBulkNote] = useState("");
  const bulkNoteRef = useDialogRef();

  const [bulkDecide]
    = useMutation<CampaignDetailPageBulkDecisionMutation>(bulkDecisionMutation);
  const [bulkFlag]
    = useMutation<CampaignDetailPageBulkFlagMutation>(bulkFlagMutation);

  const entries = source.entries?.edges ?? [];
  const entryIds = entries.map(edge => edge.node.id);

  const handleBulkDecision = (value: string) => {
    const decision = value as AccessEntryDecision;
    if (decision === "APPROVED") {
      bulkDecide({
        variables: {
          input: {
            decisions: selection.map(id => ({
              accessEntryId: id,
              decision: "APPROVED" as AccessEntryDecision,
            })),
          },
        },
        onCompleted(_, errors) {
          if (errors?.length) {
            toast({
              title: __("Error"),
              description: formatError(
                __("Failed to record decisions"),
                errors as GraphQLError[],
              ),
              variant: "error",
            });
            return;
          }
          toast({
            title: __("Success"),
            description: __("Decisions recorded successfully."),
            variant: "success",
          });
          clear();
        },
        onError(error) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to record decisions"),
              error as GraphQLError,
            ),
            variant: "error",
          });
        },
      });
    } else {
      setBulkPendingDecision(decision);
      setBulkNote("");
      bulkNoteRef.current?.open();
    }
  };

  const [bulkFlagSelection, setBulkFlagSelection] = useState<AccessEntryFlag[]>([]);
  const [bulkFlagOpen, setBulkFlagOpen] = useState(false);
  const bulkFlagOpenedWithRef = useRef<AccessEntryFlag[]>([]);

  const toggleBulkFlag = (flagValue: AccessEntryFlag) => {
    setBulkFlagSelection(prev =>
      prev.includes(flagValue)
        ? prev.filter(f => f !== flagValue)
        : [...prev, flagValue],
    );
  };

  const handleBulkFlagOpenChange = (nextOpen: boolean) => {
    if (nextOpen) {
      bulkFlagOpenedWithRef.current = [];
      setBulkFlagSelection([]);
    }

    if (!nextOpen && bulkFlagSelection.length > 0) {
      let errorCount = 0;
      let completedCount = 0;
      const total = selection.length;

      for (const entryId of selection) {
        bulkFlag({
          variables: {
            input: {
              accessEntryId: entryId,
              flags: bulkFlagSelection,
            },
          },
          onCompleted(_, errors) {
            if (errors?.length) {
              errorCount++;
            }
            completedCount++;
            if (completedCount === total) {
              if (errorCount > 0) {
                toast({
                  title: __("Error"),
                  description: sprintf(__("Failed to update flags for %d entries."), errorCount),
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Flags updated for selected entries."),
                  variant: "success",
                });
              }
              clear();
            }
          },
          onError() {
            errorCount++;
            completedCount++;
            if (completedCount === total) {
              toast({
                title: __("Error"),
                description: sprintf(__("Failed to update flags for %d entries."), errorCount),
                variant: "error",
              });
              clear();
            }
          },
        });
      }
    }

    setBulkFlagOpen(nextOpen);
  };

  return (
    <Card>
      <button
        type="button"
        className="flex w-full items-center justify-between p-4 text-left hover:bg-bg-subtle transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-3">
          {expanded
            ? <IconChevronDown className="size-4 text-txt-tertiary" />
            : <IconChevronRight className="size-4 text-txt-tertiary" />}
          <span className="font-medium">{source.name}</span>
          <Badge variant="neutral">
            {source.fetchedAccountsCount}
            {" "}
            {__("accounts")}
          </Badge>
          <Badge variant={source.fetchStatus === "SUCCESS" ? "success" : "info"}>
            {formatStatus(source.fetchStatus)}
          </Badge>
        </div>
      </button>

      {expanded && (
        <div className="border-t">
          {entries.length === 0
            ? (
                <div className="px-4 py-6 text-center text-txt-tertiary">
                  {__("No entries found for this source.")}
                </div>
              )
            : (
                <div className="relative w-full overflow-auto">
                  <table className="w-full text-left">
                    <Thead>
                      <Tr>
                        {isPendingActions && (
                          <Th className="w-12">
                            <Checkbox
                              checked={selection.length === entryIds.length && entryIds.length > 0}
                              onChange={() => selection.length === entryIds.length ? clear() : reset(entryIds)}
                            />
                          </Th>
                        )}
                        <Th>{__("Name")}</Th>
                        <Th>{__("Email")}</Th>
                        <Th>{__("Role")}</Th>
                        <Th>{__("Admin")}</Th>
                        <Th>{__("MFA")}</Th>
                        <Th>{__("Last login")}</Th>
                        <Th>{__("Flag")}</Th>
                        <Th>{__("Decision")}</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {entries.map(edge => (
                        <Tr key={edge.node.id}>
                          {isPendingActions && (
                            <Td noLink>
                              <Checkbox
                                checked={selection.includes(edge.node.id)}
                                onChange={() => toggle(edge.node.id)}
                              />
                            </Td>
                          )}
                          <Td>
                            <span className="flex items-center gap-1.5">
                              {edge.node.accountType === "SERVICE_ACCOUNT" && (
                                <IconRobot size={16} className="text-txt-tertiary shrink-0" />
                              )}
                              {edge.node.fullName || <NotAvailable />}
                            </span>
                          </Td>
                          <Td>{edge.node.email || <NotAvailable />}</Td>
                          <Td>{edge.node.role || <NotAvailable />}</Td>
                          <Td>{edge.node.isAdmin ? __("Yes") : __("No")}</Td>
                          <Td>
                            {edge.node.mfaStatus === "UNKNOWN"
                              ? <NotAvailable />
                              : (
                                  <Badge variant={edge.node.mfaStatus === "ENABLED" ? "success" : "neutral"}>
                                    {formatStatus(edge.node.mfaStatus)}
                                  </Badge>
                                )}
                          </Td>
                          <Td>
                            {edge.node.lastLogin
                              ? formatDate(edge.node.lastLogin)
                              : <NotAvailable />}
                          </Td>
                          <Td>
                            {isPendingActions
                              ? (
                                  <EntryFlagSelect
                                    entryId={edge.node.id}
                                    currentFlags={edge.node.flags}
                                  />
                                )
                              : edge.node.flags.length > 0 && (
                                <div className="flex flex-wrap gap-1">
                                  {edge.node.flags.map(f => (
                                    <Badge key={f} variant={flagBadgeVariant(f)}>
                                      {flagLabel(f)}
                                    </Badge>
                                  ))}
                                </div>
                              )}
                          </Td>
                          <Td>
                            {isPendingActions
                              ? (
                                  <EntryDecisionActions
                                    entryId={edge.node.id}
                                    decision={edge.node.decision}
                                  />
                                )
                              : edge.node.decision !== "PENDING" && (
                                <Badge variant={decisionBadgeVariant(edge.node.decision)}>
                                  {decisionLabel(__, edge.node.decision)}
                                </Badge>
                              )}
                          </Td>
                        </Tr>
                      ))}
                    </Tbody>
                  </table>
                </div>
              )}

          {selection.length > 0 && (
            <div className="flex items-center gap-4 p-4 border-t">
              <span className="text-sm text-txt-secondary">
                {selection.length}
                {" "}
                {__("selected")}
              </span>
              <Button variant="secondary" onClick={clear}>
                {__("Clear")}
              </Button>
              <Select
                variant="editor"
                placeholder={__("Set decision...")}
                onValueChange={handleBulkDecision}
              >
                <Option value="APPROVED">{__("Approve")}</Option>
                <Option value="REVOKE">{__("Revoke")}</Option>
                <Option value="DEFER">{__("Modify")}</Option>
                <Option value="ESCALATE">{__("Escalate")}</Option>
              </Select>
              <Popover.Root open={bulkFlagOpen} onOpenChange={handleBulkFlagOpenChange}>
                <Popover.Trigger asChild>
                  <Button variant="secondary">
                    {bulkFlagSelection.length > 0
                      ? `${bulkFlagSelection.length} ${__("flags")}`
                      : __("Set flags...")}
                  </Button>
                </Popover.Trigger>
                <Popover.Portal>
                  <Popover.Content
                    sideOffset={5}
                    className="z-100 w-64 rounded-[10px] bg-level-1 p-2 shadow-mid animate-in fade-in slide-in-from-top-2"
                  >
                    {flagGroups.map(group => (
                      <div key={group.label} className="mb-2 last:mb-0">
                        <div className="px-2 py-1 text-xs font-semibold text-txt-tertiary uppercase tracking-wider">
                          {__(group.label)}
                        </div>
                        {group.flags.map(flag => (
                          <label
                            key={flag.value}
                            className="flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer hover:bg-tertiary-hover"
                          >
                            <Checkbox
                              checked={bulkFlagSelection.includes(flag.value)}
                              onChange={() => toggleBulkFlag(flag.value)}
                            />
                            <span className="text-sm text-txt-primary">{__(flag.label)}</span>
                          </label>
                        ))}
                      </div>
                    ))}
                  </Popover.Content>
                </Popover.Portal>
              </Popover.Root>
            </div>
          )}

          <Dialog ref={bulkNoteRef} title={__("Decision note")}>
            <DialogContent padded className="space-y-4">
              <p className="text-sm text-txt-secondary">
                {__("Please provide a reason for this decision.")}
              </p>
              <Field
                label={__("Note")}
                type="textarea"
                value={bulkNote}
                onValueChange={setBulkNote}
              />
            </DialogContent>
            <DialogFooter>
              <Button
                disabled={!bulkNote.trim()}
                onClick={() => {
                  if (bulkPendingDecision) {
                    bulkDecide({
                      variables: {
                        input: {
                          decisions: selection.map(id => ({
                            accessEntryId: id,
                            decision: bulkPendingDecision,
                            decisionNote: bulkNote,
                          })),
                        },
                      },
                      onCompleted(_, errors) {
                        if (errors?.length) {
                          toast({
                            title: __("Error"),
                            description: formatError(
                              __("Failed to record decisions"),
                              errors as GraphQLError[],
                            ),
                            variant: "error",
                          });
                          return;
                        }
                        toast({
                          title: __("Success"),
                          description: __("Decisions recorded successfully."),
                          variant: "success",
                        });
                        clear();
                        setBulkPendingDecision(null);
                        setBulkNote("");
                        bulkNoteRef.current?.close();
                      },
                      onError(error) {
                        toast({
                          title: __("Error"),
                          description: formatError(
                            __("Failed to record decisions"),
                            error as GraphQLError,
                          ),
                          variant: "error",
                        });
                      },
                    });
                  }
                }}
              >
                {__("Confirm")}
              </Button>
            </DialogFooter>
          </Dialog>

          {source.entries?.pageInfo.hasNextPage && (
            <div className="p-4 border-t text-center">
              <p className="text-sm text-txt-tertiary">
                {sprintf(__("Showing first %d entries. Use the CLI for the full list."), entries.length)}
              </p>
            </div>
          )}
        </div>
      )}
    </Card>
  );
}
