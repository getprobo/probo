/**
 * @generated SignedSource<<769555856fa8bff973c4685249abbb2f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type PolicyStatus = "DRAFT" | "PUBLISHED" | "%future added value";
export type PolicyVersionSignatureState = "REQUESTED" | "SIGNED" | "%future added value";
import { FragmentRefs } from "relay-runtime";
export type PolicyPagePolicyFragment$data = {
  readonly id: string;
  readonly owner: {
    readonly fullName: string;
    readonly id: string;
  };
  readonly title: string;
  readonly versions: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly content: string;
        readonly id: string;
        readonly publishedAt: any | null | undefined;
        readonly signatures: {
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly state: PolicyVersionSignatureState;
            };
          }>;
        };
        readonly status: PolicyStatus;
        readonly updatedAt: any;
        readonly version: number;
        readonly " $fragmentSpreads": FragmentRefs<"PolicyVersionHistoryDialogFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "PolicyPagePolicyFragment";
};
export type PolicyPagePolicyFragment$key = {
  readonly " $data"?: PolicyPagePolicyFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"PolicyPagePolicyFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": [
          "versions"
        ]
      }
    ]
  },
  "name": "PolicyPagePolicyFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "owner",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fullName",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "alias": "versions",
      "args": null,
      "concreteType": "PolicyVersionConnection",
      "kind": "LinkedField",
      "name": "__PolicyPage_versions_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "PolicyVersionEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "PolicyVersion",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "content",
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
                  "name": "publishedAt",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "version",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "updatedAt",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "first",
                      "value": 100
                    }
                  ],
                  "concreteType": "PolicyVersionSignatureConnection",
                  "kind": "LinkedField",
                  "name": "signatures",
                  "plural": false,
                  "selections": [
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "PolicyVersionSignatureEdge",
                      "kind": "LinkedField",
                      "name": "edges",
                      "plural": true,
                      "selections": [
                        {
                          "alias": null,
                          "args": null,
                          "concreteType": "PolicyVersionSignature",
                          "kind": "LinkedField",
                          "name": "node",
                          "plural": false,
                          "selections": [
                            (v0/*: any*/),
                            {
                              "alias": null,
                              "args": null,
                              "kind": "ScalarField",
                              "name": "state",
                              "storageKey": null
                            }
                          ],
                          "storageKey": null
                        }
                      ],
                      "storageKey": null
                    }
                  ],
                  "storageKey": "signatures(first:100)"
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "PolicyVersionHistoryDialogFragment"
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
  "type": "Policy",
  "abstractKey": null
};
})();

(node as any).hash = "684db4b721c50638ee5ec1f9f4ebdd16";

export default node;
