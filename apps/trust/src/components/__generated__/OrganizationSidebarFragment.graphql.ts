/**
 * @generated SignedSource<<2d88ba7d4f5f2eea59c75dafbcc688d6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationSidebarFragment$data = {
  readonly audits: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"AuditRowFragment">;
      };
    }>;
  };
  readonly complianceBadges: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly iconUrl: string;
        readonly id: string;
        readonly name: string;
      };
    }>;
  };
  readonly darkLogoFileUrl: string | null | undefined;
  readonly logoFileUrl: string | null | undefined;
  readonly organization: {
    readonly description: string | null | undefined;
    readonly email: string | null | undefined;
    readonly headquarterAddress: string | null | undefined;
    readonly name: string;
    readonly websiteUrl: string | null | undefined;
  };
  readonly " $fragmentType": "OrganizationSidebarFragment";
};
export type OrganizationSidebarFragment$key = {
  readonly " $data"?: OrganizationSidebarFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OrganizationSidebarFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v1 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "OrganizationSidebarFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "logoFileUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "darkLogoFileUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Organization",
      "kind": "LinkedField",
      "name": "organization",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "description",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "websiteUrl",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "email",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "headquarterAddress",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "alias": null,
      "args": (v1/*: any*/),
      "concreteType": "ComplianceBadgeConnection",
      "kind": "LinkedField",
      "name": "complianceBadges",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ComplianceBadgeEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "ComplianceBadge",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v2/*: any*/),
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "iconUrl",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "complianceBadges(first:50)"
    },
    {
      "alias": null,
      "args": (v1/*: any*/),
      "concreteType": "AuditConnection",
      "kind": "LinkedField",
      "name": "audits",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "AuditEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Audit",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v2/*: any*/),
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "AuditRowFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "audits(first:50)"
    }
  ],
  "type": "TrustCenter",
  "abstractKey": null
};
})();

(node as any).hash = "cc8ab96a52d8a758dd1d7106921489ae";

export default node;
