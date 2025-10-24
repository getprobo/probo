/**
 * @generated SignedSource<<8ecf71138aa0165738506f5e589d3181>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type UpdateTrustCenterAccessInput = {
  active?: boolean | null | undefined;
  documentIds?: ReadonlyArray<string> | null | undefined;
  id: string;
  name?: string | null | undefined;
  reportIds?: ReadonlyArray<string> | null | undefined;
  trustCenterFileIds?: ReadonlyArray<string> | null | undefined;
};
export type TrustCenterAccessGraphUpdateMutation$variables = {
  input: UpdateTrustCenterAccessInput;
};
export type TrustCenterAccessGraphUpdateMutation$data = {
  readonly updateTrustCenterAccess: {
    readonly trustCenterAccess: {
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
      readonly updatedAt: any;
    };
  };
};
export type TrustCenterAccessGraphUpdateMutation = {
  response: TrustCenterAccessGraphUpdateMutation$data;
  variables: TrustCenterAccessGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
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
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "active",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasAcceptedNonDisclosureAgreement",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v9 = [
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
v10 = {
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
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "filename",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "concreteType": "TrustCenterFile",
  "kind": "LinkedField",
  "name": "trustCenterFile",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v4/*: any*/),
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
    "name": "TrustCenterAccessGraphUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "updateTrustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterAccess",
            "kind": "LinkedField",
            "name": "trustCenterAccess",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              {
                "alias": null,
                "args": (v9/*: any*/),
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
                          (v5/*: any*/),
                          (v7/*: any*/),
                          (v8/*: any*/),
                          (v10/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v11/*: any*/),
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
                          (v12/*: any*/)
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
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterAccessGraphUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "updateTrustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterAccess",
            "kind": "LinkedField",
            "name": "trustCenterAccess",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              {
                "alias": null,
                "args": (v9/*: any*/),
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
                          (v5/*: any*/),
                          (v7/*: any*/),
                          (v8/*: any*/),
                          (v10/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v11/*: any*/),
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
                                      (v4/*: any*/),
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
                          (v12/*: any*/)
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
    ]
  },
  "params": {
    "cacheID": "b66da9236d9f81563608b3aa5f730908",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphUpdateMutation(\n  $input: UpdateTrustCenterAccessInput!\n) {\n  updateTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      email\n      name\n      active\n      hasAcceptedNonDisclosureAgreement\n      createdAt\n      updatedAt\n      documentAccesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n        edges {\n          node {\n            id\n            active\n            createdAt\n            updatedAt\n            document {\n              id\n              title\n              documentType\n            }\n            report {\n              id\n              filename\n              audit {\n                id\n                framework {\n                  name\n                  id\n                }\n              }\n            }\n            trustCenterFile {\n              id\n              name\n              category\n            }\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "43e73cc9d80e898f1ba2a92eaf4646f5";

export default node;
