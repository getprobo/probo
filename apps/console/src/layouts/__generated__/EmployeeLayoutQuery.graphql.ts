/**
 * @generated SignedSource<<e1da65c68225be2c9a9768b31a0bf255>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type EmployeeLayoutQuery$variables = {
  organizationId: string;
};
export type EmployeeLayoutQuery$data = {
  readonly organization: {
    readonly id?: string;
    readonly logoUrl?: string | null | undefined;
    readonly name?: string;
  };
  readonly viewer: {
    readonly id: string;
  };
};
export type EmployeeLayoutQuery = {
  response: EmployeeLayoutQuery$data;
  variables: EmployeeLayoutQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "concreteType": "Viewer",
  "kind": "LinkedField",
  "name": "viewer",
  "plural": false,
  "selections": [
    (v1/*: any*/)
  ],
  "storageKey": null
},
v3 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "logoUrl",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EmployeeLayoutQuery",
    "selections": [
      (v2/*: any*/),
      {
        "alias": "organization",
        "args": (v3/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v1/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/)
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
    "name": "EmployeeLayoutQuery",
    "selections": [
      (v2/*: any*/),
      {
        "alias": "organization",
        "args": (v3/*: any*/),
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
          (v1/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v4/*: any*/),
              (v5/*: any*/)
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
    "cacheID": "2a4f5ec8a38110f9fe87a7f83cc38612",
    "id": null,
    "metadata": {},
    "name": "EmployeeLayoutQuery",
    "operationKind": "query",
    "text": "query EmployeeLayoutQuery(\n  $organizationId: ID!\n) {\n  viewer {\n    id\n  }\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      name\n      logoUrl\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "71cd50a44823e7919089a137ca0c282e";

export default node;
