import { useTranslate } from "@probo/i18n";
import { IconChevronDown } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import { Rows } from "#/components/Rows";
import type { UpdatesPageQuery } from "#/pages/__generated__/UpdatesPageQuery.graphql";

export const currentTrustUpdatesQuery = graphql`
  query UpdatesPageQuery {
    currentTrustCenter {
      id
      updates(first: 50) {
        edges {
          node {
            id
            title
            body
            updatedAt
          }
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<UpdatesPageQuery>;
};

export function UpdatesPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery<UpdatesPageQuery>(
    currentTrustUpdatesQuery,
    queryRef,
  );

  const items
    = data.currentTrustCenter?.updates.edges.map(e => e.node) ?? [];

  return (
    <div>
      <h2 className="font-medium mb-1">{__("Updates")}</h2>
      {items.length === 0
        ? (
            <Rows>
              <div className="text-sm text-txt-tertiary text-center py-5">
                {__("No updates have been published yet.")}
              </div>
            </Rows>
          )
        : (
            <>
              <p className="text-sm text-txt-secondary mb-4">
                {__("Latest compliance and security updates")}
              </p>
              <div className="space-y-0">
                {items.map(item => (
                  <UpdateItem key={item.id} item={item} />
                ))}
              </div>
            </>
          )}
    </div>
  );
}

type UpdateItemType = {
  id: string;
  title: string;
  body: string;
  updatedAt: string;
};

function UpdateItem({ item }: { item: UpdateItemType }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="border border-border-solid -mt-px first:rounded-t-lg last:rounded-b-lg overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center justify-between px-6 py-4 text-left hover:bg-highlight transition-colors group"
      >
        <div className="flex items-center gap-3 min-w-0">
          <span className="text-sm font-medium text-txt-primary truncate">
            {item.title}
          </span>
        </div>
        <div className="flex items-center gap-4 flex-none ml-4">
          <span className="text-xs text-txt-tertiary hidden sm:block">
            {new Date(item.updatedAt).toLocaleDateString(undefined, {
              year: "numeric",
              month: "short",
              day: "numeric",
            })}
          </span>
          <IconChevronDown
            size={16}
            className={`flex-none text-txt-tertiary transition-transform duration-200 ${open ? "rotate-180" : ""}`}
          />
        </div>
      </button>
      {open && (
        <div className="px-6 pb-5 pt-0 border-t border-border-solid bg-highlight/40">
          <p className="text-sm text-txt-secondary whitespace-pre-wrap leading-relaxed pt-4">
            {item.body}
          </p>
          <p className="text-xs text-txt-tertiary mt-3 sm:hidden">
            {new Date(item.updatedAt).toLocaleDateString(undefined, {
              year: "numeric",
              month: "short",
              day: "numeric",
            })}
          </p>
        </div>
      )}
    </div>
  );
}
