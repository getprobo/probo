/**
 * @generated SignedSource<<72bab693664bbf68879e863c85f980b5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustCenterDocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
export type TrustCenterAccessGraphLoadDocumentAccessesQuery$variables = {
  accessId: string;
};
export type TrustCenterAccessGraphLoadDocumentAccessesQuery$data = {
  readonly node: {
    readonly availableDocumentAccesses?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
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
          readonly status: TrustCenterDocumentAccessStatus;
          readonly trustCenterFile: {
            readonly category: string;
            readonly id: string;
            readonly name: string;
          } | null | undefined;
        };
      }>;
    };
    readonly id?: string;
  };
};
export type TrustCenterAccessGraphLoadDocumentAccessesQuery = {
  response: TrustCenterAccessGraphLoadDocumentAccessesQuery$data;
  variables: TrustCenterAccessGraphLoadDocumentAccessesQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "accessId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "accessId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "DESC",
      "field": "CREATED_AT"
    }
  }
],
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v5 = {
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
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "filename",
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
  "concreteType": "TrustCenterFile",
  "kind": "LinkedField",
  "name": "trustCenterFile",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v7/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "category",
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterAccessGraphLoadDocumentAccessesQuery",
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
                "alias": null,
                "args": (v3/*: any*/),
                "concreteType": "TrustCenterDocumentAccessConnection",
                "kind": "LinkedField",
                "name": "availableDocumentAccesses",
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
                          (v4/*: any*/),
                          (v5/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v6/*: any*/),
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
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "availableDocumentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              }
            ],
            "type": "TrustCenterAccess",
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
    "name": "TrustCenterAccessGraphLoadDocumentAccessesQuery",
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
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v3/*: any*/),
                "concreteType": "TrustCenterDocumentAccessConnection",
                "kind": "LinkedField",
                "name": "availableDocumentAccesses",
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
                          (v4/*: any*/),
                          (v5/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v6/*: any*/),
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
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "availableDocumentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              }
            ],
            "type": "TrustCenterAccess",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "a3c6d83c3fae9bd8ab3e474706423115",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphLoadDocumentAccessesQuery",
    "operationKind": "query",
    "text": "query TrustCenterAccessGraphLoadDocumentAccessesQuery(\n  $accessId: ID!\n) {\n  node(id: $accessId) {\n    __typename\n    ... on TrustCenterAccess {\n      id\n      availableDocumentAccesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n        edges {\n          node {\n            id\n            status\n            document {\n              id\n              title\n              documentType\n            }\n            report {\n              id\n              filename\n              audit {\n                id\n                framework {\n                  name\n                  id\n                }\n              }\n            }\n            trustCenterFile {\n              id\n              name\n              category\n            }\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "f2fad3a04a700d1a9b045d1fa09e470c";

export default node;
