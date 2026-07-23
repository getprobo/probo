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

import type { LinkedControlsDialogControlsQuery } from "#/__generated__/core/LinkedControlsDialogControlsQuery.graphql";
import type {
  LinkedControlsDialogFragment$data,
  LinkedControlsDialogFragment$key,
} from "#/__generated__/core/LinkedControlsDialogFragment.graphql";
import type { LinkedControlsDialogQuery } from "#/__generated__/core/LinkedControlsDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

const query = graphql`
  query LinkedControlsDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...LinkedControlsDialogFragment
    }
  }
`;

const controlsFragment = graphql`
  fragment LinkedControlsDialogFragment on Organization
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 1 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ControlOrder", defaultValue: null }
    filter: { type: "ControlFilter", defaultValue: null }
  )
  @refetchable(queryName: "LinkedControlsDialogControlsQuery") {
    controls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "LinkedControlsDialogControlsQuery_controls") {
      edges {
        node {
          id
          name
          sectionTitle
          framework {
            name
          }
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedControls?: { id: string }[];
  onLink: (controlId: string) => void;
  onUnlink: (controlId: string) => void;
};

type SearchRef = RefObject<{ search: (v: string) => void } | null>;

export function LinkedControlsDialog(props: Props) {
  const { t } = useTranslation();
  const searchRef: SearchRef = useRef(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [minHeight, setMinHeight] = useState(0);
  const onSearch = (v: string) => {
    setMinHeight(contentRef.current?.clientHeight ?? 0);
    searchRef.current?.search(v);
  };
  return (
    <Dialog trigger={props.children} title={t("linkedControlsDialog.title")}>
      <DialogContent>
        <div className="flex items-center gap-2 sticky top-0 py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
          <Input
            icon={IconMagnifyingGlass}
            placeholder={t("linkedControlsDialog.searchPlaceholder")}
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
            <LinkedControlsDialogContent {...props} ref={searchRef} />
          </Suspense>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function LinkedControlsDialogContent(props: Props & { ref: SearchRef }) {
  const organizationId = useOrganizationId();
  const mainData = useLazyLoadQuery<LinkedControlsDialogQuery>(query, {
    organizationId,
  });
  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginationFragment<LinkedControlsDialogControlsQuery, LinkedControlsDialogFragment$key>(
      controlsFragment,
      mainData.organization as LinkedControlsDialogFragment$key,
    );

  const controls = data.controls?.edges?.map(edge => edge.node) ?? [];
  const controlIds = useMemo(() => {
    return new Set(props.linkedControls?.map(c => c.id) ?? []);
  }, [props.linkedControls]);

  const handleSearch = useDebounceCallback((v: string) => {
    refetch({
      first: 20,
      filter: {
        query: v,
      },
    });
  }, 500);

  useEffect(() => {
    if (!props.ref.current) {
      props.ref.current = {
        search: handleSearch,
      };
    }
  });

  return (
    <>
      <div className="divide-y divide-border-low">
        {controls.map(control => (
          <ControlRow
            key={control.id}
            control={control}
            controlIds={controlIds}
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
    </>
  );
}

function ControlRow(
  props: {
    control: NodeOf<LinkedControlsDialogFragment$data["controls"]>;
    controlIds: Set<string>;
  } & Props,
) {
  const { t } = useTranslation();
  const isLinked = props.controlIds.has(props.control.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;
  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full text-start"
      onClick={() => onClick(props.control.id)}
    >
      {props.control.sectionTitle}
      {" "}
      :
      {props.control.name}
      <Badge>{props.control.framework.name}</Badge>
      <Button
        disabled={props.disabled}
        className="ml-auto"
        variant={isLinked ? "secondary" : "primary"}
        asChild
      >
        <span>
          <IconComponent size={16} />
          {" "}
          {isLinked
            ? t("linkedControlsDialog.actions.unlink")
            : t("linkedControlsDialog.actions.link")}
        </span>
      </Button>
    </button>
  );
}
