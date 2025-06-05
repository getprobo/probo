/**
 * @generated SignedSource<<006f8f3dc3c4d41a23cc71a870aa571c>>
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
    readonly additionalEmailAddresses?: ReadonlyArray<string>;
    readonly fullName?: string;
    readonly id?: string;
    readonly kind?: PeopleKind;
    readonly primaryEmailAddress?: string;
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
  "name": "kind",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "additionalEmailAddresses",
  "storageKey": null
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
              (v6/*: any*/)
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
              (v6/*: any*/)
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
    "cacheID": "319f5f4ff11db061920a25bbfda43244",
    "id": null,
    "metadata": {},
    "name": "PeopleGraphNodeQuery",
    "operationKind": "query",
    "text": "query PeopleGraphNodeQuery(\n  $peopleId: ID!\n) {\n  node(id: $peopleId) {\n    __typename\n    ... on People {\n      id\n      fullName\n      primaryEmailAddress\n      kind\n      additionalEmailAddresses\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "b4e6fe403e30fbc27ae638d464ad7543";

export default node;
