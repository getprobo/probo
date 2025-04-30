/**
 * @generated SignedSource<<e85f30bc133288dc152ea3f46ced461e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationBreadcrumbBreadcrumbMeasureViewQuery$variables = {
  measureId: string;
};
export type OrganizationBreadcrumbBreadcrumbMeasureViewQuery$data = {
  readonly measure: {
    readonly category?: string;
    readonly id: string;
    readonly name?: string;
  };
};
export type OrganizationBreadcrumbBreadcrumbMeasureViewQuery = {
  response: OrganizationBreadcrumbBreadcrumbMeasureViewQuery$data;
  variables: OrganizationBreadcrumbBreadcrumbMeasureViewQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "measureId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "measureId"
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
      "name": "name",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "category",
      "storageKey": null
    }
  ],
  "type": "Measure",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationBreadcrumbBreadcrumbMeasureViewQuery",
    "selections": [
      {
        "alias": "measure",
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
    "name": "OrganizationBreadcrumbBreadcrumbMeasureViewQuery",
    "selections": [
      {
        "alias": "measure",
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
    "cacheID": "3c48e17cf1b1191df67348c4b3fd09e4",
    "id": null,
    "metadata": {},
    "name": "OrganizationBreadcrumbBreadcrumbMeasureViewQuery",
    "operationKind": "query",
    "text": "query OrganizationBreadcrumbBreadcrumbMeasureViewQuery(\n  $measureId: ID!\n) {\n  measure: node(id: $measureId) {\n    __typename\n    id\n    ... on Measure {\n      name\n      category\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "41cebf479daa730843b47f2f9fb46a10";

export default node;
