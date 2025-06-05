/**
 * @generated SignedSource<<07652bb5a37a4cdf28c0df2f627ada0e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationBreadcrumbBreadcrumbControlQuery$variables = {
  controlId: string;
};
export type OrganizationBreadcrumbBreadcrumbControlQuery$data = {
  readonly control: {
    readonly id: string;
    readonly sectionTitle?: string;
  };
};
export type OrganizationBreadcrumbBreadcrumbControlQuery = {
  response: OrganizationBreadcrumbBreadcrumbControlQuery$data;
  variables: OrganizationBreadcrumbBreadcrumbControlQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "controlId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "controlId"
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
      "name": "sectionTitle",
      "storageKey": null
    }
  ],
  "type": "Control",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationBreadcrumbBreadcrumbControlQuery",
    "selections": [
      {
        "alias": "control",
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
    "name": "OrganizationBreadcrumbBreadcrumbControlQuery",
    "selections": [
      {
        "alias": "control",
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
    "cacheID": "6d3e964652b3f13d0524f2db09668d7d",
    "id": null,
    "metadata": {},
    "name": "OrganizationBreadcrumbBreadcrumbControlQuery",
    "operationKind": "query",
    "text": "query OrganizationBreadcrumbBreadcrumbControlQuery(\n  $controlId: ID!\n) {\n  control: node(id: $controlId) {\n    __typename\n    id\n    ... on Control {\n      sectionTitle\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "04f863d003032971d0490bf0d0c90fce";

export default node;
