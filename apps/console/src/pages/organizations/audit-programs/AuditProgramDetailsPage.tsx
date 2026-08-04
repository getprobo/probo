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
  formatDatetime,
  formatError,
  getAuditStateVariant,
  type GraphQLError,
} from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  Field,
  FrameworkLogo,
  IconTrashCan,
  Input,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useFragment,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";
import { z } from "zod";

import type { AuditProgramDetailsPageAuditsFragment$key } from "#/__generated__/core/AuditProgramDetailsPageAuditsFragment.graphql";
import type { AuditProgramDetailsPageLinkedAuditRowFragment$key } from "#/__generated__/core/AuditProgramDetailsPageLinkedAuditRowFragment.graphql";
import type { AuditProgramGraphNodeQuery } from "#/__generated__/core/AuditProgramGraphNodeQuery.graphql";
import { SortableTable } from "#/components/SortableTable";
import {
  auditProgramNodeQuery,
  useDeleteAuditProgram,
  useUpdateAuditProgram,
} from "#/hooks/graph/AuditProgramGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const updateAuditProgramSchema = z.object({
  name: z.string().trim().min(1),
  validFrom: z.string().optional(),
  validUntil: z.string().optional(),
});

const auditsFragment = graphql`
  fragment AuditProgramDetailsPageAuditsFragment on AuditProgram
  @refetchable(queryName: "AuditProgramDetailsPageAuditsQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    order: { type: "AuditOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    audits(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "AuditProgramDetailsPage_audits") {
      __id
      edges {
        node {
          id
          ...AuditProgramDetailsPageLinkedAuditRowFragment
        }
      }
    }
  }
`;

const linkedAuditRowFragment = graphql`
  fragment AuditProgramDetailsPageLinkedAuditRowFragment on Audit {
    id
    name
    state
    framework {
      id
      name
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<AuditProgramGraphNodeQuery>;
};

export default function AuditProgramDetailsPage(props: Props) {
  const auditProgram = usePreloadedQuery<AuditProgramGraphNodeQuery>(
    auditProgramNodeQuery,
    props.queryRef,
  );
  const program = auditProgram.node;
  const programId = program?.id ?? "";
  const programName = program?.name ?? "";
  const { i18n, t } = useTranslation("organizations/audit-programs");
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { toast } = useToast();

  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    auditsFragment,
    program as AuditProgramDetailsPageAuditsFragment$key | null,
  );
  const linkedAudits
    = pagination.data?.audits?.edges?.map(edge => edge.node) ?? [];

  const deleteAuditProgram = useDeleteAuditProgram(
    { id: programId, name: programName },
    ConnectionHandler.getConnectionID(
      organizationId,
      "AuditProgramsPage_auditPrograms",
    ),
    () => void navigate(`/organizations/${organizationId}/audits/programs`),
  );

  const { formState, handleSubmit, register, reset } = useFormWithSchema(
    updateAuditProgramSchema,
    {
      defaultValues: {
        name: programName,
        validFrom: program?.validFrom?.split("T")[0] || "",
        validUntil: program?.validUntil?.split("T")[0] || "",
      },
    },
  );

  const updateAuditProgram = useUpdateAuditProgram();

  const onSubmit = handleSubmit(async (formData) => {
    if (!programId) return;

    try {
      await updateAuditProgram({
        id: programId,
        name: formData.name,
        validFrom: formatDatetime(formData.validFrom) ?? null,
        validUntil: formatDatetime(formData.validUntil) ?? null,
      });
      reset(formData);
      toast({
        title: t("auditProgramDetailsPage.messages.success"),
        description: t("auditProgramDetailsPage.messages.updated"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("auditProgramDetailsPage.messages.error"),
        description: formatError(
          t("auditProgramDetailsPage.errors.update"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  if (!programId) {
    return null;
  }

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("auditProgramDetailsPage.breadcrumb.auditPrograms"),
            to: `/organizations/${organizationId}/audits/programs`,
          },
          { label: programName },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3">
            <FrameworkLogo
              name={program?.framework?.name || ""}
              lightLogoURL={program?.framework?.lightLogo?.downloadUrl}
              darkLogoURL={program?.framework?.darkLogo?.downloadUrl}
            />
            <div className="text-2xl">{program?.framework?.name}</div>
          </div>
        </div>
        <ActionDropdown variant="secondary">
          {program.canDelete && (
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteAuditProgram}
            >
              {t("auditProgramDetailsPage.actions.delete")}
            </DropdownItem>
          )}
        </ActionDropdown>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <form onSubmit={e => void onSubmit(e)} className="space-y-6">
          <Field label={t("auditProgramDetailsPage.fields.name")}>
            <Input
              {...register("name")}
              placeholder={t("auditProgramDetailsPage.fields.namePlaceholder")}
              disabled={!program.canUpdate}
            />
          </Field>

          <Field label={t("auditProgramDetailsPage.fields.validFrom")}>
            <Input
              {...register("validFrom")}
              type="date"
              disabled={!program.canUpdate}
            />
          </Field>

          <Field label={t("auditProgramDetailsPage.fields.validUntil")}>
            <Input
              {...register("validUntil")}
              type="date"
              disabled={!program.canUpdate}
            />
          </Field>

          <div className="flex justify-end">
            {formState.isDirty && program.canUpdate && (
              <Button type="submit" disabled={formState.isSubmitting}>
                {formState.isSubmitting
                  ? t("auditProgramDetailsPage.actions.updating")
                  : t("auditProgramDetailsPage.actions.update")}
              </Button>
            )}
          </div>
        </form>

        <Card padded className="space-y-4">
          <h3 className="text-lg font-medium">
            {t("auditProgramDetailsPage.audits.title")}
          </h3>
          {linkedAudits.length === 0
            ? (
                <p className="text-neutral-600">
                  {t("auditProgramDetailsPage.audits.empty")}
                </p>
              )
            : (
                <SortableTable {...pagination} pageSize={10}>
                  <Thead>
                    <Tr>
                      <Th>{t("auditProgramDetailsPage.audits.columns.name")}</Th>
                      <Th>{t("auditProgramDetailsPage.audits.columns.state")}</Th>
                    </Tr>
                  </Thead>
                  <Tbody>
                    {linkedAudits.map(audit => (
                      <LinkedAuditRow
                        key={audit.id}
                        auditKey={audit}
                        organizationId={organizationId}
                      />
                    ))}
                  </Tbody>
                </SortableTable>
              )}
          <p className="text-txt-tertiary text-xs">
            {t("auditProgramDetailsPage.meta.updated", {
              date: dateFormat(i18n.language, program.updatedAt),
            })}
          </p>
        </Card>
      </div>
    </div>
  );
}

function LinkedAuditRow({
  auditKey,
  organizationId,
}: {
  auditKey: AuditProgramDetailsPageLinkedAuditRowFragment$key;
  organizationId: string;
}) {
  const audit = useFragment(linkedAuditRowFragment, auditKey);
  const { t } = useTranslation("organizations/audit-programs");

  return (
    <Tr>
      <Td>
        <Link
          className="text-primary hover:underline"
          to={`/organizations/${organizationId}/audits/${audit.id}`}
        >
          {audit.name || audit.framework?.name}
        </Link>
      </Td>
      <Td>
        <Badge
          variant={getAuditStateVariant(audit.state || "NOT_STARTED")}
        >
          {t(
            `auditProgramDetailsPage.auditStates.${(audit.state || "NOT_STARTED").toLowerCase()}`,
          )}
        </Badge>
      </Td>
    </Tr>
  );
}
