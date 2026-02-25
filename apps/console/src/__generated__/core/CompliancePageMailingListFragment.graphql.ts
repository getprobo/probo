/**
 * @generated SignedSource<<e9dd50980011bde2eb30f8c1f7fd5344>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MailingListSubscriberStatus = "CONFIRMED" | "PENDING";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageMailingListFragment$data = {
  readonly id: string;
  readonly mailingList: {
    readonly id: string;
    readonly subscribers: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly createdAt: string;
          readonly email: string;
          readonly fullName: string;
          readonly id: string;
          readonly status: MailingListSubscriberStatus;
        };
      }>;
      readonly pageInfo: {
        readonly endCursor: string | null | undefined;
        readonly hasNextPage: boolean;
      };
    };
  } | null | undefined;
  readonly " $fragmentType": "CompliancePageMailingListFragment";
};
export type CompliancePageMailingListFragment$key = {
  readonly " $data"?: CompliancePageMailingListFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageMailingListFragment">;
};

import CompliancePageMailingListQuery_graphql from './CompliancePageMailingListQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "mailingList",
  "subscribers"
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
      "defaultValue": 20,
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
        "node"
      ],
      "operation": CompliancePageMailingListQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "CompliancePageMailingListFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "MailingList",
      "kind": "LinkedField",
      "name": "mailingList",
      "plural": false,
      "selections": [
        (v1/*: any*/),
        {
          "alias": "subscribers",
          "args": null,
          "concreteType": "MailingListSubscriberConnection",
          "kind": "LinkedField",
          "name": "__CompliancePageMailingList_subscribers_connection",
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
                  "name": "endCursor",
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "MailingListSubscriberEdge",
              "kind": "LinkedField",
              "name": "edges",
              "plural": true,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "MailingListSubscriber",
                  "kind": "LinkedField",
                  "name": "node",
                  "plural": false,
                  "selections": [
                    (v1/*: any*/),
                    {
                      "alias": null,
                      "args": null,
                      "kind": "ScalarField",
                      "name": "fullName",
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
                      "name": "status",
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
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    (v1/*: any*/)
  ],
  "type": "TrustCenter",
  "abstractKey": null
};
})();

(node as any).hash = "b1828ed7ded38bb3e3e150b0875ea0a6";

export default node;
