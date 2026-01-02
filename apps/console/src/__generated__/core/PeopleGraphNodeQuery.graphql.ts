/**
 * @generated SignedSource<<547fca558bc3b4667355cef5d830c9ca>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type PeopleGraphNodeQuery$variables = {
  peopleId: string;
};
export type PeopleGraphNodeQuery$data = {
  readonly node: {
    readonly additionalEmailAddresses?: ReadonlyArray<any>;
    readonly canDelete?: boolean;
    readonly canUpdate?: boolean;
    readonly contractEndDate?: any | null | undefined;
    readonly contractStartDate?: any | null | undefined;
    readonly fullName?: string;
    readonly id?: string;
    readonly kind?: PeopleKind;
    readonly position?: string | null | undefined;
    readonly primaryEmailAddress?: any;
  };
};
export type PeopleGraphNodeQuery = {
  response: PeopleGraphNodeQuery$data;
  variables: PeopleGraphNodeQuery$variables;
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
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "primaryEmailAddress",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "position",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "kind",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "additionalEmailAddresses",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "contractStartDate",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "contractEndDate",
  "storageKey": null
},
v10 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:people:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:people:update\")"
},
v11 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:people:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:people:delete\")"
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PeopleGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/)
            ],
            "type": "People",
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
    "name": "PeopleGraphNodeQuery",
    "selections": [
      {
        "alias": null,
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
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/)
            ],
            "type": "People",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "b770515fd005b877138eaddfa7c1bd2f",
    "id": null,
    "metadata": {},
    "name": "PeopleGraphNodeQuery",
    "operationKind": "query",
    "text": "query PeopleGraphNodeQuery(\n  $peopleId: ID!\n) {\n  node(id: $peopleId) {\n    __typename\n    ... on People {\n      id\n      fullName\n      primaryEmailAddress\n      position\n      kind\n      additionalEmailAddresses\n      contractStartDate\n      contractEndDate\n      canUpdate: permission(action: \"core:people:update\")\n      canDelete: permission(action: \"core:people:delete\")\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "a2bbe9f5064287c6a6e193c948a2c0cc";

export default node;
