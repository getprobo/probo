import { useTranslate } from "@probo/i18n";

import { Option } from "../../Atoms/Select/Select";

export function ImpactOptions() {
  const { __ } = useTranslate();

  const descriptions = {
    LOW: {
      label: __("Low"),
      description: __("Minimal impact on business"),
    },
    MEDIUM: {
      label: __("Medium"),
      description: __("Moderate impact on business"),
    },
    HIGH: {
      label: __("High"),
      description: __("Significant business impact"),
    },
    CRITICAL: {
      label: __("Critical"),
      description: __("Critical to business operations"),
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
