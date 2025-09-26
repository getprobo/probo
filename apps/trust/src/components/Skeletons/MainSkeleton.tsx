import { Skeleton, TabLink, Tabs } from "@probo/ui";
import { TabSkeleton } from "./TabSkeleton";
import { useParams } from "react-router";
import { useTranslate } from "@probo/i18n";

export function MainSkeleton() {
  const { slug } = useParams<{ slug: string }>();
  const baseTabUrl = `/trust/${slug}`;
  const { __ } = useTranslate();
  return (
    <div className="grid grid-cols-1 max-w-[1280px] mx-4 pt-6 gap-4 lg:mx-auto lg:gap-10 lg:pt-20 lg:grid-cols-[400px_1fr] ">
      <Skeleton className="w-full h-300" />
      <main>
        <Tabs className="mb-8">
          <TabLink to={`${baseTabUrl}/overview`}>{__("Overview")}</TabLink>
          <TabLink to={`${baseTabUrl}/documents`}>{__("Documents")}</TabLink>
          <TabLink to={`${baseTabUrl}/subprocessors`}>
            {__("Subprocessors")}
          </TabLink>
        </Tabs>
        <TabSkeleton />
      </main>
    </div>
  );
}
