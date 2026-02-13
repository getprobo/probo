/**
 * @generated SignedSource<<fa40a22c2297c683c649ccac9f6ebe50>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentLayoutQuery$variables = {
  documentId: string;
  versionId: string;
  versionSpecified: boolean;
};
export type DocumentLayoutQuery$data = {
  readonly document: {
    readonly __typename: "Document";
    readonly canPublish: boolean;
    readonly controlInfo: {
      readonly totalCount: number;
    };
    readonly id: string;
    readonly lastVersion?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly signatures: {
            readonly totalCount: number;
          };
          readonly signedSignatures: {
            readonly totalCount: number;
          };
          readonly status: DocumentStatus;
          readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdown_versionFragment" | "DocumentLayoutDrawer_versionFragment">;
        };
      }>;
    };
    readonly title: string;
    readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdown_documentFragment" | "DocumentLayoutDrawer_documentFragment" | "DocumentTitleFormFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
  readonly version?: {
    readonly __typename: "DocumentVersion";
    readonly id: string;
    readonly signatures: {
      readonly totalCount: number;
    };
    readonly signedSignatures: {
      readonly totalCount: number;
    };
    readonly status: DocumentStatus;
    readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdown_versionFragment" | "DocumentLayoutDrawer_versionFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type DocumentLayoutQuery = {
  response: DocumentLayoutQuery$data;
  variables: DocumentLayoutQuery$variables;
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
    "name": "versionId"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "versionSpecified"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "versionId"
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
  "name": "status",
  "storageKey": null
},
v5 = {
  "kind": "Literal",
  "name": "first",
  "value": 0
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v7 = [
  (v6/*: any*/)
],
v8 = {
  "alias": null,
  "args": [
    {
      "kind": "Literal",
      "name": "filter",
      "value": {
        "activeContract": true
      }
    },
    (v5/*: any*/)
  ],
  "concreteType": "DocumentVersionSignatureConnection",
  "kind": "LinkedField",
  "name": "signatures",
  "plural": false,
  "selections": (v7/*: any*/),
  "storageKey": "signatures(filter:{\"activeContract\":true},first:0)"
},
v9 = {
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
    (v5/*: any*/)
  ],
  "concreteType": "DocumentVersionSignatureConnection",
  "kind": "LinkedField",
  "name": "signatures",
  "plural": false,
  "selections": (v7/*: any*/),
  "storageKey": "signatures(filter:{\"activeContract\":true,\"states\":[\"SIGNED\"]},first:0)"
},
v10 = [
  (v3/*: any*/),
  (v4/*: any*/),
  {
    "args": null,
    "kind": "FragmentSpread",
    "name": "DocumentActionsDropdown_versionFragment"
  },
  {
    "args": null,
    "kind": "FragmentSpread",
    "name": "DocumentLayoutDrawer_versionFragment"
  },
  (v8/*: any*/),
  (v9/*: any*/)
],
v11 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "documentId"
  }
],
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "title",
  "storageKey": null
},
v13 = {
  "alias": "canPublish",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:document-version:publish"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:document-version:publish\")"
},
v14 = {
  "alias": "controlInfo",
  "args": [
    (v5/*: any*/)
  ],
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": (v7/*: any*/),
  "storageKey": "controls(first:0)"
},
v15 = [
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
v16 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 20
  }
],
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "version",
  "storageKey": null
},
v18 = {
  "alias": "canDeleteDraft",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:document-version:delete-draft"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:document-version:delete-draft\")"
},
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "classification",
  "storageKey": null
},
v20 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v21 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "publishedAt",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentLayoutQuery",
    "selections": [
      {
        "condition": "versionSpecified",
        "kind": "Condition",
        "passingValue": true,
        "selections": [
          {
            "alias": "version",
            "args": (v1/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              {
                "kind": "InlineFragment",
                "selections": (v10/*: any*/),
                "type": "DocumentVersion",
                "abstractKey": null
              }
            ],
            "storageKey": null
          }
        ]
      },
      {
        "alias": "document",
        "args": (v11/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentTitleFormFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentActionsDropdown_documentFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentLayoutDrawer_documentFragment"
              },
              {
                "condition": "versionSpecified",
                "kind": "Condition",
                "passingValue": false,
                "selections": [
                  {
                    "alias": "lastVersion",
                    "args": (v15/*: any*/),
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
                            "selections": (v10/*: any*/),
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "versions(first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentLayoutQuery",
    "selections": [
      {
        "alias": "document",
        "args": (v11/*: any*/),
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
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              {
                "alias": "canUpdate",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:document:update"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:document:update\")"
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
                "args": (v16/*: any*/),
                "concreteType": "DocumentVersionConnection",
                "kind": "LinkedField",
                "name": "versions",
                "plural": false,
                "selections": [
                  (v6/*: any*/),
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
                  },
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
                          (v4/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "content",
                            "storageKey": null
                          },
                          (v2/*: any*/)
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
                  }
                ],
                "storageKey": "versions(first:20)"
              },
              {
                "alias": null,
                "args": (v16/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "DocumentLayout_versions",
                "kind": "LinkedHandle",
                "name": "versions"
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
                "condition": "versionSpecified",
                "kind": "Condition",
                "passingValue": false,
                "selections": [
                  {
                    "alias": "lastVersion",
                    "args": (v15/*: any*/),
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
                              (v4/*: any*/),
                              (v17/*: any*/),
                              (v18/*: any*/),
                              (v19/*: any*/),
                              (v20/*: any*/),
                              (v21/*: any*/),
                              (v8/*: any*/),
                              (v9/*: any*/)
                            ],
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "versions(first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                  }
                ]
              }
            ],
            "type": "Document",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      {
        "condition": "versionSpecified",
        "kind": "Condition",
        "passingValue": true,
        "selections": [
          {
            "alias": "version",
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
                  (v4/*: any*/),
                  (v17/*: any*/),
                  (v18/*: any*/),
                  (v19/*: any*/),
                  (v20/*: any*/),
                  (v21/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/)
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
    "cacheID": "4291fe2afc2c86c77cbfd01d05f16e75",
    "id": null,
    "metadata": {},
    "name": "DocumentLayoutQuery",
    "operationKind": "query",
    "text": "query DocumentLayoutQuery(\n  $documentId: ID!\n  $versionId: ID!\n  $versionSpecified: Boolean!\n) {\n  version: node(id: $versionId) @include(if: $versionSpecified) {\n    __typename\n    ... on DocumentVersion {\n      id\n      status\n      ...DocumentActionsDropdown_versionFragment\n      ...DocumentLayoutDrawer_versionFragment\n      signatures(first: 0, filter: {activeContract: true}) {\n        totalCount\n      }\n      signedSignatures: signatures(first: 0, filter: {states: [SIGNED], activeContract: true}) {\n        totalCount\n      }\n    }\n    id\n  }\n  document: node(id: $documentId) {\n    __typename\n    ... on Document {\n      id\n      title\n      canPublish: permission(action: \"core:document-version:publish\")\n      controlInfo: controls(first: 0) {\n        totalCount\n      }\n      ...DocumentTitleFormFragment\n      ...DocumentActionsDropdown_documentFragment\n      ...DocumentLayoutDrawer_documentFragment\n      lastVersion: versions(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) @skip(if: $versionSpecified) {\n        edges {\n          node {\n            id\n            status\n            ...DocumentActionsDropdown_versionFragment\n            ...DocumentLayoutDrawer_versionFragment\n            signatures(first: 0, filter: {activeContract: true}) {\n              totalCount\n            }\n            signedSignatures: signatures(first: 0, filter: {states: [SIGNED], activeContract: true}) {\n              totalCount\n            }\n          }\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment DocumentActionsDropdown_documentFragment on Document {\n  id\n  title\n  canUpdate: permission(action: \"core:document:update\")\n  canDelete: permission(action: \"core:document:delete\")\n  versions(first: 20) {\n    totalCount\n  }\n  ...UpdateVersionDialogFragment\n}\n\nfragment DocumentActionsDropdown_versionFragment on DocumentVersion {\n  id\n  version\n  status\n  canDeleteDraft: permission(action: \"core:document-version:delete-draft\")\n}\n\nfragment DocumentLayoutDrawer_documentFragment on Document {\n  id\n  documentType\n  canUpdate: permission(action: \"core:document:update\")\n  approvers(first: 100) {\n    edges {\n      node {\n        id\n        fullName\n      }\n    }\n  }\n}\n\nfragment DocumentLayoutDrawer_versionFragment on DocumentVersion {\n  id\n  classification\n  version\n  status\n  updatedAt\n  publishedAt\n}\n\nfragment DocumentTitleFormFragment on Document {\n  id\n  title\n  canUpdate: permission(action: \"core:document:update\")\n}\n\nfragment UpdateVersionDialogFragment on Document {\n  id\n  versions(first: 20) {\n    edges {\n      node {\n        id\n        status\n        content\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2ae22db6e21db2cade3476a25b912eec";

export default node;
