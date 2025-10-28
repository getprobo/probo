/**
 * @generated SignedSource<<ef22c40e1c881ddd440a0468bae615f9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustCenterAccessGraphLoadDocumentAccessesQuery$variables = {
  accessId: string;
};
export type TrustCenterAccessGraphLoadDocumentAccessesQuery$data = {
  readonly node: {
    readonly availableDocumentAccesses?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly active: boolean;
          readonly document: {
            readonly documentType: DocumentType;
            readonly id: string;
            readonly title: string;
          } | null | undefined;
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
          readonly requested: boolean;
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
  "name": "active",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "requested",
  "storageKey": null
},
v6 = {
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
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "filename",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "concreteType": "TrustCenterFile",
  "kind": "LinkedField",
  "name": "trustCenterFile",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v8/*: any*/),
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
                          (v4/*: any*/),
                          (v5/*: any*/),
                          (v6/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v7/*: any*/),
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
                                      (v8/*: any*/)
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": null
                          },
                          (v9/*: any*/)
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
                          (v4/*: any*/),
                          (v5/*: any*/),
                          (v6/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v7/*: any*/),
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
                                      (v8/*: any*/),
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
                          (v9/*: any*/)
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
    "cacheID": "50de5f19f1c13ee4d4a0b125231b989f",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphLoadDocumentAccessesQuery",
    "operationKind": "query",
    "text": "query TrustCenterAccessGraphLoadDocumentAccessesQuery(\n  $accessId: ID!\n) {\n  node(id: $accessId) {\n    __typename\n    ... on TrustCenterAccess {\n      id\n      availableDocumentAccesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n        edges {\n          node {\n            active\n            requested\n            document {\n              id\n              title\n              documentType\n            }\n            report {\n              id\n              filename\n              audit {\n                id\n                framework {\n                  name\n                  id\n                }\n              }\n            }\n            trustCenterFile {\n              id\n              name\n              category\n            }\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "64c69558e335d4cecb14192d8c5b66b8";

export default node;
