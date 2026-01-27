/**
 * @generated SignedSource<<b3e322a9f5ab2b390b1a675d96bbfc55>>
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
          readonly content: string;
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
    readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdown_documentFragment" | "DocumentControlsTabFragment" | "DocumentLayoutDrawer_documentFragment" | "DocumentTitleFormFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
  readonly version?: {
    readonly __typename: "DocumentVersion";
    readonly content: string;
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
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "content",
  "storageKey": null
},
v6 = {
  "kind": "Literal",
  "name": "first",
  "value": 0
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v8 = [
  (v7/*: any*/)
],
v9 = {
  "alias": null,
  "args": [
    {
      "kind": "Literal",
      "name": "filter",
      "value": {
        "activeContract": true
      }
    },
    (v6/*: any*/)
  ],
  "concreteType": "DocumentVersionSignatureConnection",
  "kind": "LinkedField",
  "name": "signatures",
  "plural": false,
  "selections": (v8/*: any*/),
  "storageKey": "signatures(filter:{\"activeContract\":true},first:0)"
},
v10 = {
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
    (v6/*: any*/)
  ],
  "concreteType": "DocumentVersionSignatureConnection",
  "kind": "LinkedField",
  "name": "signatures",
  "plural": false,
  "selections": (v8/*: any*/),
  "storageKey": "signatures(filter:{\"activeContract\":true,\"states\":[\"SIGNED\"]},first:0)"
},
v11 = [
  (v3/*: any*/),
  (v4/*: any*/),
  (v5/*: any*/),
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
  (v9/*: any*/),
  (v10/*: any*/)
],
v12 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "documentId"
  }
],
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "title",
  "storageKey": null
},
v14 = {
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
v15 = {
  "alias": "controlInfo",
  "args": [
    (v6/*: any*/)
  ],
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": (v8/*: any*/),
  "storageKey": "controls(first:0)"
},
v16 = [
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
v17 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 20
  }
],
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
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v20 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v21 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v22 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v23 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "version",
  "storageKey": null
},
v24 = {
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
v25 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "classification",
  "storageKey": null
},
v26 = {
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
v27 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v28 = {
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
                "selections": (v11/*: any*/),
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
        "args": (v12/*: any*/),
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
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
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
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentControlsTabFragment"
              },
              {
                "condition": "versionSpecified",
                "kind": "Condition",
                "passingValue": false,
                "selections": [
                  {
                    "alias": "lastVersion",
                    "args": (v16/*: any*/),
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
                            "selections": (v11/*: any*/),
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
        "args": (v12/*: any*/),
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
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
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
                "args": (v17/*: any*/),
                "concreteType": "DocumentVersionConnection",
                "kind": "LinkedField",
                "name": "versions",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
                  (v18/*: any*/),
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
                          (v5/*: any*/),
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
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
                      (v20/*: any*/),
                      (v21/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "versions(first:20)"
              },
              {
                "alias": null,
                "args": (v17/*: any*/),
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
                "args": (v17/*: any*/),
                "concreteType": "ControlConnection",
                "kind": "LinkedField",
                "name": "controls",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ControlEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Control",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
                          (v22/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "sectionTitle",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Framework",
                            "kind": "LinkedField",
                            "name": "framework",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v22/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
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
                      (v20/*: any*/),
                      (v21/*: any*/),
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
                  (v18/*: any*/)
                ],
                "storageKey": "controls(first:20)"
              },
              {
                "alias": null,
                "args": (v17/*: any*/),
                "filters": [
                  "orderBy",
                  "filter"
                ],
                "handle": "connection",
                "key": "DocumentControlsTab_controls",
                "kind": "LinkedHandle",
                "name": "controls"
              },
              {
                "condition": "versionSpecified",
                "kind": "Condition",
                "passingValue": false,
                "selections": [
                  {
                    "alias": "lastVersion",
                    "args": (v16/*: any*/),
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
                              (v5/*: any*/),
                              (v23/*: any*/),
                              (v24/*: any*/),
                              (v25/*: any*/),
                              (v26/*: any*/),
                              (v27/*: any*/),
                              (v28/*: any*/),
                              (v9/*: any*/),
                              (v10/*: any*/)
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
                  (v5/*: any*/),
                  (v23/*: any*/),
                  (v24/*: any*/),
                  (v25/*: any*/),
                  (v26/*: any*/),
                  (v27/*: any*/),
                  (v28/*: any*/),
                  (v9/*: any*/),
                  (v10/*: any*/)
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
    "cacheID": "9d915bec70a2a5c1b9924d73e537d3b1",
    "id": null,
    "metadata": {},
    "name": "DocumentLayoutQuery",
    "operationKind": "query",
    "text": "query DocumentLayoutQuery(\n  $documentId: ID!\n  $versionId: ID!\n  $versionSpecified: Boolean!\n) {\n  version: node(id: $versionId) @include(if: $versionSpecified) {\n    __typename\n    ... on DocumentVersion {\n      id\n      status\n      content\n      ...DocumentActionsDropdown_versionFragment\n      ...DocumentLayoutDrawer_versionFragment\n      signatures(first: 0, filter: {activeContract: true}) {\n        totalCount\n      }\n      signedSignatures: signatures(first: 0, filter: {states: [SIGNED], activeContract: true}) {\n        totalCount\n      }\n    }\n    id\n  }\n  document: node(id: $documentId) {\n    __typename\n    ... on Document {\n      id\n      title\n      canPublish: permission(action: \"core:document-version:publish\")\n      controlInfo: controls(first: 0) {\n        totalCount\n      }\n      ...DocumentTitleFormFragment\n      ...DocumentActionsDropdown_documentFragment\n      ...DocumentLayoutDrawer_documentFragment\n      ...DocumentControlsTabFragment\n      lastVersion: versions(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) @skip(if: $versionSpecified) {\n        edges {\n          node {\n            id\n            status\n            content\n            ...DocumentActionsDropdown_versionFragment\n            ...DocumentLayoutDrawer_versionFragment\n            signatures(first: 0, filter: {activeContract: true}) {\n              totalCount\n            }\n            signedSignatures: signatures(first: 0, filter: {states: [SIGNED], activeContract: true}) {\n              totalCount\n            }\n          }\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment DocumentActionsDropdown_documentFragment on Document {\n  id\n  title\n  canUpdate: permission(action: \"core:document:update\")\n  canDelete: permission(action: \"core:document:delete\")\n  versions(first: 20) {\n    totalCount\n  }\n  ...UpdateVersionDialogFragment\n}\n\nfragment DocumentActionsDropdown_versionFragment on DocumentVersion {\n  id\n  version\n  status\n  canDeleteDraft: permission(action: \"core:document-version:delete-draft\")\n}\n\nfragment DocumentControlsTabFragment on Document {\n  id\n  controls(first: 20) {\n    edges {\n      node {\n        id\n        ...LinkedControlsCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment DocumentLayoutDrawer_documentFragment on Document {\n  id\n  documentType\n  canUpdate: permission(action: \"core:document:update\")\n}\n\nfragment DocumentLayoutDrawer_versionFragment on DocumentVersion {\n  id\n  classification\n  owner {\n    id\n    fullName\n  }\n  version\n  status\n  updatedAt\n  publishedAt\n}\n\nfragment DocumentTitleFormFragment on Document {\n  id\n  title\n  canUpdate: permission(action: \"core:document:update\")\n}\n\nfragment LinkedControlsCardFragment on Control {\n  id\n  name\n  sectionTitle\n  framework {\n    id\n    name\n  }\n}\n\nfragment UpdateVersionDialogFragment on Document {\n  id\n  versions(first: 20) {\n    edges {\n      node {\n        id\n        status\n        content\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ef2d8e2f721cb77607af7421f76cd695";

export default node;
