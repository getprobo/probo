/**
 * @generated SignedSource<<dc9a8dba9780e827b273ced118578604>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY";
export type TrustCenterAccessGraphQuery$variables = {
  trustCenterId: string;
};
export type TrustCenterAccessGraphQuery$data = {
  readonly node: {
    readonly accesses?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly cursor: any;
        readonly node: {
          readonly active: boolean;
          readonly createdAt: any;
          readonly documentAccesses: {
            readonly edges: ReadonlyArray<{
              readonly node: {
                readonly active: boolean;
                readonly createdAt: any;
                readonly document: {
                  readonly documentType: DocumentType;
                  readonly id: string;
                  readonly title: string;
                } | null | undefined;
                readonly id: string;
                readonly report: {
                  readonly audit: {
                    readonly framework: {
                      readonly name: string;
                    };
                    readonly id: string;
                  } | null | undefined;
                  readonly filename: string;
                  readonly id: string;
                } | null | undefined;
                readonly updatedAt: any;
              };
            }>;
          };
          readonly email: string;
          readonly hasAcceptedNonDisclosureAgreement: boolean;
          readonly id: string;
          readonly name: string;
        };
      }>;
      readonly pageInfo: {
        readonly endCursor: any | null | undefined;
        readonly hasNextPage: boolean;
        readonly hasPreviousPage: boolean;
        readonly startCursor: any | null | undefined;
      };
    };
    readonly id?: string;
  };
};
export type TrustCenterAccessGraphQuery = {
  response: TrustCenterAccessGraphQuery$data;
  variables: TrustCenterAccessGraphQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "trustCenterId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "trustCenterId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "DESC",
    "field": "CREATED_AT"
  }
},
v4 = {
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
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "active",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasAcceptedNonDisclosureAgreement",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v11 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  },
  (v3/*: any*/)
],
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
  "concreteType": "Document",
  "kind": "LinkedField",
  "name": "document",
  "plural": false,
  "selections": [
    (v2/*: any*/),
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
    }
  ],
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "filename",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v16 = {
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
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterAccessGraphQuery",
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
              (v2/*: any*/),
              {
                "alias": "accesses",
                "args": [
                  (v3/*: any*/)
                ],
                "concreteType": "TrustCenterAccessConnection",
                "kind": "LinkedField",
                "name": "__TrustCenterAccessTab_accesses_connection",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterAccessEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      (v5/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterAccess",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v6/*: any*/),
                          (v7/*: any*/),
                          (v8/*: any*/),
                          (v9/*: any*/),
                          (v10/*: any*/),
                          {
                            "alias": null,
                            "args": (v11/*: any*/),
                            "concreteType": "TrustCenterDocumentAccessConnection",
                            "kind": "LinkedField",
                            "name": "documentAccesses",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "TrustCenterDocumentAccessEdge",
                                "kind": "LinkedField",
                                "name": "edges",
                                "plural": true,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "TrustCenterDocumentAccess",
                                    "kind": "LinkedField",
                                    "name": "node",
                                    "plural": false,
                                    "selections": [
                                      (v2/*: any*/),
                                      (v8/*: any*/),
                                      (v10/*: any*/),
                                      (v12/*: any*/),
                                      (v13/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "concreteType": "Report",
                                        "kind": "LinkedField",
                                        "name": "report",
                                        "plural": false,
                                        "selections": [
                                          (v2/*: any*/),
                                          (v14/*: any*/),
                                          {
                                            "alias": null,
                                            "args": null,
                                            "concreteType": "Audit",
                                            "kind": "LinkedField",
                                            "name": "audit",
                                            "plural": false,
                                            "selections": [
                                              (v2/*: any*/),
                                              {
                                                "alias": null,
                                                "args": null,
                                                "concreteType": "Framework",
                                                "kind": "LinkedField",
                                                "name": "framework",
                                                "plural": false,
                                                "selections": [
                                                  (v7/*: any*/)
                                                ],
                                                "storageKey": null
                                              }
                                            ],
                                            "storageKey": null
                                          }
                                        ],
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "documentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                          },
                          (v15/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  (v16/*: any*/)
                ],
                "storageKey": "__TrustCenterAccessTab_accesses_connection(orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              }
            ],
            "type": "TrustCenter",
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
    "name": "TrustCenterAccessGraphQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v15/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v11/*: any*/),
                "concreteType": "TrustCenterAccessConnection",
                "kind": "LinkedField",
                "name": "accesses",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterAccessEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      (v5/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterAccess",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v6/*: any*/),
                          (v7/*: any*/),
                          (v8/*: any*/),
                          (v9/*: any*/),
                          (v10/*: any*/),
                          {
                            "alias": null,
                            "args": (v11/*: any*/),
                            "concreteType": "TrustCenterDocumentAccessConnection",
                            "kind": "LinkedField",
                            "name": "documentAccesses",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "TrustCenterDocumentAccessEdge",
                                "kind": "LinkedField",
                                "name": "edges",
                                "plural": true,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "TrustCenterDocumentAccess",
                                    "kind": "LinkedField",
                                    "name": "node",
                                    "plural": false,
                                    "selections": [
                                      (v2/*: any*/),
                                      (v8/*: any*/),
                                      (v10/*: any*/),
                                      (v12/*: any*/),
                                      (v13/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "concreteType": "Report",
                                        "kind": "LinkedField",
                                        "name": "report",
                                        "plural": false,
                                        "selections": [
                                          (v2/*: any*/),
                                          (v14/*: any*/),
                                          {
                                            "alias": null,
                                            "args": null,
                                            "concreteType": "Audit",
                                            "kind": "LinkedField",
                                            "name": "audit",
                                            "plural": false,
                                            "selections": [
                                              (v2/*: any*/),
                                              {
                                                "alias": null,
                                                "args": null,
                                                "concreteType": "Framework",
                                                "kind": "LinkedField",
                                                "name": "framework",
                                                "plural": false,
                                                "selections": [
                                                  (v7/*: any*/),
                                                  (v2/*: any*/)
                                                ],
                                                "storageKey": null
                                              }
                                            ],
                                            "storageKey": null
                                          }
                                        ],
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "documentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                          },
                          (v15/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  (v16/*: any*/)
                ],
                "storageKey": "accesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "TrustCenterAccessTab_accesses",
                "kind": "LinkedHandle",
                "name": "accesses"
              }
            ],
            "type": "TrustCenter",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "72db0f61769f773d6f7ba20735f93643",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "accesses"
          ]
        }
      ]
    },
    "name": "TrustCenterAccessGraphQuery",
    "operationKind": "query",
    "text": "query TrustCenterAccessGraphQuery(\n  $trustCenterId: ID!\n) {\n  node(id: $trustCenterId) {\n    __typename\n    ... on TrustCenter {\n      id\n      accesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n        pageInfo {\n          hasNextPage\n          hasPreviousPage\n          startCursor\n          endCursor\n        }\n        edges {\n          cursor\n          node {\n            id\n            email\n            name\n            active\n            hasAcceptedNonDisclosureAgreement\n            createdAt\n            documentAccesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n              edges {\n                node {\n                  id\n                  active\n                  createdAt\n                  updatedAt\n                  document {\n                    id\n                    title\n                    documentType\n                  }\n                  report {\n                    id\n                    filename\n                    audit {\n                      id\n                      framework {\n                        name\n                        id\n                      }\n                    }\n                  }\n                }\n              }\n            }\n            __typename\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "a7c9ec0c6a2a5060886b68dc2be112fa";

export default node;
