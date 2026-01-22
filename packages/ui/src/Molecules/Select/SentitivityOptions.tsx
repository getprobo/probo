import { useTranslate } from "@probo/i18n";

import { Option } from "../../Atoms/Select/Select";

export function SentitivityOptions() {
  const { __ } = useTranslate();

  const descriptions = {
    NONE: {
      label: __("None"),
      description: __("No sensitive data"),
    },
    LOW: {
      label: __("Low"),
      description: __("Public or non-sensitive data"),
    },
    MEDIUM: {
      label: __("Medium"),
      description: __("Internal/restricted data"),
    },
    HIGH: {
      label: __("High"),
      description: __("Confidential data"),
    },
    CRITICAL: {
      label: __("Critical"),
      description: __("Regulated/PII/financial data"),
    },
  } as const;

  return (
    <>
      {Object.entries(descriptions).map(([key, description]) => (
        <Option
          key={key}
          value={key}
          className="border-b border-border-low"
        >
          <span>
            <span className="text-sm font-bold">
              {description.label}
            </span>
            ,
            {" "}
            <span className="text-sm text-txt-secondary">
              {description.description}
            </span>
          </span>
        </Option>
      ))}
    </>
  );
}
