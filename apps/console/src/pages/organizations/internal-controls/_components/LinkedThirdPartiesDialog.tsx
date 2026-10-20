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

import { faviconUrl } from "@probo/helpers";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  InfiniteScrollTrigger,
  Input,
  Spinner,
} from "@probo/ui";
import {
  type ReactNode,
  type RefObject,
  Suspense,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type {
  LinkedThirdPartiesDialogFragment$data,
  LinkedThirdPartiesDialogFragment$key,
} from "#/__generated__/core/LinkedThirdPartiesDialogFragment.graphql";
import type { LinkedThirdPartiesDialogQuery } from "#/__generated__/core/LinkedThirdPartiesDialogQuery.graphql";
import type { LinkedThirdPartiesDialogRefetchQuery } from "#/__generated__/core/LinkedThirdPartiesDialogRefetchQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

const query = graphql`
  query LinkedThirdPartiesDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...LinkedThirdPartiesDialogFragment
    }
  }
`;

const thirdPartiesFragment = graphql`
  fragment LinkedThirdPartiesDialogFragment on Organization
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ThirdPartyOrder", defaultValue: null }
    filter: { type: "ThirdPartyFilter", defaultValue: { level: 1 } }
  )
  @refetchable(queryName: "LinkedThirdPartiesDialogRefetchQuery") {
    thirdParties(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "LinkedThirdPartiesDialogRefetchQuery_thirdParties", filters: ["filter"]) {
      edges {
        node {
          id
          name
          category
          websiteUrl
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedThirdParties?: { id: string }[];
  onLink: (thirdPartyId: string) => void;
  onUnlink: (thirdPartyId: string) => void;
};

type SearchRef = RefObject<{ search: (v: string) => void } | null>;

export function LinkedThirdPartiesDialog(props: Props) {
  const { t } = useTranslation();
  const searchRef: SearchRef = useRef(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [minHeight, setMinHeight] = useState(0);
  const onSearch = (v: string) => {
    setMinHeight(contentRef.current?.clientHeight ?? 0);
    searchRef.current?.search(v);
  };
  return (
    <Dialog trigger={props.children} title={t("linkedThirdPartiesDialog.title")}>
      <DialogContent>
        <div className="flex items-center gap-2 sticky top-0 py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
          <Input
            icon={IconMagnifyingGlass}
            placeholder={t("linkedThirdPartiesDialog.searchPlaceholder")}
            onValueChange={onSearch}
          />
        </div>
        <div ref={contentRef}>
          <Suspense
            fallback={(
              <div style={{ minHeight }}>
                <Spinner centered />
              </div>
            )}
          >
            <LinkedThirdPartiesDialogContent {...props} ref={searchRef} />
          </Suspense>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function LinkedThirdPartiesDialogContent({
  ref: searchRef,
  ...props
}: Props & { ref: SearchRef }) {
  const organizationId = useOrganizationId();
  const mainData = useLazyLoadQuery<LinkedThirdPartiesDialogQuery>(query, {
    organizationId,
  });
  const { data, loadNext, hasNext, isLoadingNext, refetch } = usePaginationFragment<
    LinkedThirdPartiesDialogRefetchQuery,
    LinkedThirdPartiesDialogFragment$key
  >(
    thirdPartiesFragment,
    mainData.organization as LinkedThirdPartiesDialogFragment$key,
  );

  const thirdParties = data.thirdParties?.edges?.map(edge => edge.node) ?? [];
  const linkedIds = useMemo(() => {
    return new Set(props.linkedThirdParties?.map(t => t.id) ?? []);
  }, [props.linkedThirdParties]);

  const handleSearch = useDebounceCallback((v: string) => {
    refetch({
      first: 20,
      filter: {
        level: 1,
        query: v,
      },
    });
  }, 500);

  useEffect(() => {
    searchRef.current = { search: handleSearch };
    return () => {
      searchRef.current = null;
    };
  }, [handleSearch, searchRef]);

  return (
    <div className="divide-y divide-border-low">
      {thirdParties.map(thirdParty => (
        <ThirdPartyRow
          key={thirdParty.id}
          thirdParty={thirdParty}
          linkedIds={linkedIds}
          {...props}
        />
      ))}
      {hasNext && (
        <InfiniteScrollTrigger
          loading={isLoadingNext}
          onView={() => loadNext(20)}
        />
      )}
    </div>
  );
}

function ThirdPartyRow(
  props: {
    thirdParty: NodeOf<LinkedThirdPartiesDialogFragment$data["thirdParties"]>;
    linkedIds: Set<string>;
  } & Props,
) {
  const { t } = useTranslation();
  const isLinked = props.linkedIds.has(props.thirdParty.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;
  const logo = faviconUrl(props.thirdParty.websiteUrl);
  return (
    <button
      type="button"
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full text-start disabled:cursor-not-allowed disabled:opacity-50"
      disabled={props.disabled}
      onClick={() => onClick(props.thirdParty.id)}
    >
      {logo && (
        <img src={logo} alt={props.thirdParty.name} className="rounded h-5 w-5" />
      )}
      {props.thirdParty.name}
      <Badge>{props.thirdParty.category}</Badge>
      <Button
        disabled={props.disabled}
        className="ml-auto"
        variant={isLinked ? "secondary" : "primary"}
        asChild
      >
        <span>
          <IconComponent size={16} />
          {" "}
          {isLinked ? t("linkedThirdPartiesDialog.actions.unlink") : t("linkedThirdPartiesDialog.actions.link")}
        </span>
      </Button>
    </button>
  );
}
