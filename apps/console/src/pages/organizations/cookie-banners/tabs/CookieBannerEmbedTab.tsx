import { useTranslate } from "@probo/i18n";
import { Button, Card } from "@probo/ui";
import { useCallback } from "react";
import { useOutletContext } from "react-router";

import type { CookieBannerGraphNodeQuery$data } from "#/__generated__/core/CookieBannerGraphNodeQuery.graphql";

export default function CookieBannerEmbedTab() {
  const { banner } = useOutletContext<{
    banner: CookieBannerGraphNodeQuery$data["node"];
  }>();

  const { __ } = useTranslate();

  const snippet = banner.embedSnippet ?? "";

  const handleCopy = useCallback(() => {
    void navigator.clipboard.writeText(snippet);
  }, [snippet]);

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Embed Snippet")}</h2>
        <p className="text-sm text-txt-secondary">
          {__(
            "Add this script tag to your website's HTML to display the cookie consent banner. Place it in the <head> section of your page.",
          )}
        </p>
        <Card padded>
          <pre className="overflow-x-auto text-sm font-mono bg-surface-secondary p-4 rounded-lg whitespace-pre-wrap break-all">
            {snippet}
          </pre>
          <div className="flex justify-end mt-4">
            <Button variant="secondary" onClick={handleCopy}>
              {__("Copy to clipboard")}
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
}
