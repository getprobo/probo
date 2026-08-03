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

import type { ReactNode } from "react";
import { createContext, useCallback, useContext, useMemo, useState } from "react";

import type { IntegrationStepId } from "./onboardingSteps";

type OnboardingSessionContextValue = {
  deferredSteps: ReadonlySet<IntegrationStepId>;
  deferStep: (stepId: IntegrationStepId) => void;
  clearDeferred: (stepId: IntegrationStepId) => void;
};

const OnboardingSessionContext = createContext<OnboardingSessionContextValue | null>(
  null,
);

export function OnboardingSessionProvider(props: { children: ReactNode }) {
  const { children } = props;
  const [deferredSteps, setDeferredSteps] = useState<ReadonlySet<IntegrationStepId>>(
    () => new Set(),
  );

  const deferStep = useCallback((stepId: IntegrationStepId) => {
    setDeferredSteps(prev => new Set(prev).add(stepId));
  }, []);

  const clearDeferred = useCallback((stepId: IntegrationStepId) => {
    setDeferredSteps((prev) => {
      if (!prev.has(stepId)) return prev;
      const next = new Set(prev);
      next.delete(stepId);
      return next;
    });
  }, []);

  const value = useMemo(
    () => ({ deferredSteps, deferStep, clearDeferred }),
    [clearDeferred, deferStep, deferredSteps],
  );

  return (
    <OnboardingSessionContext.Provider value={value}>
      {children}
    </OnboardingSessionContext.Provider>
  );
}

export function useOnboardingSession(): OnboardingSessionContextValue {
  const ctx = useContext(OnboardingSessionContext);
  if (!ctx) {
    throw new Error("useOnboardingSession must be used within OnboardingSessionProvider");
  }
  return ctx;
}
