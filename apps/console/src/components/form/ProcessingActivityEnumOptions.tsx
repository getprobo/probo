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

import { Option } from "@probo/ui";
import { useTranslation } from "react-i18next";

import type {
  ProcessingActivityDataProtectionImpactAssessment,
  ProcessingActivityLawfulBasis,
  ProcessingActivitySpecialOrCriminalDatum,
  ProcessingActivityTransferImpactAssessment,
} from "#/__generated__/core/ProcessingActivityGraphCreateMutation.graphql";

type Translator = (key: string) => string;

export function SpecialOrCriminalDataOptions() {
  const { t } = useTranslation();

  const options: Array<{
    value: ProcessingActivitySpecialOrCriminalDatum;
    label: string;
  }> = [
    {
      value: "YES",
      label: t("processingActivityEnumOptions.specialOrCriminalData.yes"),
    },
    {
      value: "NO",
      label: t("processingActivityEnumOptions.specialOrCriminalData.no"),
    },
    {
      value: "POSSIBLE",
      label: t("processingActivityEnumOptions.specialOrCriminalData.possible"),
    },
  ];

  return (
    <>
      {options.map(option => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function LawfulBasisOptions() {
  const { t } = useTranslation();

  const options: Array<{
    value: ProcessingActivityLawfulBasis;
    label: string;
  }> = [
    {
      value: "CONSENT",
      label: t("processingActivityEnumOptions.lawfulBasis.consent"),
    },
    {
      value: "CONTRACTUAL_NECESSITY",
      label: t(
        "processingActivityEnumOptions.lawfulBasis.contractualNecessity",
      ),
    },
    {
      value: "LEGAL_OBLIGATION",
      label: t("processingActivityEnumOptions.lawfulBasis.legalObligation"),
    },
    {
      value: "LEGITIMATE_INTEREST",
      label: t("processingActivityEnumOptions.lawfulBasis.legitimateInterest"),
    },
    {
      value: "PUBLIC_TASK",
      label: t("processingActivityEnumOptions.lawfulBasis.publicTask"),
    },
    {
      value: "VITAL_INTERESTS",
      label: t("processingActivityEnumOptions.lawfulBasis.vitalInterests"),
    },
  ];

  return (
    <>
      {options.map(option => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function getLawfulBasisLabel(
  value: ProcessingActivityLawfulBasis | null | undefined,
  t: Translator,
): string {
  if (!value) return "-";

  const labels = {
    CONSENT:
      t("processingActivityEnumOptions.lawfulBasis.consent"),
    CONTRACTUAL_NECESSITY:
      t("processingActivityEnumOptions.lawfulBasis.contractualNecessity"),
    LEGAL_OBLIGATION:
      t("processingActivityEnumOptions.lawfulBasis.legalObligation"),
    LEGITIMATE_INTEREST:
      t("processingActivityEnumOptions.lawfulBasis.legitimateInterest"),
    PUBLIC_TASK:
      t("processingActivityEnumOptions.lawfulBasis.publicTask"),
    VITAL_INTERESTS:
      t("processingActivityEnumOptions.lawfulBasis.vitalInterests"),
  };

  return labels[value] || value;
}

export function getResidualRiskLabel(
  value: "LOW" | "MEDIUM" | "HIGH" | null | undefined,
  t: Translator,
): string {
  if (!value) return "-";

  const labels = {
    LOW: t("processingActivityEnumOptions.residualRisk.low") || "Low",
    MEDIUM: t("processingActivityEnumOptions.residualRisk.medium") || "Medium",
    HIGH: t("processingActivityEnumOptions.residualRisk.high") || "High",
  };

  return labels[value] || value;
}

export function TransferSafeguardsOptions() {
  const { t } = useTranslation();

  const options: Array<{
    value: string;
    label: string;
  }> = [
    {
      value: "__NONE__",
      label: t("processingActivityEnumOptions.transferSafeguards.none"),
    },
    {
      value: "STANDARD_CONTRACTUAL_CLAUSES",
      label: t(
        "processingActivityEnumOptions.transferSafeguards.standardContractualClauses",
      ),
    },
    {
      value: "BINDING_CORPORATE_RULES",
      label: t(
        "processingActivityEnumOptions.transferSafeguards.bindingCorporateRules",
      ),
    },
    {
      value: "ADEQUACY_DECISION",
      label: t(
        "processingActivityEnumOptions.transferSafeguards.adequacyDecision",
      ),
    },
    {
      value: "DEROGATIONS",
      label: t("processingActivityEnumOptions.transferSafeguards.derogations"),
    },
    {
      value: "CODES_OF_CONDUCT",
      label: t(
        "processingActivityEnumOptions.transferSafeguards.codesOfConduct",
      ),
    },
    {
      value: "CERTIFICATION_MECHANISMS",
      label: t(
        "processingActivityEnumOptions.transferSafeguards.certificationMechanisms",
      ),
    },
  ];

  return (
    <>
      {options.map(option => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function DataProtectionImpactAssessmentOptions() {
  const { t } = useTranslation();

  const options: Array<{
    value: ProcessingActivityDataProtectionImpactAssessment;
    label: string;
  }> = [
    {
      value: "NEEDED",
      label: t(
        "processingActivityEnumOptions.dataProtectionImpactAssessment.needed",
      ),
    },
    {
      value: "NOT_NEEDED",
      label: t(
        "processingActivityEnumOptions.dataProtectionImpactAssessment.notNeeded",
      ),
    },
  ];

  return (
    <>
      {options.map(option => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function TransferImpactAssessmentOptions() {
  const { t } = useTranslation();

  const options: Array<{
    value: ProcessingActivityTransferImpactAssessment;
    label: string;
  }> = [
    {
      value: "NEEDED",
      label: t("processingActivityEnumOptions.transferImpactAssessment.needed"),
    },
    {
      value: "NOT_NEEDED",
      label: t(
        "processingActivityEnumOptions.transferImpactAssessment.notNeeded",
      ),
    },
  ];

  return (
    <>
      {options.map(option => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}

export function RoleOptions() {
  const { t } = useTranslation();

  const options: Array<{
    value: "CONTROLLER" | "PROCESSOR";
    label: string;
  }> = [
    {
      value: "CONTROLLER",
      label: t("processingActivityEnumOptions.roles.controller"),
    },
    {
      value: "PROCESSOR",
      label: t("processingActivityEnumOptions.roles.processor"),
    },
  ];

  return (
    <>
      {options.map(option => (
        <Option key={option.value} value={option.value}>
          {option.label}
        </Option>
      ))}
    </>
  );
}
