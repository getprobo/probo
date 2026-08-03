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

export const ONBOARDING_STEP_IDS = [
  "scim",
  "accessReview",
  "agent",
  "mcp",
  "congrats",
] as const;

export type OnboardingStepId = (typeof ONBOARDING_STEP_IDS)[number];

export const INTEGRATION_STEP_IDS = [
  "scim",
  "accessReview",
  "agent",
] as const satisfies readonly OnboardingStepId[];

export type IntegrationStepId = (typeof INTEGRATION_STEP_IDS)[number];

export function isOnboardingStepId(value: string | null): value is OnboardingStepId {
  return value !== null && (ONBOARDING_STEP_IDS as readonly string[]).includes(value);
}

export type OnboardingCompletion = {
  scim: boolean;
  accessReview: boolean;
  agent: boolean;
};

export function stepIsComplete(
  stepId: IntegrationStepId,
  completion: OnboardingCompletion,
): boolean {
  return completion[stepId];
}

export function firstIncompleteStep(
  completion: OnboardingCompletion,
  deferred: ReadonlySet<IntegrationStepId>,
): OnboardingStepId {
  for (const stepId of INTEGRATION_STEP_IDS) {
    if (!stepIsComplete(stepId, completion) && !deferred.has(stepId)) {
      return stepId;
    }
  }
  if (!deferred.has("scim") && !completion.scim) return "scim";
  if (!deferred.has("accessReview") && !completion.accessReview) return "accessReview";
  if (!deferred.has("agent") && !completion.agent) return "agent";
  return "mcp";
}

export function allIntegrationsComplete(completion: OnboardingCompletion): boolean {
  return completion.scim && completion.accessReview && completion.agent;
}
