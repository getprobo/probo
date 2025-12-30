/**
 * @generated SignedSource<<5b0e12ba40232dd4328c0fd766513fb2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TrustCenterAccessGraph_accesses$data = {
  readonly accesses: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly cursor: any;
      readonly node: {
        readonly active: boolean;
        readonly activeCount: number;
        readonly createdAt: any;
        readonly email: any;
        readonly hasAcceptedNonDisclosureAgreement: boolean;
        readonly id: string;
        readonly lastTokenExpiresAt: any | null | undefined;
        readonly name: string;
        readonly pendingRequestCount: number;
      };
    }>;
    readonly pageInfo: {
      readonly endCursor: any | null | undefined;
      readonly hasNextPage: boolean;
      readonly hasPreviousPage: boolean;
      readonly startCursor: any | null | undefined;
    };
  };
  readonly id: string;
  readonly " $fragmentType": "TrustCenterAccessGraph_accesses";
};
export type TrustCenterAccessGraph_accesses$key = {
  readonly " $data"?: TrustCenterAccessGraph_accesses$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterAccessGraph_accesses">;
};

import TrustCenterAccessGraphPaginationQuery_graphql from './TrustCenterAccessGraphPaginationQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "accesses"
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
      "kind": "RootArgument",
      "name": "count"
    },
    {
      "kind": "RootArgument",
      "name": "cursor"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": "count",
        "cursor": "cursor",
        "direction": "forward",
        "path": (v0/*: any*/)
      }
    ],
    "refetch": {
      "connection": {
        "forward": {
          "count": "count",
          "cursor": "cursor"
        },
        "backward": null,
        "path": (v0/*: any*/)
      },
      "fragmentPathInResult": [
        "node"
      ],
      "operation": TrustCenterAccessGraphPaginationQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "TrustCenterAccessGraph_accesses",
  "selections": [
    {
      "alias": "accesses",
      "args": [
        {
          "kind": "Literal",
          "name": "orderBy",
          "value": {
            "direction": "DESC",
            "field": "CREATED_AT"
          }
        }
      ],
      "concreteType": "TrustCenterAccessConnection",
      "kind": "LinkedField",
      "name": "__TrustCenterAccessGraph_accesses_connection",
      "plural": false,
      "selections": [
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
              "name": "hasPreviousPage",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "startCursor",
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
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "TrustCenterAccessEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "cursor",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "TrustCenterAccess",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v1/*: any*/),
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
                  "name": "name",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "active",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "hasAcceptedNonDisclosureAgreement",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "createdAt",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "lastTokenExpiresAt",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "pendingRequestCount",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "activeCount",
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
            }
          ],
          "storageKey": null
        },
        {
          "kind": "ClientExtension",
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "__id",
              "storageKey": null
            }
          ]
        }
      ],
      "storageKey": "__TrustCenterAccessGraph_accesses_connection(orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
    },
    (v1/*: any*/)
  ],
  "type": "TrustCenter",
  "abstractKey": null
};
})();

(node as any).hash = "10bb084980182a1e2334be77c29cf1ec";

export default node;
