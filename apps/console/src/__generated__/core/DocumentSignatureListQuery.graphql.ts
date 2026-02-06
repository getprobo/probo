/**
 * @generated SignedSource<<c8559b76173b328f86973840e35312a5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
export type DocumentVersionSignatureFilter = {
  activeContract?: boolean | null | undefined;
  states?: ReadonlyArray<DocumentVersionSignatureState> | null | undefined;
};
export type DocumentSignatureListQuery$variables = {
  count?: number | null | undefined;
  cursor?: string | null | undefined;
  id: string;
  signatureFilter?: DocumentVersionSignatureFilter | null | undefined;
};
export type DocumentSignatureListQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureList_versionFragment">;
  };
};
export type DocumentSignatureListQuery = {
  response: DocumentSignatureListQuery$data;
  variables: DocumentSignatureListQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": 1000,
  "kind": "LocalArgument",
  "name": "count"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "cursor"
},
v2 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "id"
},
v3 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "signatureFilter"
},
v4 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
],
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v7 = [
  {
    "kind": "Variable",
    "name": "after",
    "variableName": "cursor"
  },
  {
    "kind": "Variable",
    "name": "filter",
    "variableName": "signatureFilter"
  },
  {
    "kind": "Variable",
    "name": "first",
    "variableName": "count"
  }
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v2/*: any*/),
      (v3/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentSignatureListQuery",
    "selections": [
      {
        "alias": null,
        "args": (v4/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "args": [
              {
                "kind": "Variable",
                "name": "count",
                "variableName": "count"
              },
              {
                "kind": "Variable",
                "name": "cursor",
                "variableName": "cursor"
              },
              {
                "kind": "Variable",
                "name": "signatureFilter",
                "variableName": "signatureFilter"
              }
            ],
            "kind": "FragmentSpread",
            "name": "DocumentSignatureList_versionFragment"
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
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v3/*: any*/),
      (v2/*: any*/)
    ],
    "kind": "Operation",
    "name": "DocumentSignatureListQuery",
    "selections": [
      {
        "alias": null,
        "args": (v4/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          (v6/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "status",
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v7/*: any*/),
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
                          (v6/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Profile",
                            "kind": "LinkedField",
                            "name": "signedBy",
                            "plural": false,
                            "selections": [
                              (v6/*: any*/),
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
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "state",
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
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:document-version-signature:cancel"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:document-version-signature:cancel\")"
                          },
                          (v5/*: any*/)
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
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v7/*: any*/),
                "filters": [
                  "filter"
                ],
                "handle": "connection",
                "key": "DocumentSignaturesTab_signatures",
                "kind": "LinkedHandle",
                "name": "signatures"
              }
            ],
            "type": "DocumentVersion",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "25f679c343e22278c8a3d5e2db9bb3d0",
    "id": null,
    "metadata": {},
    "name": "DocumentSignatureListQuery",
    "operationKind": "query",
    "text": "query DocumentSignatureListQuery(\n  $count: Int = 1000\n  $cursor: CursorKey\n  $signatureFilter: DocumentVersionSignatureFilter\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...DocumentSignatureList_versionFragment_1vp7QE\n    id\n  }\n}\n\nfragment DocumentSignatureListItemFragment on DocumentVersionSignature {\n  id\n  signedBy {\n    fullName\n    id\n  }\n  state\n  signedAt\n  requestedAt\n  canCancel: permission(action: \"core:document-version-signature:cancel\")\n}\n\nfragment DocumentSignatureList_versionFragment_1vp7QE on DocumentVersion {\n  ...DocumentSignaturePlaceholder_versionFragment\n  signatures(first: $count, after: $cursor, filter: $signatureFilter) {\n    edges {\n      node {\n        id\n        signedBy {\n          id\n        }\n        ...DocumentSignatureListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  id\n}\n\nfragment DocumentSignaturePlaceholder_versionFragment on DocumentVersion {\n  id\n  status\n}\n"
  }
};
})();

(node as any).hash = "cc402c3b46eacfb04db1186bbd744f67";

export default node;
