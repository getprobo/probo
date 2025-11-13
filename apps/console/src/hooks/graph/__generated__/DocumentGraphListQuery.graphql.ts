/**
 * @generated SignedSource<<abc733a779bfad444e7dfb11e6bca243>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentGraphListQuery$variables = {
  includeSignatures?: boolean | null | undefined;
  organizationId: string;
  useRequestedDocuments?: boolean | null | undefined;
};
export type DocumentGraphListQuery$data = {
  readonly organization: {
    readonly id: string;
    readonly " $fragmentSpreads": FragmentRefs<"DocumentsPageListFragment" | "DocumentsPageRequestedListFragment">;
  };
};
export type DocumentGraphListQuery = {
  response: DocumentGraphListQuery$data;
  variables: DocumentGraphListQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": false,
  "kind": "LocalArgument",
  "name": "includeSignatures"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "organizationId"
},
v2 = {
  "defaultValue": false,
  "kind": "LocalArgument",
  "name": "useRequestedDocuments"
},
v3 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
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
    "kind": "Variable",
    "name": "includeSignatures",
    "variableName": "includeSignatures"
  }
],
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v7 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "ASC",
      "field": "TITLE"
    }
  }
],
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "title",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "documentType",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "classification",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v4/*: any*/),
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
v14 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1
  }
],
v15 = [
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
            "name": "version",
            "storageKey": null
          },
          {
            "condition": "includeSignatures",
            "kind": "Condition",
            "passingValue": true,
            "selections": [
              {
                "alias": null,
                "args": [
                  {
                    "kind": "Literal",
                    "name": "first",
                    "value": 1000
                  }
                ],
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
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "signatures(first:1000)"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
],
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v17 = {
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
    }
  ],
  "storageKey": null
},
v18 = {
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
v19 = [
  "orderBy"
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v2/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentGraphListQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v3/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          {
            "condition": "useRequestedDocuments",
            "kind": "Condition",
            "passingValue": false,
            "selections": [
              {
                "args": (v5/*: any*/),
                "kind": "FragmentSpread",
                "name": "DocumentsPageListFragment"
              }
            ]
          },
          {
            "condition": "useRequestedDocuments",
            "kind": "Condition",
            "passingValue": true,
            "selections": [
              {
                "args": (v5/*: any*/),
                "kind": "FragmentSpread",
                "name": "DocumentsPageRequestedListFragment"
              }
            ]
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
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/),
      (v2/*: any*/)
    ],
    "kind": "Operation",
    "name": "DocumentGraphListQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v3/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v6/*: any*/),
          (v4/*: any*/),
          {
            "condition": "useRequestedDocuments",
            "kind": "Condition",
            "passingValue": false,
            "selections": [
              {
                "kind": "InlineFragment",
                "selections": [
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "concreteType": "DocumentConnection",
                    "kind": "LinkedField",
                    "name": "documents",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "DocumentEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Document",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v4/*: any*/),
                              (v8/*: any*/),
                              (v9/*: any*/),
                              (v10/*: any*/),
                              (v11/*: any*/),
                              (v12/*: any*/),
                              (v13/*: any*/),
                              {
                                "alias": null,
                                "args": (v14/*: any*/),
                                "concreteType": "DocumentVersionConnection",
                                "kind": "LinkedField",
                                "name": "versions",
                                "plural": false,
                                "selections": (v15/*: any*/),
                                "storageKey": "versions(first:1)"
                              },
                              (v6/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v16/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v17/*: any*/),
                      (v18/*: any*/)
                    ],
                    "storageKey": "documents(first:50,orderBy:{\"direction\":\"ASC\",\"field\":\"TITLE\"})"
                  },
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "filters": (v19/*: any*/),
                    "handle": "connection",
                    "key": "DocumentsListQuery_documents",
                    "kind": "LinkedHandle",
                    "name": "documents"
                  }
                ],
                "type": "Organization",
                "abstractKey": null
              }
            ]
          },
          {
            "condition": "useRequestedDocuments",
            "kind": "Condition",
            "passingValue": true,
            "selections": [
              {
                "kind": "InlineFragment",
                "selections": [
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "concreteType": "DocumentConnection",
                    "kind": "LinkedField",
                    "name": "requestedDocuments",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "DocumentEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Document",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v4/*: any*/),
                              (v8/*: any*/),
                              (v9/*: any*/),
                              (v10/*: any*/),
                              (v11/*: any*/),
                              (v12/*: any*/),
                              (v13/*: any*/),
                              {
                                "alias": null,
                                "args": (v14/*: any*/),
                                "concreteType": "DocumentVersionConnection",
                                "kind": "LinkedField",
                                "name": "requestedVersions",
                                "plural": false,
                                "selections": (v15/*: any*/),
                                "storageKey": "requestedVersions(first:1)"
                              },
                              (v6/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v16/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v17/*: any*/),
                      (v18/*: any*/)
                    ],
                    "storageKey": "requestedDocuments(first:50,orderBy:{\"direction\":\"ASC\",\"field\":\"TITLE\"})"
                  },
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "filters": (v19/*: any*/),
                    "handle": "connection",
                    "key": "DocumentsRequestedListQuery_requestedDocuments",
                    "kind": "LinkedHandle",
                    "name": "requestedDocuments"
                  }
                ],
                "type": "Organization",
                "abstractKey": null
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "e61cf81ab03f70a35af3d1c22730b0a1",
    "id": null,
    "metadata": {},
    "name": "DocumentGraphListQuery",
    "operationKind": "query",
    "text": "query DocumentGraphListQuery(\n  $organizationId: ID!\n  $includeSignatures: Boolean = false\n  $useRequestedDocuments: Boolean = false\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    id\n    ...DocumentsPageListFragment_ipv6e @skip(if: $useRequestedDocuments)\n    ...DocumentsPageRequestedListFragment_ipv6e @include(if: $useRequestedDocuments)\n  }\n}\n\nfragment DocumentsPageListFragment_ipv6e on Organization {\n  documents(first: 50, orderBy: {field: TITLE, direction: ASC}) {\n    edges {\n      node {\n        id\n        ...DocumentsPageRowFragment_2ykUZW\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment DocumentsPageRequestedListFragment_ipv6e on Organization {\n  requestedDocuments(first: 50, orderBy: {field: TITLE, direction: ASC}) {\n    edges {\n      node {\n        id\n        ...DocumentsPageRowFragment_1L3zd7\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment DocumentsPageRowFragment_1L3zd7 on Document {\n  id\n  title\n  description\n  documentType\n  classification\n  updatedAt\n  owner {\n    id\n    fullName\n  }\n  requestedVersions(first: 1) {\n    edges {\n      node {\n        id\n        status\n        version\n        signatures(first: 1000) @include(if: $includeSignatures) {\n          edges {\n            node {\n              id\n              state\n            }\n          }\n        }\n      }\n    }\n  }\n}\n\nfragment DocumentsPageRowFragment_2ykUZW on Document {\n  id\n  title\n  description\n  documentType\n  classification\n  updatedAt\n  owner {\n    id\n    fullName\n  }\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n        version\n        signatures(first: 1000) @include(if: $includeSignatures) {\n          edges {\n            node {\n              id\n              state\n            }\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bcf14a11712d078d723206971379271f";

export default node;
