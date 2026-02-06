/**
 * @generated SignedSource<<e9fedeb822edc0b01bc54ec76a816761>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturesPageQuery$variables = {
  documentId: string;
  organizationId: string;
  versionId: string;
  versionSpecified: boolean;
};
export type DocumentSignaturesPageQuery$data = {
  readonly document?: {
    readonly __typename: "Document";
    readonly lastVersion: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureList_versionFragment">;
        };
      }>;
    };
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
  readonly organization: {
    readonly __typename: string;
    readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureList_peopleFragment">;
  };
  readonly version?: {
    readonly __typename: string;
    readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureList_versionFragment">;
  };
};
export type DocumentSignaturesPageQuery = {
  response: DocumentSignaturesPageQuery$data;
  variables: DocumentSignaturesPageQuery$variables;
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
    "name": "organizationId"
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
  "kind": "Literal",
  "name": "filter",
  "value": {
    "excludeContractEnded": true
  }
},
v4 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "documentId"
  }
],
v5 = [
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
v6 = {
  "args": null,
  "kind": "FragmentSpread",
  "name": "DocumentSignatureList_versionFragment"
},
v7 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "versionId"
  }
],
v8 = {
  "kind": "Literal",
  "name": "first",
  "value": 1000
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v12 = [
  (v8/*: any*/)
],
v13 = {
  "alias": null,
  "args": (v12/*: any*/),
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
            (v9/*: any*/),
            {
              "alias": null,
              "args": null,
              "concreteType": "Profile",
              "kind": "LinkedField",
              "name": "signedBy",
              "plural": false,
              "selections": [
                (v9/*: any*/),
                (v10/*: any*/)
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
  "storageKey": "signatures(first:1000)"
},
v14 = {
  "alias": null,
  "args": (v12/*: any*/),
  "filters": [
    "filter"
  ],
  "handle": "connection",
  "key": "DocumentSignaturesTab_signatures",
  "kind": "LinkedHandle",
  "name": "signatures"
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentSignaturesPageQuery",
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
            "args": [
              (v3/*: any*/)
            ],
            "kind": "FragmentSpread",
            "name": "DocumentSignatureList_peopleFragment"
          }
        ],
        "storageKey": null
      },
      {
        "condition": "versionSpecified",
        "kind": "Condition",
        "passingValue": false,
        "selections": [
          {
            "alias": "document",
            "args": (v4/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              {
                "kind": "InlineFragment",
                "selections": [
                  {
                    "alias": "lastVersion",
                    "args": (v5/*: any*/),
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
                              (v6/*: any*/)
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
                "type": "Document",
                "abstractKey": null
              }
            ],
            "storageKey": null
          }
        ]
      },
      {
        "condition": "versionSpecified",
        "kind": "Condition",
        "passingValue": true,
        "selections": [
          {
            "alias": "version",
            "args": (v7/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v6/*: any*/)
            ],
            "storageKey": null
          }
        ]
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentSignaturesPageQuery",
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
              {
                "alias": "canRequestSignature",
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
                "args": [
                  (v3/*: any*/),
                  (v8/*: any*/),
                  {
                    "kind": "Literal",
                    "name": "orderBy",
                    "value": {
                      "direction": "ASC",
                      "field": "FULL_NAME"
                    }
                  }
                ],
                "concreteType": "ProfileConnection",
                "kind": "LinkedField",
                "name": "profiles",
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
                          (v9/*: any*/),
                          (v10/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "emailAddress",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "profiles(filter:{\"excludeContractEnded\":true},first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"FULL_NAME\"})"
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v9/*: any*/)
        ],
        "storageKey": null
      },
      {
        "condition": "versionSpecified",
        "kind": "Condition",
        "passingValue": false,
        "selections": [
          {
            "alias": "document",
            "args": (v4/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              {
                "kind": "InlineFragment",
                "selections": [
                  {
                    "alias": "lastVersion",
                    "args": (v5/*: any*/),
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
                              (v9/*: any*/),
                              (v11/*: any*/),
                              (v13/*: any*/),
                              (v14/*: any*/)
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
                "type": "Document",
                "abstractKey": null
              },
              (v9/*: any*/)
            ],
            "storageKey": null
          }
        ]
      },
      {
        "condition": "versionSpecified",
        "kind": "Condition",
        "passingValue": true,
        "selections": [
          {
            "alias": "version",
            "args": (v7/*: any*/),
            "concreteType": null,
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v9/*: any*/),
              {
                "kind": "InlineFragment",
                "selections": [
                  (v11/*: any*/),
                  (v13/*: any*/),
                  (v14/*: any*/)
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
    "cacheID": "7aa234078557c0ee417530dcbebba110",
    "id": null,
    "metadata": {},
    "name": "DocumentSignaturesPageQuery",
    "operationKind": "query",
    "text": "query DocumentSignaturesPageQuery(\n  $documentId: ID!\n  $organizationId: ID!\n  $versionId: ID!\n  $versionSpecified: Boolean!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ...DocumentSignatureList_peopleFragment_4oWXMW\n    id\n  }\n  document: node(id: $documentId) @skip(if: $versionSpecified) {\n    __typename\n    ... on Document {\n      lastVersion: versions(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {\n        edges {\n          node {\n            ...DocumentSignatureList_versionFragment\n            id\n          }\n        }\n      }\n    }\n    id\n  }\n  version: node(id: $versionId) @include(if: $versionSpecified) {\n    __typename\n    ...DocumentSignatureList_versionFragment\n    id\n  }\n}\n\nfragment DocumentSignatureListItemFragment on DocumentVersionSignature {\n  id\n  signedBy {\n    fullName\n    id\n  }\n  state\n  signedAt\n  requestedAt\n  canCancel: permission(action: \"core:document-version-signature:cancel\")\n}\n\nfragment DocumentSignatureList_peopleFragment_4oWXMW on Organization {\n  ...DocumentSignaturePlaceholder_organizationFragment\n  profiles(first: 1000, orderBy: {direction: ASC, field: FULL_NAME}, filter: {excludeContractEnded: true}) {\n    edges {\n      node {\n        id\n        ...DocumentSignaturePlaceholder_personFragment\n      }\n    }\n  }\n}\n\nfragment DocumentSignatureList_versionFragment on DocumentVersion {\n  ...DocumentSignaturePlaceholder_versionFragment\n  signatures(first: 1000) {\n    edges {\n      node {\n        id\n        signedBy {\n          id\n        }\n        ...DocumentSignatureListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  id\n}\n\nfragment DocumentSignaturePlaceholder_organizationFragment on Organization {\n  canRequestSignature: permission(action: \"core:document-version:request-signature\")\n}\n\nfragment DocumentSignaturePlaceholder_personFragment on Profile {\n  id\n  fullName\n  emailAddress\n}\n\nfragment DocumentSignaturePlaceholder_versionFragment on DocumentVersion {\n  id\n  status\n}\n"
  }
};
})();

(node as any).hash = "2468bd3b1b46b19f6f7a35e9b69beab6";

export default node;
