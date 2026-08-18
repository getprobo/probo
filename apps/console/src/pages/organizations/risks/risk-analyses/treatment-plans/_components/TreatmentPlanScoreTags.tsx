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

import { SeverityBadge } from "@probo/ui";

import type { MatrixSize } from "../../_components/matrixSize";

interface TreatmentPlanScoreTagsProps {
  inherentLikelihood: number;
  inherentImpact: number;
  residualLikelihood: number;
  residualImpact: number;
  matrixSize: MatrixSize;
}

function scoreTag(likelihood: number, impact: number, matrixSize: MatrixSize) {
  return (
    <SeverityBadge
      score={likelihood * impact}
      label={`${likelihood} × ${impact}`}
      matrixSize={matrixSize}
    />
  );
}

export function TreatmentPlanScoreTags({
  inherentLikelihood,
  inherentImpact,
  residualLikelihood,
  residualImpact,
  matrixSize,
}: TreatmentPlanScoreTagsProps) {
  return (
    <div className="flex items-center gap-1.5">
      {scoreTag(inherentLikelihood, inherentImpact, matrixSize)}
      {scoreTag(residualLikelihood, residualImpact, matrixSize)}
    </div>
  );
}
