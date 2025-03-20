/**
 * @generated SignedSource<<2a843b85994d7b675c2e0504ea8c687e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationLayoutBreadcrumbPeopleOverviewQuery$variables = {
  peopleId: string;
};
export type OrganizationLayoutBreadcrumbPeopleOverviewQuery$data = {
  readonly people: {
    readonly fullName?: string;
    readonly id: string;
  };
};
export type OrganizationLayoutBreadcrumbPeopleOverviewQuery = {
  response: OrganizationLayoutBreadcrumbPeopleOverviewQuery$data;
  variables: OrganizationLayoutBreadcrumbPeopleOverviewQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "peopleId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "peopleId"
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
      "name": "fullName",
      "storageKey": null
    }
  ],
  "type": "People",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationLayoutBreadcrumbPeopleOverviewQuery",
    "selections": [
      {
        "alias": "people",
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
    "name": "OrganizationLayoutBreadcrumbPeopleOverviewQuery",
    "selections": [
      {
        "alias": "people",
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
    "cacheID": "3f43dd20847240f85b9e4f7e1d2ae1e6",
    "id": null,
    "metadata": {},
    "name": "OrganizationLayoutBreadcrumbPeopleOverviewQuery",
    "operationKind": "query",
    "text": "query OrganizationLayoutBreadcrumbPeopleOverviewQuery(\n  $peopleId: ID!\n) {\n  people: node(id: $peopleId) {\n    __typename\n    id\n    ... on People {\n      fullName\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "715ffb6eaef0bdeec2c1a42e17d37ba9";

export default node;
