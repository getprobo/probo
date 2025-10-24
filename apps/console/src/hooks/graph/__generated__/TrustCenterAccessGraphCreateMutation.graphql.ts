/**
 * @generated SignedSource<<d527603a814e765a1a2be0900d1c59e1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type CreateTrustCenterAccessInput = {
  active: boolean;
  email: string;
  name: string;
  trustCenterId: string;
};
export type TrustCenterAccessGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateTrustCenterAccessInput;
};
export type TrustCenterAccessGraphCreateMutation$data = {
  readonly createTrustCenterAccess: {
    readonly trustCenterAccessEdge: {
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
              readonly trustCenterFile: {
                readonly category: string;
                readonly id: string;
                readonly name: string;
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
    };
  };
};
export type TrustCenterAccessGraphCreateMutation = {
  response: TrustCenterAccessGraphCreateMutation$data;
  variables: TrustCenterAccessGraphCreateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "active",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasAcceptedNonDisclosureAgreement",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v10 = [
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
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "concreteType": "Document",
  "kind": "LinkedField",
  "name": "document",
  "plural": false,
  "selections": [
    (v4/*: any*/),
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
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "filename",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "concreteType": "TrustCenterFile",
  "kind": "LinkedField",
  "name": "trustCenterFile",
  "plural": false,
  "selections": [
    (v4/*: any*/),
    (v6/*: any*/),
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
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterAccessGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "createTrustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterAccessEdge",
            "kind": "LinkedField",
            "name": "trustCenterAccessEdge",
            "plural": false,
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterAccess",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  {
                    "alias": null,
                    "args": (v10/*: any*/),
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
                              (v4/*: any*/),
                              (v7/*: any*/),
                              (v9/*: any*/),
                              (v11/*: any*/),
                              (v12/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "Report",
                                "kind": "LinkedField",
                                "name": "report",
                                "plural": false,
                                "selections": [
                                  (v4/*: any*/),
                                  (v13/*: any*/),
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "Audit",
                                    "kind": "LinkedField",
                                    "name": "audit",
                                    "plural": false,
                                    "selections": [
                                      (v4/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "concreteType": "Framework",
                                        "kind": "LinkedField",
                                        "name": "framework",
                                        "plural": false,
                                        "selections": [
                                          (v6/*: any*/)
                                        ],
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              },
                              (v14/*: any*/)
                            ],
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "documentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "TrustCenterAccessGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "createTrustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterAccessEdge",
            "kind": "LinkedField",
            "name": "trustCenterAccessEdge",
            "plural": false,
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterAccess",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  {
                    "alias": null,
                    "args": (v10/*: any*/),
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
                              (v4/*: any*/),
                              (v7/*: any*/),
                              (v9/*: any*/),
                              (v11/*: any*/),
                              (v12/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "Report",
                                "kind": "LinkedField",
                                "name": "report",
                                "plural": false,
                                "selections": [
                                  (v4/*: any*/),
                                  (v13/*: any*/),
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "Audit",
                                    "kind": "LinkedField",
                                    "name": "audit",
                                    "plural": false,
                                    "selections": [
                                      (v4/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "concreteType": "Framework",
                                        "kind": "LinkedField",
                                        "name": "framework",
                                        "plural": false,
                                        "selections": [
                                          (v6/*: any*/),
                                          (v4/*: any*/)
                                        ],
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              },
                              (v14/*: any*/)
                            ],
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "documentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "trustCenterAccessEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "4422a02632b77f81f3b5048b5acbac98",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphCreateMutation(\n  $input: CreateTrustCenterAccessInput!\n) {\n  createTrustCenterAccess(input: $input) {\n    trustCenterAccessEdge {\n      cursor\n      node {\n        id\n        email\n        name\n        active\n        hasAcceptedNonDisclosureAgreement\n        createdAt\n        documentAccesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n          edges {\n            node {\n              id\n              active\n              createdAt\n              updatedAt\n              document {\n                id\n                title\n                documentType\n              }\n              report {\n                id\n                filename\n                audit {\n                  id\n                  framework {\n                    name\n                    id\n                  }\n                }\n              }\n              trustCenterFile {\n                id\n                name\n                category\n              }\n            }\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c09c8a22a45226949a08f2262ddd0b1c";

export default node;
