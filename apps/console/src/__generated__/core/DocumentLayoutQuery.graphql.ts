/**
 * @generated SignedSource<<2d8a714fdbea520c2365e64524b68da5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
export type DocumentLayoutQuery$variables = {
  documentId: string;
};
export type DocumentLayoutQuery$data = {
  readonly document: {
    readonly __typename: "Document";
    readonly canPublish: boolean;
    readonly classification: DocumentClassification;
    readonly controlInfo: {
      readonly totalCount: number;
    };
    readonly id: string;
    readonly owner: {
      readonly fullName: string;
      readonly id: string;
    };
    readonly title: string;
    readonly versions: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly canDeleteDraft: boolean;
          readonly classification: DocumentClassification;
          readonly content: string;
          readonly id: string;
          readonly owner: {
            readonly fullName: string;
            readonly id: string;
          };
          readonly publishedAt: string | null | undefined;
          readonly signatures: {
            readonly __id: string;
            readonly edges: ReadonlyArray<{
              readonly node: {
                readonly id: string;
                readonly signedBy: {
                  readonly id: string;
                };
                readonly state: DocumentVersionSignatureState;
                readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesTab_signature">;
              };
            }>;
          };
          readonly status: DocumentStatus;
          readonly updatedAt: string;
          readonly version: number;
          readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesTab_version">;
        };
      }>;
    };
    readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdownFragment" | "DocumentControlsTabFragment" | "DocumentLayoutDrawerFragment" | "DocumentTitleFormFragment">;
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
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "documentId"
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
  "name": "title",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "classification",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v3/*: any*/),
    (v6/*: any*/)
  ],
  "storageKey": null
},
v8 = {
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
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v10 = {
  "alias": "controlInfo",
  "args": [
    {
      "kind": "Literal",
      "name": "first",
      "value": 0
    }
  ],
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": [
    (v9/*: any*/)
  ],
  "storageKey": "controls(first:0)"
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "content",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "publishedAt",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "version",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v16 = {
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
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "state",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v20 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v21 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    (v19/*: any*/),
    (v20/*: any*/)
  ],
  "storageKey": null
},
v22 = {
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
v23 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 20
  }
],
v24 = [
  {
    "kind": "Literal",
    "name": "action",
    "value": "core:document-version:request-signature"
  }
],
v25 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  }
],
v26 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
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
        "alias": "document",
        "args": (v1/*: any*/),
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
              (v4/*: any*/),
              (v5/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v10/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentTitleFormFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentActionsDropdownFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentLayoutDrawerFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "DocumentControlsTabFragment"
              },
              {
                "alias": "versions",
                "args": null,
                "concreteType": "DocumentVersionConnection",
                "kind": "LinkedField",
                "name": "__DocumentLayout_versions_connection",
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
                          (v11/*: any*/),
                          (v12/*: any*/),
                          (v13/*: any*/),
                          (v14/*: any*/),
                          (v15/*: any*/),
                          (v5/*: any*/),
                          (v7/*: any*/),
                          (v16/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "DocumentSignaturesTab_version"
                          },
                          {
                            "alias": "signatures",
                            "args": null,
                            "concreteType": "DocumentVersionSignatureConnection",
                            "kind": "LinkedField",
                            "name": "__DocumentDetailPage_signatures_connection",
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
                                      (v3/*: any*/),
                                      (v17/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "concreteType": "People",
                                        "kind": "LinkedField",
                                        "name": "signedBy",
                                        "plural": false,
                                        "selections": [
                                          (v3/*: any*/)
                                        ],
                                        "storageKey": null
                                      },
                                      {
                                        "args": null,
                                        "kind": "FragmentSpread",
                                        "name": "DocumentSignaturesTab_signature"
                                      },
                                      (v2/*: any*/)
                                    ],
                                    "storageKey": null
                                  },
                                  (v18/*: any*/)
                                ],
                                "storageKey": null
                              },
                              (v21/*: any*/),
                              (v22/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v21/*: any*/),
                  (v22/*: any*/)
                ],
                "storageKey": null
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
              (v7/*: any*/),
              (v8/*: any*/),
              (v10/*: any*/),
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
                "args": (v23/*: any*/),
                "concreteType": "DocumentVersionConnection",
                "kind": "LinkedField",
                "name": "versions",
                "plural": false,
                "selections": [
                  (v9/*: any*/),
                  (v22/*: any*/),
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
                          (v12/*: any*/),
                          (v11/*: any*/),
                          (v2/*: any*/),
                          (v13/*: any*/),
                          (v14/*: any*/),
                          (v15/*: any*/),
                          (v5/*: any*/),
                          (v7/*: any*/),
                          (v16/*: any*/),
                          {
                            "alias": "canRequestSignature",
                            "args": (v24/*: any*/),
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:document-version:request-signature\")"
                          },
                          {
                            "alias": null,
                            "args": (v25/*: any*/),
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
                                      (v3/*: any*/),
                                      (v17/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "concreteType": "People",
                                        "kind": "LinkedField",
                                        "name": "signedBy",
                                        "plural": false,
                                        "selections": [
                                          (v3/*: any*/),
                                          (v6/*: any*/),
                                          {
                                            "alias": null,
                                            "args": null,
                                            "kind": "ScalarField",
                                            "name": "primaryEmailAddress",
                                            "storageKey": null
                                          }
                                        ],
                                        "storageKey": null
                                      },
                                      {
                                        "alias": null,
                                        "args": null,
                                        "kind": "ScalarField",
                                        "name": "signedAt",
                                        "storageKey": null
                                      },
                                      {
                                        "alias": null,
                                        "args": null,
                                        "kind": "ScalarField",
                                        "name": "requestedAt",
                                        "storageKey": null
                                      },
                                      {
                                        "alias": "canCancel",
                                        "args": (v24/*: any*/),
                                        "kind": "ScalarField",
                                        "name": "permission",
                                        "storageKey": "permission(action:\"core:document-version:request-signature\")"
                                      },
                                      (v2/*: any*/)
                                    ],
                                    "storageKey": null
                                  },
                                  (v18/*: any*/)
                                ],
                                "storageKey": null
                              },
                              (v21/*: any*/),
                              (v22/*: any*/)
                            ],
                            "storageKey": "signatures(first:1000)"
                          },
                          {
                            "alias": null,
                            "args": (v25/*: any*/),
                            "filters": [
                              "filter"
                            ],
                            "handle": "connection",
                            "key": "DocumentSignaturesTab_signatures",
                            "kind": "LinkedHandle",
                            "name": "signatures"
                          },
                          {
                            "alias": null,
                            "args": (v25/*: any*/),
                            "filters": [],
                            "handle": "connection",
                            "key": "DocumentDetailPage_signatures",
                            "kind": "LinkedHandle",
                            "name": "signatures"
                          }
                        ],
                        "storageKey": null
                      },
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v21/*: any*/)
                ],
                "storageKey": "versions(first:20)"
              },
              {
                "alias": null,
                "args": (v23/*: any*/),
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
                "args": (v23/*: any*/),
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
                          (v26/*: any*/),
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
                              (v26/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v18/*: any*/)
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
                      (v19/*: any*/),
                      (v20/*: any*/),
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
                  (v22/*: any*/)
                ],
                "storageKey": "controls(first:20)"
              },
              {
                "alias": null,
                "args": (v23/*: any*/),
                "filters": [
                  "orderBy",
                  "filter"
                ],
                "handle": "connection",
                "key": "DocumentControlsTab_controls",
                "kind": "LinkedHandle",
                "name": "controls"
              }
            ],
            "type": "Document",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "fdb52683d7fce578839cfff215840057",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": null
        },
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "document",
            "versions"
          ]
        }
      ]
    },
    "name": "DocumentLayoutQuery",
    "operationKind": "query",
    "text": "query DocumentLayoutQuery(\n  $documentId: ID!\n) {\n  document: node(id: $documentId) {\n    __typename\n    ... on Document {\n      id\n      title\n      classification\n      owner {\n        id\n        fullName\n      }\n      canPublish: permission(action: \"core:document-version:publish\")\n      controlInfo: controls(first: 0) {\n        totalCount\n      }\n      ...DocumentTitleFormFragment\n      ...DocumentActionsDropdownFragment\n      ...DocumentLayoutDrawerFragment\n      ...DocumentControlsTabFragment\n      versions(first: 20) {\n        edges {\n          node {\n            id\n            content\n            status\n            publishedAt\n            version\n            updatedAt\n            classification\n            owner {\n              id\n              fullName\n            }\n            canDeleteDraft: permission(action: \"core:document-version:delete-draft\")\n            ...DocumentSignaturesTab_version\n            signatures(first: 1000) {\n              edges {\n                node {\n                  id\n                  state\n                  signedBy {\n                    id\n                  }\n                  ...DocumentSignaturesTab_signature\n                  __typename\n                }\n                cursor\n              }\n              pageInfo {\n                endCursor\n                hasNextPage\n              }\n            }\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment DocumentActionsDropdownFragment on Document {\n  id\n  title\n  canUpdate: permission(action: \"core:document:update\")\n  canDelete: permission(action: \"core:document:delete\")\n  versions(first: 20) {\n    totalCount\n  }\n  ...UpdateVersionDialogFragment\n}\n\nfragment DocumentControlsTabFragment on Document {\n  id\n  controls(first: 20) {\n    edges {\n      node {\n        id\n        ...LinkedControlsCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment DocumentLayoutDrawerFragment on Document {\n  id\n  documentType\n  canUpdate: permission(action: \"core:document:update\")\n}\n\nfragment DocumentSignaturesTab_signature on DocumentVersionSignature {\n  id\n  state\n  signedAt\n  requestedAt\n  signedBy {\n    fullName\n    primaryEmailAddress\n    id\n  }\n  canCancel: permission(action: \"core:document-version:request-signature\")\n}\n\nfragment DocumentSignaturesTab_version on DocumentVersion {\n  id\n  status\n  canRequestSignature: permission(action: \"core:document-version:request-signature\")\n  signatures(first: 1000) {\n    edges {\n      node {\n        id\n        state\n        signedBy {\n          id\n          fullName\n          primaryEmailAddress\n        }\n        ...DocumentSignaturesTab_signature\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment DocumentTitleFormFragment on Document {\n  id\n  title\n  canUpdate: permission(action: \"core:document:update\")\n}\n\nfragment LinkedControlsCardFragment on Control {\n  id\n  name\n  sectionTitle\n  framework {\n    id\n    name\n  }\n}\n\nfragment UpdateVersionDialogFragment on Document {\n  id\n  versions(first: 20) {\n    edges {\n      node {\n        id\n        status\n        content\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "eb6bf3db2691763a5d1f4f2435f986e6";

export default node;
