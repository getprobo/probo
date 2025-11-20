/**
 * @generated SignedSource<<6caec1429cd03b026975c0e9ef7c76f6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentOrderField = "CREATED_AT" | "DOCUMENT_TYPE" | "TITLE";
export type OrderDirection = "ASC" | "DESC";
export type DocumentOrder = {
  direction: OrderDirection;
  field: DocumentOrderField;
};
export type EmployeeDocumentsListQuery$variables = {
  after?: any | null | undefined;
  before?: any | null | undefined;
  first?: number | null | undefined;
  last?: number | null | undefined;
  order?: DocumentOrder | null | undefined;
  organizationId: string;
};
export type EmployeeDocumentsListQuery$data = {
  readonly viewer: {
    readonly " $fragmentSpreads": FragmentRefs<"EmployeeDocumentsPageListFragment">;
  };
};
export type EmployeeDocumentsListQuery = {
  response: EmployeeDocumentsListQuery$data;
  variables: EmployeeDocumentsListQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "after"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "before"
  },
  {
    "defaultValue": 50,
    "kind": "LocalArgument",
    "name": "first"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "last"
  },
  {
    "defaultValue": {
      "direction": "DESC",
      "field": "CREATED_AT"
    },
    "kind": "LocalArgument",
    "name": "order"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = {
  "kind": "Variable",
  "name": "after",
  "variableName": "after"
},
v2 = {
  "kind": "Variable",
  "name": "before",
  "variableName": "before"
},
v3 = {
  "kind": "Variable",
  "name": "first",
  "variableName": "first"
},
v4 = {
  "kind": "Variable",
  "name": "last",
  "variableName": "last"
},
v5 = {
  "kind": "Variable",
  "name": "organizationId",
  "variableName": "organizationId"
},
v6 = [
  (v1/*: any*/),
  (v2/*: any*/),
  (v3/*: any*/),
  (v4/*: any*/),
  {
    "kind": "Variable",
    "name": "orderBy",
    "variableName": "order"
  },
  (v5/*: any*/)
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EmployeeDocumentsListQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "args": [
              (v1/*: any*/),
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              {
                "kind": "Variable",
                "name": "order",
                "variableName": "order"
              },
              (v5/*: any*/)
            ],
            "kind": "FragmentSpread",
            "name": "EmployeeDocumentsPageListFragment"
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
    "name": "EmployeeDocumentsListQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": (v6/*: any*/),
            "concreteType": "SignableDocumentConnection",
            "kind": "LinkedField",
            "name": "signableDocuments",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "SignableDocumentEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "SignableDocument",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v7/*: any*/),
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
                        "name": "signed",
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
                        "args": null,
                        "kind": "ScalarField",
                        "name": "__typename",
                        "storageKey": null
                      }
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
            "storageKey": null
          },
          {
            "alias": null,
            "args": (v6/*: any*/),
            "filters": [
              "organizationId",
              "orderBy"
            ],
            "handle": "connection",
            "key": "EmployeeDocumentsListQuery_signableDocuments",
            "kind": "LinkedHandle",
            "name": "signableDocuments"
          },
          (v7/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "32e793fba0e3d2e46857cd1b8c436ba2",
    "id": null,
    "metadata": {},
    "name": "EmployeeDocumentsListQuery",
    "operationKind": "query",
    "text": "query EmployeeDocumentsListQuery(\n  $after: CursorKey = null\n  $before: CursorKey = null\n  $first: Int = 50\n  $last: Int = null\n  $order: DocumentOrder = {field: CREATED_AT, direction: DESC}\n  $organizationId: ID!\n) {\n  viewer {\n    ...EmployeeDocumentsPageListFragment_KjvVI\n    id\n  }\n}\n\nfragment EmployeeDocumentsPageListFragment_KjvVI on Viewer {\n  signableDocuments(organizationId: $organizationId, first: $first, after: $after, last: $last, before: $before, orderBy: $order) {\n    edges {\n      node {\n        id\n        ...EmployeeDocumentsPageRowFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment EmployeeDocumentsPageRowFragment on SignableDocument {\n  id\n  title\n  documentType\n  classification\n  signed\n  updatedAt\n}\n"
  }
};
})();

(node as any).hash = "ebd2703f79cdf6900b5e42fc3b28932a";

export default node;
