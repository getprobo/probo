/**
 * @generated SignedSource<<2cac1fe037796fc1b9542c9388ed930d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery$variables = {
  documentId: string;
};
export type OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery$data = {
  readonly document: {
    readonly id: string;
    readonly title?: string;
  };
};
export type OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery = {
  response: OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery$data;
  variables: OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery$variables;
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
  "name": "id",
  "storageKey": null
},
v3 = {
  "kind": "InlineFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "title",
      "storageKey": null
    }
  ],
  "type": "Document",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery",
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
          (v3/*: any*/)
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
    "name": "OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery",
    "selections": [
      {
        "alias": "document",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "bf577e8122cd59894e36f951f7a3551a",
    "id": null,
    "metadata": {},
    "name": "OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery",
    "operationKind": "query",
    "text": "query OrganizationBreadcrumbBreadcrumbDocumentOverviewQuery(\n  $documentId: ID!\n) {\n  document: node(id: $documentId) {\n    __typename\n    id\n    ... on Document {\n      title\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7ef8052c69e997cf2252d7c00b3f1c1f";

export default node;
