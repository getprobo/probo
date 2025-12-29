import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";
import type { DomainSettingsPageQuery } from "./__generated__/DomainSettingsPageQuery.graphql";
import { useTranslate } from "@probo/i18n";
import { DomainCard } from "./_components/DomainCard";
import { NewDomainDialog } from "./_components/NewDomainDialog";
import { Button, Card, IconPlusLarge } from "@probo/ui";

export const domainSettingsPageQuery = graphql`
  query DomainSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        customDomain {
          domain
          ...DomainCardFragment
        }
      }
    }
  }
`;

export function DomainSettingsPage(props: {
  queryRef: PreloadedQuery<DomainSettingsPageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<DomainSettingsPageQuery>(
    domainSettingsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("Custom Domain")}</h2>
      {organization.customDomain ? (
        <DomainCard fKey={organization.customDomain} />
      ) : (
        <Card padded>
          <div className="text-center py-8">
            <h3 className="text-lg font-semibold mb-2">
              {__("No custom domain configured")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__(
                "Add your own domain to make your trust center more professional",
              )}
            </p>
            <div className="flex justify-center">
              {/* {isAuthorized("Organization", "createCustomDomain") && ( */}
              <NewDomainDialog>
                <Button icon={IconPlusLarge}>{__("Add Domain")}</Button>
              </NewDomainDialog>
              {/* )} */}
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}
