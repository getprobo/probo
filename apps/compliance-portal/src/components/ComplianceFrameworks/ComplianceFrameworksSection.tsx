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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { InlineError } from "@probo/ui/src/v2/InlineError/InlineError";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { HomeSection } from "#/components/HomeSection/HomeSection";

import type { ComplianceFrameworksSection_trustCenter$key } from "./__generated__/ComplianceFrameworksSection_trustCenter.graphql";
import { ComplianceFrameworkListItem } from "./ComplianceFrameworkListItem";

// @throwOnFieldError makes a field error in this fragment throw at the read
// below, where the section's ErrorBoundary contains it. See
// contrib/claude/error-handling.md.
const complianceFrameworksSectionFragment = graphql`
  fragment ComplianceFrameworksSection_trustCenter on TrustCenter @throwOnFieldError {
    complianceFrameworks(first: 8) {
      edges {
        node {
          id
          ...ComplianceFrameworkListItem_complianceFramework
        }
      }
    }
  }
`;

interface ComplianceFrameworksSectionProps {
  trustCenterKey: ComplianceFrameworksSection_trustCenter$key;
}

// "Compliance" section: the grid of certification frameworks the trust center
// covers. Wraps its data-reading content in a boundary so a load failure
// degrades to an inline error instead of taking down the page.
export function ComplianceFrameworksSection({ trustCenterKey }: ComplianceFrameworksSectionProps) {
  const { t } = useTranslation();

  return (
    <ErrorBoundary
      fallback={(_, reset) => (
        <HomeSection title={t("home.sections.compliance")}>
          <InlineError
            message={t("errors.inline.message")}
            retryLabel={t("errors.inline.retry")}
            onRetry={reset}
          />
        </HomeSection>
      )}
    >
      <ComplianceFrameworksSectionContent trustCenterKey={trustCenterKey} />
    </ErrorBoundary>
  );
}

function ComplianceFrameworksSectionContent({ trustCenterKey }: ComplianceFrameworksSectionProps) {
  const { t } = useTranslation();
  const data = useFragment(complianceFrameworksSectionFragment, trustCenterKey);
  const frameworks = data.complianceFrameworks.edges.map(edge => edge.node);

  if (frameworks.length === 0) {
    return null;
  }

  return (
    <HomeSection title={t("home.sections.compliance")}>
      <div className="grid grid-cols-6 gap-4">
        {frameworks.map(framework => (
          <ComplianceFrameworkListItem key={framework.id} complianceFrameworkKey={framework} />
        ))}
      </div>
    </HomeSection>
  );
}
