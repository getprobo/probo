/**
 * @generated SignedSource<<f73f6da9282db845b9a9b03738304b8b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ComplianceFrameworkVisibility = "NONE" | "PUBLIC";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageFrameworkList_compliancePageFragment$data = {
  readonly canUpdate: boolean;
  readonly complianceFrameworks: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly framework: {
          readonly darkLogoURL: string | null | undefined;
          readonly id: string;
          readonly lightLogoURL: string | null | undefined;
          readonly name: string;
        };
        readonly id: string;
        readonly rank: number;
        readonly visibility: ComplianceFrameworkVisibility;
      };
    }>;
  };
  readonly id: string;
  readonly " $fragmentType": "CompliancePageFrameworkList_compliancePageFragment";
};
export type CompliancePageFrameworkList_compliancePageFragment$key = {
  readonly " $data"?: CompliancePageFrameworkList_compliancePageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageFrameworkList_compliancePageFragment">;
};

import CompliancePageFrameworkList_compliancePageRefetchQuery_graphql from './CompliancePageFrameworkList_compliancePageRefetchQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "complianceFrameworks"
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "after"
    },
    {
      "defaultValue": 100,
      "kind": "LocalArgument",
      "name": "first"
    },
    {
      "defaultValue": {
        "direction": "ASC",
        "field": "RANK"
      },
      "kind": "LocalArgument",
      "name": "order"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": "first",
        "cursor": "after",
        "direction": "forward",
        "path": (v0/*: any*/)
      }
    ],
    "refetch": {
      "connection": {
        "forward": {
          "count": "first",
          "cursor": "after"
        },
        "backward": null,
        "path": (v0/*: any*/)
      },
      "fragmentPathInResult": [
        "node"
      ],
      "operation": CompliancePageFrameworkList_compliancePageRefetchQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "CompliancePageFrameworkList_compliancePageFragment",
  "selections": [
    (v1/*: any*/),
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:update\")"
    },
    {
      "alias": "complianceFrameworks",
      "args": [
        {
          "kind": "Variable",
          "name": "orderBy",
          "variableName": "order"
        }
      ],
      "concreteType": "ComplianceFrameworkConnection",
      "kind": "LinkedField",
      "name": "__CompliancePageFrameworkList_complianceFrameworks_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ComplianceFrameworkEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "ComplianceFramework",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v1/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "rank",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "visibility",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "Framework",
                  "kind": "LinkedField",
                  "name": "framework",
                  "plural": false,
                  "selections": [
                    (v1/*: any*/),
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "name",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "lightLogoURL",
                      "storageKey": null
                    },
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "darkLogoURL",
                      "storageKey": null
                    }
                  ],
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "__typename",
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "cursor",
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "PageInfo",
          "kind": "LinkedField",
          "name": "pageInfo",
          "plural": false,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "endCursor",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "hasNextPage",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "TrustCenter",
  "abstractKey": null
};
})();

(node as any).hash = "b67fb402ac09aee269e109129ad129a2";

export default node;
