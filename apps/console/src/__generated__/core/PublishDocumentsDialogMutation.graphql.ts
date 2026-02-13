/**
 * @generated SignedSource<<3d906005c4507bfced1f9d7036939157>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type BulkPublishDocumentVersionsInput = {
  changelog: string;
  documentIds: ReadonlyArray<string>;
};
export type PublishDocumentsDialogMutation$variables = {
  input: BulkPublishDocumentVersionsInput;
};
export type PublishDocumentsDialogMutation$data = {
  readonly bulkPublishDocumentVersions: {
    readonly documentEdges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"DocumentListItemFragment">;
      };
    }>;
  };
};
export type PublishDocumentsDialogMutation = {
  response: PublishDocumentsDialogMutation$data;
  variables: PublishDocumentsDialogMutation$variables;
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
  "kind": "Literal",
  "name": "first",
  "value": 0
},
v4 = [
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PublishDocumentsDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "BulkPublishDocumentVersionsPayload",
        "kind": "LinkedField",
        "name": "bulkPublishDocumentVersions",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "DocumentEdge",
            "kind": "LinkedField",
            "name": "documentEdges",
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
                  (v2/*: any*/),
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PublishDocumentsDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "BulkPublishDocumentVersionsPayload",
        "kind": "LinkedField",
        "name": "bulkPublishDocumentVersions",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "DocumentEdge",
            "kind": "LinkedField",
            "name": "documentEdges",
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
                              (v2/*: any*/),
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
                              (v2/*: any*/),
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
                                  (v3/*: any*/)
                                ],
                                "concreteType": "DocumentVersionSignatureConnection",
                                "kind": "LinkedField",
                                "name": "signatures",
                                "plural": false,
                                "selections": (v4/*: any*/),
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
                                  (v3/*: any*/)
                                ],
                                "concreteType": "DocumentVersionSignatureConnection",
                                "kind": "LinkedField",
                                "name": "signatures",
                                "plural": false,
                                "selections": (v4/*: any*/),
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
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "a62979260c1527e43c12c1effa025815",
    "id": null,
    "metadata": {},
    "name": "PublishDocumentsDialogMutation",
    "operationKind": "mutation",
    "text": "mutation PublishDocumentsDialogMutation(\n  $input: BulkPublishDocumentVersionsInput!\n) {\n  bulkPublishDocumentVersions(input: $input) {\n    documentEdges {\n      node {\n        id\n        ...DocumentListItemFragment\n      }\n    }\n  }\n}\n\nfragment DocumentListItemFragment on Document {\n  id\n  title\n  documentType\n  classification\n  updatedAt\n  canDelete: permission(action: \"core:document:delete\")\n  approvers(first: 100) {\n    edges {\n      node {\n        id\n        fullName\n      }\n    }\n  }\n  lastVersion: versions(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {\n    edges {\n      node {\n        id\n        status\n        version\n        signatures(first: 0, filter: {activeContract: true}) {\n          totalCount\n        }\n        signedSignatures: signatures(first: 0, filter: {states: [SIGNED], activeContract: true}) {\n          totalCount\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c28e9ea550947901058300bc4ba8f7f1";

export default node;
