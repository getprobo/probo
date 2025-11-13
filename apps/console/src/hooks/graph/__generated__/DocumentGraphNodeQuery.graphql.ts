/**
 * @generated SignedSource<<d9c272a0703ff8fbc3eecfa14f3033c3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentGraphNodeQuery$variables = {
  documentId: string;
  includeControls: boolean;
  includeSignatures: boolean;
  useRequestedVersions?: boolean | null | undefined;
};
export type DocumentGraphNodeQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"DocumentDetailPageDocumentFragment">;
  };
};
export type DocumentGraphNodeQuery = {
  response: DocumentGraphNodeQuery$data;
  variables: DocumentGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "documentId"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "includeControls"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "includeSignatures"
  },
  {
    "defaultValue": false,
    "kind": "LocalArgument",
    "name": "useRequestedVersions"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "documentId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "classification",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v3/*: any*/),
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
v6 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 20
  }
],
v7 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  }
],
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v9 = {
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
v10 = {
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
},
v11 = [
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
          (v3/*: any*/),
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
          (v4/*: any*/),
          (v5/*: any*/),
          {
            "condition": "includeSignatures",
            "kind": "Condition",
            "passingValue": true,
            "selections": [
              {
                "alias": null,
                "args": (v7/*: any*/),
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
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "state",
                            "storageKey": null
                          },
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v8/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v9/*: any*/),
                  (v10/*: any*/)
                ],
                "storageKey": "signatures(first:1000)"
              },
              {
                "alias": null,
                "args": (v7/*: any*/),
                "filters": ([]/*: any*/),
                "handle": "connection",
                "key": "DocumentDetailPage_signatures",
                "kind": "LinkedHandle",
                "name": "signatures"
              }
            ]
          },
          (v2/*: any*/)
        ],
        "storageKey": null
      },
      (v8/*: any*/)
    ],
    "storageKey": null
  },
  (v9/*: any*/),
  (v10/*: any*/)
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "args": [
                  {
                    "kind": "Variable",
                    "name": "includeControls",
                    "variableName": "includeControls"
                  },
                  {
                    "kind": "Variable",
                    "name": "includeSignatures",
                    "variableName": "includeSignatures"
                  },
                  {
                    "kind": "Variable",
                    "name": "useRequestedVersions",
                    "variableName": "useRequestedVersions"
                  }
                ],
                "kind": "FragmentSpread",
                "name": "DocumentDetailPageDocumentFragment"
              }
            ],
            "type": "Document",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
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
                "kind": "ScalarField",
                "name": "documentType",
                "storageKey": null
              },
              (v4/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Organization",
                "kind": "LinkedField",
                "name": "organization",
                "plural": false,
                "selections": [
                  (v3/*: any*/)
                ],
                "storageKey": null
              },
              (v5/*: any*/),
              {
                "condition": "includeControls",
                "kind": "Condition",
                "passingValue": true,
                "selections": [
                  {
                    "alias": "controlsInfo",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "first",
                        "value": 0
                      }
                    ],
                    "concreteType": "ControlConnection",
                    "kind": "LinkedField",
                    "name": "controls",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "totalCount",
                        "storageKey": null
                      }
                    ],
                    "storageKey": "controls(first:0)"
                  }
                ]
              },
              {
                "condition": "useRequestedVersions",
                "kind": "Condition",
                "passingValue": false,
                "selections": [
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "DocumentVersionConnection",
                    "kind": "LinkedField",
                    "name": "versions",
                    "plural": false,
                    "selections": (v11/*: any*/),
                    "storageKey": "versions(first:20)"
                  },
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "filters": null,
                    "handle": "connection",
                    "key": "DocumentDetailPage_versions",
                    "kind": "LinkedHandle",
                    "name": "versions"
                  }
                ]
              },
              {
                "condition": "useRequestedVersions",
                "kind": "Condition",
                "passingValue": true,
                "selections": [
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "DocumentVersionConnection",
                    "kind": "LinkedField",
                    "name": "requestedVersions",
                    "plural": false,
                    "selections": (v11/*: any*/),
                    "storageKey": "requestedVersions(first:20)"
                  },
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "filters": null,
                    "handle": "connection",
                    "key": "DocumentDetailPage_requestedVersions",
                    "kind": "LinkedHandle",
                    "name": "requestedVersions"
                  }
                ]
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
  "params": {
    "cacheID": "a2546bb9174ab8078f5b14e57a64dd1e",
    "id": null,
    "metadata": {},
    "name": "DocumentGraphNodeQuery",
    "operationKind": "query",
    "text": "query DocumentGraphNodeQuery(\n  $documentId: ID!\n  $includeControls: Boolean!\n  $includeSignatures: Boolean!\n  $useRequestedVersions: Boolean = false\n) {\n  node(id: $documentId) {\n    __typename\n    ... on Document {\n      ...DocumentDetailPageDocumentFragment_3N33mI\n    }\n    id\n  }\n}\n\nfragment DocumentDetailPageDocumentFragment_3N33mI on Document {\n  id\n  title\n  documentType\n  classification\n  organization {\n    id\n  }\n  owner {\n    id\n    fullName\n  }\n  controlsInfo: controls(first: 0) @include(if: $includeControls) {\n    totalCount\n  }\n  versions(first: 20) @skip(if: $useRequestedVersions) {\n    edges {\n      node {\n        id\n        content\n        status\n        publishedAt\n        version\n        updatedAt\n        classification\n        owner {\n          id\n          fullName\n        }\n        signatures(first: 1000) @include(if: $includeSignatures) {\n          edges {\n            node {\n              id\n              state\n              __typename\n            }\n            cursor\n          }\n          pageInfo {\n            endCursor\n            hasNextPage\n          }\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  requestedVersions(first: 20) @include(if: $useRequestedVersions) {\n    edges {\n      node {\n        id\n        content\n        status\n        publishedAt\n        version\n        updatedAt\n        classification\n        owner {\n          id\n          fullName\n        }\n        signatures(first: 1000) @include(if: $includeSignatures) {\n          edges {\n            node {\n              id\n              state\n              __typename\n            }\n            cursor\n          }\n          pageInfo {\n            endCursor\n            hasNextPage\n          }\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5c3be4665b543875f6bc8a8f9c383b86";

export default node;
