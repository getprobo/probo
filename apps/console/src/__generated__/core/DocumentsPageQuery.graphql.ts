/**
 * @generated SignedSource<<ea3e34a592ca995eba32df84e7f42c24>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentsPageQuery$variables = {
  organizationId: string;
};
export type DocumentsPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly canCreateDocument: boolean;
    readonly documents: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly canSendSigningNotifications: boolean;
        };
      }>;
    };
    readonly " $fragmentSpreads": FragmentRefs<"DocumentListFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type DocumentsPageQuery = {
  response: DocumentsPageQuery$data;
  variables: DocumentsPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
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
  "alias": "canCreateDocument",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:document:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:document:create\")"
},
v4 = {
  "kind": "Literal",
  "name": "first",
  "value": 50
},
v5 = {
  "direction": "ASC",
  "field": "TITLE"
},
v6 = [
  (v4/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v5/*: any*/)
  }
],
v7 = {
  "alias": "canSendSigningNotifications",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:document:send-signing-notifications"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:document:send-signing-notifications\")"
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v9 = {
  "kind": "Literal",
  "name": "first",
  "value": 0
},
v10 = [
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
    "name": "DocumentsPageQuery",
    "selections": [
      {
        "alias": "organization",
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
              {
                "args": [
                  (v4/*: any*/),
                  {
                    "kind": "Literal",
                    "name": "order",
                    "value": (v5/*: any*/)
                  }
                ],
                "kind": "FragmentSpread",
                "name": "DocumentListFragment"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "documents",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "DocumentEdge",
                    "kind": "LinkedField",
                    "name": "edges",
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
                          (v7/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "documents(first:50,orderBy:{\"direction\":\"ASC\",\"field\":\"TITLE\"})"
              }
            ],
            "type": "Organization",
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
    "name": "DocumentsPageQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v8/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": (v6/*: any*/),
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "documents",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "DocumentEdge",
                    "kind": "LinkedField",
                    "name": "edges",
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
                          (v8/*: any*/),
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
                            "alias": "canRequestSignatures",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:document-version:request-signature"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:document-version:request-signature\")"
                          },
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
                                      (v8/*: any*/),
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
                                      (v8/*: any*/),
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
                                          (v9/*: any*/)
                                        ],
                                        "concreteType": "DocumentVersionSignatureConnection",
                                        "kind": "LinkedField",
                                        "name": "signatures",
                                        "plural": false,
                                        "selections": (v10/*: any*/),
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
                                          (v9/*: any*/)
                                        ],
                                        "concreteType": "DocumentVersionSignatureConnection",
                                        "kind": "LinkedField",
                                        "name": "signatures",
                                        "plural": false,
                                        "selections": (v10/*: any*/),
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
                          },
                          (v2/*: any*/),
                          (v7/*: any*/)
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
                      }
                    ],
                    "storageKey": null
                  },
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
                  }
                ],
                "storageKey": "documents(first:50,orderBy:{\"direction\":\"ASC\",\"field\":\"TITLE\"})"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "DocumentsListQuery_documents",
                "kind": "LinkedHandle",
                "name": "documents"
              }
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "638d1d0f41916e750ecb5ac9312b0c8c",
    "id": null,
    "metadata": {},
    "name": "DocumentsPageQuery",
    "operationKind": "query",
    "text": "query DocumentsPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      canCreateDocument: permission(action: \"core:document:create\")\n      ...DocumentListFragment_1WMmEg\n      documents(first: 50, orderBy: {field: TITLE, direction: ASC}) {\n        edges {\n          node {\n            canSendSigningNotifications: permission(action: \"core:document:send-signing-notifications\")\n            id\n          }\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment DocumentListFragment_1WMmEg on Organization {\n  documents(first: 50, orderBy: {field: TITLE, direction: ASC}) {\n    edges {\n      node {\n        id\n        canUpdate: permission(action: \"core:document:update\")\n        canDelete: permission(action: \"core:document:delete\")\n        canRequestSignatures: permission(action: \"core:document-version:request-signature\")\n        ...DocumentListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment DocumentListItemFragment on Document {\n  id\n  title\n  documentType\n  classification\n  updatedAt\n  canDelete: permission(action: \"core:document:delete\")\n  approvers(first: 100) {\n    edges {\n      node {\n        id\n        fullName\n      }\n    }\n  }\n  lastVersion: versions(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {\n    edges {\n      node {\n        id\n        status\n        version\n        signatures(first: 0, filter: {activeContract: true}) {\n          totalCount\n        }\n        signedSignatures: signatures(first: 0, filter: {states: [SIGNED], activeContract: true}) {\n          totalCount\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "03658a05617ef4f971ed6b0c2758fa87";

export default node;
