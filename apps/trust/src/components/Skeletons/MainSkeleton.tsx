import { getTrustCenterUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Skeleton, TabLink, Tabs } from "@probo/ui";

import { TabSkeleton } from "./TabSkeleton";

export function MainSkeleton() {
  const { __ } = useTranslate();
  return (
    <div className="grid grid-cols-1 max-w-[1280px] mx-4 pt-6 gap-4 lg:mx-auto lg:gap-10 lg:pt-20 lg:grid-cols-[400px_1fr] ">
      <Skeleton className="w-full h-300" />
      <main>
        <Tabs className="mb-8">
          <TabLink to={getTrustCenterUrl("overview")}>{__("Overview")}</TabLink>
          <TabLink to={getTrustCenterUrl("documents")}>{__("Documents")}</TabLink>
          <TabLink to={getTrustCenterUrl("subprocessors")}>
            {__("Subprocessors")}
          </TabLink>
        </Tabs>
        <TabSkeleton />
      </main>
    </div>
  );
}
