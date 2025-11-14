/**
 * @generated SignedSource<<73a2215bea77d40a4fb5ce4b5fada8a3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturesTabQuery$variables = {
  documentId: string;
  hasVersionId: boolean;
  versionId: string;
};
export type DocumentSignaturesTabQuery$data = {
  readonly document?: {
    readonly id?: string;
    readonly versions?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesTab_version">;
        };
      }>;
    };
  };
  readonly version?: {
    readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesTab_version">;
  };
};
export type DocumentSignaturesTabQuery = {
  response: DocumentSignaturesTabQuery$data;
  variables: DocumentSignaturesTabQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "documentId"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "hasVersionId"
},
v2 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "versionId"
},
v3 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "documentId"
  }
],
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v5 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1
  }
],
v6 = {
  "args": null,
  "kind": "FragmentSpread",
  "name": "DocumentSignaturesTab_version"
},
v7 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "versionId"
  }
],
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v10 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  }
],
v11 = {
  "alias": null,
  "args": (v10/*: any*/),
  "concreteType": "DocumentVersionSignatureConnection",
  "kind": "LinkedField",
  "name": "signatures",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "DocumentVersionSignatureEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "DocumentVersionSignature",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v4/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "state",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "People",
              "kind": "LinkedField",
              "name": "signedBy",
              "plural": false,
              "selections": [
                (v4/*: any*/),
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
                  "name": "primaryEmailAddress",
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "signedAt",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "requestedAt",
              "storageKey": null
            },
            (v8/*: any*/)
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
  "storageKey": "signatures(first:1000)"
},
v12 = {
  "alias": null,
  "args": (v10/*: any*/),
  "filters": [
    "filter"
  ],
  "handle": "connection",
  "key": "DocumentSignaturesTab_signatures",
  "kind": "LinkedHandle",
  "name": "signatures"
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v2/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentSignaturesTabQuery",
    "selections": [
      {
        "condition": "hasVersionId",
        "kind": "Condition",
        "passingValue": false,
        "selections": [
          {
            "alias": "document",
            "args": (v3/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              {
                "kind": "InlineFragment",
                "selections": [
                  (v4/*: any*/),
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
                    "concreteType": "DocumentVersionConnection",
                    "kind": "LinkedField",
                    "name": "versions",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "DocumentVersionEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "DocumentVersion",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v4/*: any*/),
                              (v6/*: any*/)
                            ],
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "versions(first:1)"
                  }
                ],
                "type": "Document",
                "abstractKey": null
              }
            ],
            "storageKey": null
          }
        ]
      },
      {
        "condition": "hasVersionId",
        "kind": "Condition",
        "passingValue": true,
        "selections": [
          {
            "alias": "version",
            "args": (v7/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              {
                "kind": "InlineFragment",
                "selections": [
                  (v6/*: any*/)
                ],
                "type": "DocumentVersion",
                "abstractKey": null
              }
            ],
            "storageKey": null
          }
        ]
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v2/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Operation",
    "name": "DocumentSignaturesTabQuery",
    "selections": [
      {
        "condition": "hasVersionId",
        "kind": "Condition",
        "passingValue": false,
        "selections": [
          {
            "alias": "document",
            "args": (v3/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v8/*: any*/),
              (v4/*: any*/),
              {
                "kind": "InlineFragment",
                "selections": [
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
                    "concreteType": "DocumentVersionConnection",
                    "kind": "LinkedField",
                    "name": "versions",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "DocumentVersionEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "DocumentVersion",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v4/*: any*/),
                              (v9/*: any*/),
                              (v11/*: any*/),
                              (v12/*: any*/)
                            ],
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "versions(first:1)"
                  }
                ],
                "type": "Document",
                "abstractKey": null
              }
            ],
            "storageKey": null
          }
        ]
      },
      {
        "condition": "hasVersionId",
        "kind": "Condition",
        "passingValue": true,
        "selections": [
          {
            "alias": "version",
            "args": (v7/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v8/*: any*/),
              (v4/*: any*/),
              {
                "kind": "InlineFragment",
                "selections": [
                  (v9/*: any*/),
                  (v11/*: any*/),
                  (v12/*: any*/)
                ],
                "type": "DocumentVersion",
                "abstractKey": null
              }
            ],
            "storageKey": null
          }
        ]
      }
    ]
  },
  "params": {
    "cacheID": "f2c160324eac9d2c557fb0fd151db60f",
    "id": null,
    "metadata": {},
    "name": "DocumentSignaturesTabQuery",
    "operationKind": "query",
    "text": "query DocumentSignaturesTabQuery(\n  $documentId: ID!\n  $versionId: ID!\n  $hasVersionId: Boolean!\n) {\n  document: node(id: $documentId) @skip(if: $hasVersionId) {\n    __typename\n    ... on Document {\n      id\n      versions(first: 1) {\n        edges {\n          node {\n            id\n            ...DocumentSignaturesTab_version\n          }\n        }\n      }\n    }\n    id\n  }\n  version: node(id: $versionId) @include(if: $hasVersionId) {\n    __typename\n    ... on DocumentVersion {\n      ...DocumentSignaturesTab_version\n    }\n    id\n  }\n}\n\nfragment DocumentSignaturesTab_signature on DocumentVersionSignature {\n  id\n  state\n  signedAt\n  requestedAt\n  signedBy {\n    fullName\n    primaryEmailAddress\n    id\n  }\n}\n\nfragment DocumentSignaturesTab_version on DocumentVersion {\n  id\n  status\n  signatures(first: 1000) {\n    edges {\n      node {\n        id\n        state\n        signedBy {\n          id\n          fullName\n          primaryEmailAddress\n        }\n        ...DocumentSignaturesTab_signature\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2fce1e9356142830ec022c7c625bd9f1";

export default node;
