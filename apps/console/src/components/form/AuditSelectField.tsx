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

import { getAuditStateLabel, getAuditStateVariant } from "@probo/helpers";
import { Badge, Field, Option, Select } from "@probo/ui";
import { type ComponentProps, Suspense } from "react";
import {
  type Control,
  Controller,
  type FieldValues,
  type Path,
} from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { AuditSelectFieldQuery } from "#/__generated__/core/AuditSelectFieldQuery.graphql";

const auditsQuery = graphql`
  query AuditSelectFieldQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        audits(first: 100) {
          edges {
            node {
              id
              name
              framework {
                id
                name
              }
              state
            }
          }
        }
      }
    }
  }
`;

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: Path<T>;
  label?: string;
  error?: string;
} & ComponentProps<typeof Field>;

export function AuditSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  ...props
}: Props<T>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" loading placeholder="Loading..." />}
      >
        <AuditSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
        />
      </Suspense>
    </Field>
  );
}

function AuditSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "organizationId" | "control" | "name" | "disabled">,
) {
  const { t } = useTranslation();
  const { name, organizationId, control } = props;
  const data = useLazyLoadQuery<AuditSelectFieldQuery>(
    auditsQuery,
    { organizationId },
    { fetchPolicy: "network-only" },
  );
  const audits
    = data?.organization?.audits?.edges
      ?.map(edge => edge.node)
      .filter(node => node !== null) ?? [];

  const NONE_VALUE = "__NONE__";

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          disabled={props.disabled}
          id={name}
          variant="editor"
          placeholder={t("auditSelectField.placeholder")}
          onValueChange={value =>
            field.onChange(value === NONE_VALUE ? "" : value)}
          key={audits?.length.toString() ?? "0"}
          {...field}
          className="w-full"
          value={field.value || NONE_VALUE}
        >
          <Option value={NONE_VALUE}>
            <span className="text-txt-tertiary">{t("auditSelectField.none")}</span>
          </Option>
          {audits?.map(audit => (
            <Option key={audit.id} value={audit.id}>
              <div className="flex items-center justify-between w-full">
                <span>
                  {audit.name
                    ? `${audit.framework?.name} - ${audit.name}`
                    : audit.framework?.name}
                </span>
                <div className="ml-3">
                  <Badge variant={getAuditStateVariant(audit.state)}>
                    {getAuditStateLabel(t, audit.state)}
                  </Badge>
                </div>
              </div>
            </Option>
          ))}
        </Select>
      )}
    />
  );
}
