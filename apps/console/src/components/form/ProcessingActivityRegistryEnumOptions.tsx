import { useTranslate } from "@probo/i18n";
import { Option } from "@probo/ui";
import type {
  ProcessingActivityRegistrySpecialOrCriminalData,
  ProcessingActivityRegistryLawfulBasis,
  ProcessingActivityRegistryDataProtectionImpactAssessment,
  ProcessingActivityRegistryTransferImpactAssessment,
} from "../../hooks/graph/__generated__/ProcessingActivityRegistryGraphCreateMutation.graphql";

export function SpecialOrCriminalDataOptions() {
  const { __ } = useTranslate();

  const options: Array<{
    value: ProcessingActivityRegistrySpecialOrCriminalData;
    label: string;
  }> = [
    { value: "YES", label: __("Yes") },
    { value: "NO", label: __("No") },
    { value: "POSSIBLE", label: __("Possible") },
  ];

  return (
    <>
      {options.map((option) => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function LawfulBasisOptions() {
  const { __ } = useTranslate();

  const options: Array<{
    value: ProcessingActivityRegistryLawfulBasis;
    label: string;
  }> = [
    { value: "CONSENT", label: __("Consent") },
    { value: "CONTRACTUAL_NECESSITY", label: __("Contractual Necessity") },
    { value: "LEGAL_OBLIGATION", label: __("Legal Obligation") },
    { value: "LEGITIMATE_INTEREST", label: __("Legitimate Interest") },
    { value: "PUBLIC_TASK", label: __("Public Task") },
    { value: "VITAL_INTERESTS", label: __("Vital Interests") },
  ];

  return (
    <>
      {options.map((option) => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function getLawfulBasisLabel(value: ProcessingActivityRegistryLawfulBasis | null | undefined, __: (key: string) => string): string {
  if (!value) return "-";

  const labels = {
    "CONSENT": __("Consent"),
    "CONTRACTUAL_NECESSITY": __("Contractual Necessity"),
    "LEGAL_OBLIGATION": __("Legal Obligation"),
    "LEGITIMATE_INTEREST": __("Legitimate Interest"),
    "PUBLIC_TASK": __("Public Task"),
    "VITAL_INTERESTS": __("Vital Interests"),
  };

  return labels[value] || value;
}

export function TransferSafeguardsOptions() {
  const { __ } = useTranslate();

  const options: Array<{
    value: string;
    label: string;
  }> = [
    { value: "__NONE__", label: __("None") },
    { value: "STANDARD_CONTRACTUAL_CLAUSES", label: __("Standard Contractual Clauses") },
    { value: "BINDING_CORPORATE_RULES", label: __("Binding Corporate Rules") },
    { value: "ADEQUACY_DECISION", label: __("Adequacy Decision") },
    { value: "DEROGATIONS", label: __("Derogations") },
    { value: "CODES_OF_CONDUCT", label: __("Codes of Conduct") },
    { value: "CERTIFICATION_MECHANISMS", label: __("Certification Mechanisms") },
  ];

  return (
    <>
      {options.map((option) => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function DataProtectionImpactAssessmentOptions() {
  const { __ } = useTranslate();

  const options: Array<{
    value: ProcessingActivityRegistryDataProtectionImpactAssessment;
    label: string;
  }> = [
    { value: "NEEDED", label: __("Needed") },
    { value: "NOT_NEEDED", label: __("Not Needed") },
  ];

  return (
    <>
      {options.map((option) => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function TransferImpactAssessmentOptions() {
  const { __ } = useTranslate();

  const options: Array<{
    value: ProcessingActivityRegistryTransferImpactAssessment;
    label: string;
  }> = [
    { value: "NEEDED", label: __("Needed") },
    { value: "NOT_NEEDED", label: __("Not Needed") },
  ];

  return (
    <>
      {options.map((option) => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}
