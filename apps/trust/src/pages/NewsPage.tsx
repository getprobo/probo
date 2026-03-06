import { useTranslate } from "@probo/i18n";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { TrustGraphCurrentNewsQuery } from "#/queries/__generated__/TrustGraphCurrentNewsQuery.graphql";
import { currentTrustNewsQuery } from "#/queries/TrustGraph";

type Props = {
  queryRef: PreloadedQuery<TrustGraphCurrentNewsQuery>;
};

export function NewsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery<TrustGraphCurrentNewsQuery>(
    currentTrustNewsQuery,
    queryRef,
  );

  const items
    = data.currentTrustCenter?.updates.edges.map(e => e.node) ?? [];

  return (
    <div>
      <h2 className="font-medium mb-1">{__("News")}</h2>
      <p className="text-sm text-txt-secondary mb-4">
        {__("Latest compliance and security news:")}
      </p>
      <div className="space-y-0">
        {items.map(item => (
          <NewsItem key={item.id} item={item} />
        ))}
      </div>
    </div>
  );
}

type NewsItemType = {
  id: string;
  title: string;
  body: string;
  updatedAt: string;
};

function NewsItem({ item }: { item: NewsItemType }) {
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
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            className={`flex-none text-txt-tertiary transition-transform duration-200 ${open ? "rotate-180" : ""}`}
          >
            <path
              d="M4 6l4 4 4-4"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
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
