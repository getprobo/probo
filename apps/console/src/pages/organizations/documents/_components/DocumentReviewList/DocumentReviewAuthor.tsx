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

import { relativeDateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { CardButton } from "@probo/ui/src/v2/Card/CardButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useContext } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentReviewAuthor_profile$key } from "#/__generated__/core/DocumentReviewAuthor_profile.graphql";

import { documentReviewAuthor } from "../../variants";

import { DocumentReviewItemContext } from "./documentReviewItemContext";

const documentReviewAuthorFragment = graphql`
  fragment DocumentReviewAuthor_profile on Profile {
    fullName
    emailAddress
    avatar {
      downloadUrl
    }
  }
`;

interface DocumentReviewAuthorProps {
  profileKey: DocumentReviewAuthor_profile$key;
  createdAt: string;
  // The card's title row: its author name is the card's stretched button.
  primary?: boolean;
  trailing?: ReactNode;
}

export function DocumentReviewAuthor({
  profileKey,
  createdAt,
  primary = false,
  trailing,
}: DocumentReviewAuthorProps) {
  const profile = useFragment(documentReviewAuthorFragment, profileKey);
  const item = useContext(DocumentReviewItemContext);
  const { i18n, t } = useTranslation("organizations/documents");
  const createdLabel = relativeDateFormat(i18n.language, createdAt)
    || t("details.review.justNow");
  const { root, identity, name, time, trailing: trailingClass } = documentReviewAuthor();
  const nameText = (
    <Text size={1} weight="medium" className={name()}>
      {profile.fullName}
    </Text>
  );

  return (
    <div className={root()}>
      <Avatar
        name={profile.fullName}
        email={profile.emailAddress}
        src={profile.avatar?.downloadUrl}
        size={1}
        radius="full"
      />
      <div className={identity()}>
        {primary && item != null
          ? (
              <CardButton
                aria-label={item.kind === "comment"
                  ? t("details.review.openComment", { name: profile.fullName })
                  : t("details.review.openSuggestion", { name: profile.fullName })}
                aria-pressed={item.active}
                className={name()}
                onClick={item.onActivate}
              >
                {nameText}
              </CardButton>
            )
          : nameText}
        <Text size={1} color="faint" className={time()}>
          {createdLabel}
        </Text>
      </div>
      {trailing != null && <div className={trailingClass()}>{trailing}</div>}
    </div>
  );
}
