import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
  const organization = useFragment(fragment, props.organization);
  const context = organization.context;

  const sections: SectionConfig[] = [
    {
      key: "product",
      title: __("Product"),
      description: __("Describe what your product does, its main features, and value proposition."),
      placeholder: __("Describe your product in markdown format..."),
    },
    {
      key: "architecture",
      title: __("Architecture"),
      description: __("Describe your technical architecture, infrastructure, and key design decisions."),
      placeholder: __("Describe your architecture in markdown format..."),
    },
    {
      key: "team",
      title: __("Team"),
      description: __("Describe your team structure, roles, and responsibilities."),
      placeholder: __("Describe your team in markdown format..."),
    },
    {
      key: "processes",
      title: __("Processes"),
      description: __("Describe your key processes, workflows, and operational procedures."),
      placeholder: __("Describe your processes in markdown format..."),
    },
    {
      key: "customers",
      title: __("Customers"),
      description: __("Describe your target market, customer segments, and use cases."),
      placeholder: __("Describe your customers in markdown format..."),
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
  const { __ } = useTranslate();
  const [isEditing, setIsEditing] = useState(false);
  const [text, setText] = useState(value ?? "");
  const [displayedValue, setDisplayedValue] = useState(value ?? "");
  const justSavedRef = useRef(false);

  const [updateContext, isUpdating]
    = useMutationWithToasts<ContextPage_UpdateMutation>(
      updateMutation,
      {
        successMessage: __("Context updated successfully"),
        errorMessage: __("Failed to update context"),
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
                  {__("Cancel")}
                </Button>
                <Button
                  icon={IconCheckmark1}
                  onClick={() => void handleSave()}
                  disabled={isUpdating}
                >
                  {__("Save")}
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
                    {__("Edit")}
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
                        {__("No content yet. Click Edit to add one.")}
                      </div>
                    )}
              </div>
            </div>
          )}
    </Card>
  );
}
