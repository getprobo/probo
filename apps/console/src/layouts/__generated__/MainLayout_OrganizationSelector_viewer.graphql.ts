/**
 * @generated SignedSource<<2eda1054dbabd2bb5e45d8db5c7a01fc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MainLayout_OrganizationSelector_viewer$data = {
  readonly organizations: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly logoUrl: string | null | undefined;
        readonly name: string;
      };
    }>;
    readonly pageInfo: {
      readonly endCursor: any | null | undefined;
      readonly hasNextPage: boolean;
    };
  };
  readonly " $fragmentType": "MainLayout_OrganizationSelector_viewer";
};
export type MainLayout_OrganizationSelector_viewer$key = {
  readonly " $data"?: MainLayout_OrganizationSelector_viewer$data;
  readonly " $fragmentSpreads": FragmentRefs<"MainLayout_OrganizationSelector_viewer">;
};

import MainLayoutOrganizationSelectorPaginationQuery_graphql from './MainLayoutOrganizationSelectorPaginationQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "organizations"
];
return {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "after"
    },
    {
      "defaultValue": 25,
      "kind": "LocalArgument",
      "name": "first"
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
        "viewer"
      ],
      "operation": MainLayoutOrganizationSelectorPaginationQuery_graphql
    }
  },
  "name": "MainLayout_OrganizationSelector_viewer",
  "selections": [
    {
      "alias": "organizations",
      "args": [
        {
          "kind": "Literal",
          "name": "orderBy",
          "value": {
            "direction": "ASC",
            "field": "NAME"
          }
        }
      ],
      "concreteType": "OrganizationConnection",
      "kind": "LinkedField",
      "name": "__MainLayout_OrganizationSelector_organizations_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "OrganizationEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Organization",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "id",
                  "storageKey": null
                },
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
                  "name": "logoUrl",
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
              "name": "hasNextPage",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "endCursor",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "__MainLayout_OrganizationSelector_organizations_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"NAME\"})"
    }
  ],
  "type": "Viewer",
  "abstractKey": null
};
})();

(node as any).hash = "0dc2841aeeb6ba5291cb45d8644ec4af";

export default node;
