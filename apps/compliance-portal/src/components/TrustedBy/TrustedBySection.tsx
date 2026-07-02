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

import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { HomeSection } from "#/components/HomeSection/HomeSection";

import type { TrustedBySection_trustCenter$key } from "./__generated__/TrustedBySection_trustCenter.graphql";
import { TrustCenterReferenceListItem } from "./TrustCenterReferenceListItem";

const trustedBySectionFragment = graphql`
  fragment TrustedBySection_trustCenter on TrustCenter {
    references(first: 12) {
      edges {
        node {
          id
          ...TrustCenterReferenceListItem_reference
        }
      }
    }
  }
`;

interface TrustedBySectionProps {
  trustCenterKey: TrustedBySection_trustCenter$key;
}

// "Trusted by" section: a grid of customer / reference logos.
export function TrustedBySection({ trustCenterKey }: TrustedBySectionProps) {
  const { t } = useTranslation();
  const data = useFragment(trustedBySectionFragment, trustCenterKey);
  const references = data.references.edges.map(edge => edge.node);

  if (references.length === 0) {
    return null;
  }

  return (
    <HomeSection title={t("home.sections.trustedBy")}>
      <div className="grid grid-cols-6 gap-4">
        {references.map(reference => (
          <TrustCenterReferenceListItem key={reference.id} referenceKey={reference} />
        ))}
      </div>
    </HomeSection>
  );
}
