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
  Button,
  Card,
  IconCheckmark1,
  IconCrossLargeX,
  IconPencil,
  Markdown,
  Textarea,
} from "@probo/ui";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ContextPage_UpdateMutation } from "#/__generated__/core/ContextPage_UpdateMutation.graphql";
import type { ContextPageFragment$key } from "#/__generated__/core/ContextPageFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const fragment = graphql`
  fragment ContextPageFragment on Organization {
    id
    canUpdateContext: permission(action: "core:organization-context:update")
    context {
      product
      architecture
      team
      processes
      customers
    }
  }
`;

const updateMutation = graphql`
  mutation ContextPage_UpdateMutation(
    $input: UpdateOrganizationContextInput!
  ) {
    updateOrganizationContext(input: $input) {
      context {
        organizationId
        product
        architecture
        team
        processes
        customers
      }
    }
  }
`;

type SectionKey = "product" | "architecture" | "team" | "processes" | "customers";

type SectionConfig = {
  key: SectionKey;
  title: string;
  description: string;
  placeholder: string;
};

type Props = {
  organization: ContextPageFragment$key;
};

export default function ContextPage(props: Props) {
  const { t } = useTranslation();
  const organization = useFragment(fragment, props.organization);
  const context = organization.context;

  const sections: SectionConfig[] = [
    {
      key: "product",
      title: t("context.sections.product.title"),
      description: t("context.sections.product.description"),
      placeholder: t("context.sections.product.placeholder"),
    },
    {
      key: "architecture",
      title: t("context.sections.architecture.title"),
      description: t("context.sections.architecture.description"),
      placeholder: t("context.sections.architecture.placeholder"),
    },
    {
      key: "team",
      title: t("context.sections.team.title"),
      description: t("context.sections.team.description"),
      placeholder: t("context.sections.team.placeholder"),
    },
    {
      key: "processes",
      title: t("context.sections.processes.title"),
      description: t("context.sections.processes.description"),
      placeholder: t("context.sections.processes.placeholder"),
    },
    {
      key: "customers",
      title: t("context.sections.customers.title"),
      description: t("context.sections.customers.description"),
      placeholder: t("context.sections.customers.placeholder"),
    },
  ];

  const values: Record<SectionKey, string | null> = {
    product: context?.product ?? null,
    architecture: context?.architecture ?? null,
    team: context?.team ?? null,
    processes: context?.processes ?? null,
    customers: context?.customers ?? null,
  };

  return (
    <div className="space-y-6">
      {sections.map(section => (
        <ContextSection
          key={section.key}
          section={section}
          organizationId={organization.id}
          value={values[section.key]}
          canEdit={organization.canUpdateContext}
        />
      ))}
    </div>
  );
}

function ContextSection({
  section,
  organizationId,
  value,
  canEdit,
}: {
  section: SectionConfig;
  organizationId: string;
  value: string | null;
  canEdit: boolean;
}) {
  const { t } = useTranslation();
  const [isEditing, setIsEditing] = useState(false);
  const [text, setText] = useState(value ?? "");
  const [displayedValue, setDisplayedValue] = useState(value ?? "");
  const justSavedRef = useRef(false);

  const [updateContext, isUpdating]
    = useMutationWithToasts<ContextPage_UpdateMutation>(
      updateMutation,
      {
        successMessage: t("context.messages.updated"),
        errorMessage: t("context.errors.update"),
      },
    );

  const handleSave = async () => {
    const valueToSave = text.trim();
    const previousValue = value ?? "";
    setDisplayedValue(valueToSave);
    justSavedRef.current = true;

    const valueToSend = valueToSave.length > 0 ? valueToSave : null;

    await updateContext({
      variables: {
        input: {
          organizationId,
          [section.key]: valueToSend,
        },
      },
      onError: () => {
        setDisplayedValue(previousValue);
        justSavedRef.current = false;
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          setDisplayedValue(previousValue);
          justSavedRef.current = false;
        }

        setIsEditing(false);
      },
    });
  };

  const handleCancel = () => {
    setText(value ?? "");
    setIsEditing(false);
  };

  return (
    <Card padded>
      {isEditing
        ? (
            <div className="space-y-4">
              <div>
                <h3 className="text-sm font-semibold">{section.title}</h3>
                <p className="text-xs text-txt-tertiary mt-1">
                  {section.description}
                </p>
              </div>
              <Textarea
                value={text}
                onChange={e => setText(e.target.value)}
                autogrow
                className="min-h-32 font-mono text-sm"
                placeholder={section.placeholder}
              />
              <div className="flex gap-2 justify-end">
                <Button
                  variant="secondary"
                  icon={IconCrossLargeX}
                  onClick={handleCancel}
                  disabled={isUpdating}
                >
                  {t("context.actions.cancel")}
                </Button>
                <Button
                  icon={IconCheckmark1}
                  onClick={() => void handleSave()}
                  disabled={isUpdating}
                >
                  {t("context.actions.save")}
                </Button>
              </div>
            </div>
          )
        : (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-semibold">{section.title}</h3>
                  <p className="text-xs text-txt-tertiary mt-1">
                    {section.description}
                  </p>
                </div>
                {canEdit && (
                  <Button
                    variant="quaternary"
                    icon={IconPencil}
                    onClick={() => {
                      setText(value ?? "");
                      setIsEditing(true);
                    }}
                  >
                    {t("context.actions.edit")}
                  </Button>
                )}
              </div>
              <div className="w-full">
                {displayedValue
                  ? (
                      <div className="prose prose-sm max-w-none w-full [&_.prose]:max-w-none">
                        <Markdown content={displayedValue} />
                      </div>
                    )
                  : (
                      <div className="text-txt-tertiary text-sm italic">
                        {t("context.empty")}
                      </div>
                    )}
              </div>
            </div>
          )}
    </Card>
  );
}
