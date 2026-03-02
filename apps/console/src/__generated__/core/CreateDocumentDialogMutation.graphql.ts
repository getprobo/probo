/**
 * @generated SignedSource<<a96a88e4c176f254788930b251e7108e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type CreateDocumentInput = {
  approverIds: ReadonlyArray<string>;
  classification: DocumentClassification;
  content: string;
  documentType: DocumentType;
  organizationId: string;
  title: string;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
};
export type CreateDocumentDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateDocumentInput;
};
export type CreateDocumentDialogMutation$data = {
  readonly createDocument: {
    readonly documentEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"DocumentListItemFragment">;
      };
    };
  };
};
export type CreateDocumentDialogMutation = {
  response: CreateDocumentDialogMutation$data;
  variables: CreateDocumentDialogMutation$variables;
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
  "name": "id",
  "storageKey": null
},
v4 = {
  "kind": "Literal",
  "name": "first",
  "value": 0
},
v5 = [
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "totalCount",
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "CreateDocumentDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDocumentPayload",
        "kind": "LinkedField",
        "name": "createDocument",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "DocumentEdge",
            "kind": "LinkedField",
            "name": "documentEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Document",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "DocumentListItemFragment"
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
    "name": "CreateDocumentDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDocumentPayload",
        "kind": "LinkedField",
        "name": "createDocument",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "DocumentEdge",
            "kind": "LinkedField",
            "name": "documentEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Document",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
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
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "classification",
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
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:document:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:document:delete\")"
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
                    "concreteType": "ProfileConnection",
                    "kind": "LinkedField",
                    "name": "approvers",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ProfileEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Profile",
                            "kind": "LinkedField",
                            "name": "node",
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
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "approvers(first:100)"
                  },
                  {
                    "alias": "lastVersion",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "first",
                        "value": 1
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
                              (v3/*: any*/),
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
                                "alias": null,
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "filter",
                                    "value": {
                                      "activeContract": true
                                    }
                                  },
                                  (v4/*: any*/)
                                ],
                                "concreteType": "DocumentVersionSignatureConnection",
                                "kind": "LinkedField",
                                "name": "signatures",
                                "plural": false,
                                "selections": (v5/*: any*/),
                                "storageKey": "signatures(filter:{\"activeContract\":true},first:0)"
                              },
                              {
                                "alias": "signedSignatures",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "filter",
                                    "value": {
                                      "activeContract": true,
                                      "states": [
                                        "SIGNED"
                                      ]
                                    }
                                  },
                                  (v4/*: any*/)
                                ],
                                "concreteType": "DocumentVersionSignatureConnection",
                                "kind": "LinkedField",
                                "name": "signatures",
                                "plural": false,
                                "selections": (v5/*: any*/),
                                "storageKey": "signatures(filter:{\"activeContract\":true,\"states\":[\"SIGNED\"]},first:0)"
                              }
                            ],
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "versions(first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
            "name": "documentEdge",
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
    "cacheID": "927e922143fa7eac4270e7ac0ae7da39",
    "id": null,
    "metadata": {},
    "name": "CreateDocumentDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateDocumentDialogMutation(\n  $input: CreateDocumentInput!\n) {\n  createDocument(input: $input) {\n    documentEdge {\n      node {\n        id\n        ...DocumentListItemFragment\n      }\n    }\n  }\n}\n\nfragment DocumentListItemFragment on Document {\n  id\n  title\n  documentType\n  classification\n  updatedAt\n  canDelete: permission(action: \"core:document:delete\")\n  approvers(first: 100) {\n    edges {\n      node {\n        id\n        fullName\n      }\n    }\n  }\n  lastVersion: versions(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {\n    edges {\n      node {\n        id\n        status\n        version\n        signatures(first: 0, filter: {activeContract: true}) {\n          totalCount\n        }\n        signedSignatures: signatures(first: 0, filter: {states: [SIGNED], activeContract: true}) {\n          totalCount\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "30ea1213f809faa45e873b069e27df26";

export default node;
