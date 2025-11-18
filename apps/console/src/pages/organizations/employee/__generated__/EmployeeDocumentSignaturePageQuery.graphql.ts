/**
 * @generated SignedSource<<9e8c0459987993bb4b217618a46fdbf5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type EmployeeDocumentSignaturePageQuery$variables = {
  documentId: string;
};
export type EmployeeDocumentSignaturePageQuery$data = {
  readonly viewer: {
    readonly id: string;
    readonly signableDocument: {
      readonly id: string;
      readonly " $fragmentSpreads": FragmentRefs<"EmployeeDocumentSignaturePageDocumentFragment">;
    } | null | undefined;
  };
};
export type EmployeeDocumentSignaturePageQuery = {
  response: EmployeeDocumentSignaturePageQuery$data;
  variables: EmployeeDocumentSignaturePageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "documentId"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "documentId"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "signed",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EmployeeDocumentSignaturePageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": (v2/*: any*/),
            "concreteType": "SignableDocument",
            "kind": "LinkedField",
            "name": "signableDocument",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "EmployeeDocumentSignaturePageDocumentFragment"
              }
            ],
            "storageKey": null
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
    "name": "EmployeeDocumentSignaturePageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": (v2/*: any*/),
            "concreteType": "SignableDocument",
            "kind": "LinkedField",
            "name": "signableDocument",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "title",
                "storageKey": null
              },
              (v3/*: any*/),
              {
                "alias": null,
                "args": [
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
                          (v1/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "version",
                            "storageKey": null
                          },
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "publishedAt",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "versions(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
    "cacheID": "a4866db288e1bf46e6562aa06bff4231",
    "id": null,
    "metadata": {},
    "name": "EmployeeDocumentSignaturePageQuery",
    "operationKind": "query",
    "text": "query EmployeeDocumentSignaturePageQuery(\n  $documentId: ID!\n) {\n  viewer {\n    id\n    signableDocument(id: $documentId) {\n      id\n      ...EmployeeDocumentSignaturePageDocumentFragment\n    }\n  }\n}\n\nfragment EmployeeDocumentSignaturePageDocumentFragment on SignableDocument {\n  id\n  title\n  signed\n  versions(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n    edges {\n      node {\n        id\n        ...EmployeeDocumentSignaturePageVersionFragment\n      }\n    }\n  }\n}\n\nfragment EmployeeDocumentSignaturePageVersionFragment on DocumentVersion {\n  id\n  version\n  signed\n  publishedAt\n}\n"
  }
};
})();

(node as any).hash = "a2ac190853e8f078ff90213605a66e29";

export default node;
