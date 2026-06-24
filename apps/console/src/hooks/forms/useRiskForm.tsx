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

import { z } from "zod";

import type { FormRiskDialog_risk$data } from "#/__generated__/core/FormRiskDialog_risk.graphql";

import { useFormWithSchema } from "../useFormWithSchema";

export type RiskNode = Pick<
  FormRiskDialog_risk$data,
  | "id"
  | "name"
  | "category"
  | "description"
  | "treatment"
  | "inherentLikelihood"
  | "inherentImpact"
  | "residualLikelihood"
  | "residualImpact"
  | "inherentRiskScore"
  | "residualRiskScore"
  | "note"
  | "owner"
>;

export const riskSchema = z.object({
  category: z.string().min(1, "Category is required"),
  name: z.string().min(1, "Name is required"),
  description: z.string().optional().nullable(),
  ownerId: z.string().min(1, "Owner is required"),
  treatment: z.enum(["AVOIDED", "MITIGATED", "TRANSFERRED", "ACCEPTED"]),
  inherentLikelihood: z.coerce.number().min(1).max(5),
  inherentImpact: z.coerce.number().min(1).max(5),
  residualLikelihood: z.coerce.number().min(1).max(5),
  residualImpact: z.coerce.number().min(1).max(5),
  note: z.string().optional(),
});

export const useRiskForm = (risk?: RiskNode) => {
  return useFormWithSchema(riskSchema, {
    defaultValues: risk
      ? {
          ...risk,
          description: risk.description ?? undefined,
          ownerId: risk.owner?.id,
        }
      : {
          inherentLikelihood: 3,
          inherentImpact: 3,
          residualLikelihood: 3,
          residualImpact: 3,
        },
  });
};

export type RiskForm = ReturnType<typeof useRiskForm>;

export type RiskData = z.infer<typeof riskSchema>;
