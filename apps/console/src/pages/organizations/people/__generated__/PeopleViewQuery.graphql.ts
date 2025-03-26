/**
<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewQuery.graphql.ts
 * @generated SignedSource<<207a779cf8cbf3e42191be9825d4b653>>
=======
 * @generated SignedSource<<6b0c95489030980c012399f3d52f6bd8>>
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageQuery.graphql.ts
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type PeopleViewQuery$variables = {
  peopleId: string;
};
export type PeopleViewQuery$data = {
  readonly node: {
    readonly additionalEmailAddresses?: ReadonlyArray<string>;
    readonly createdAt?: string;
    readonly fullName?: string;
    readonly id?: string;
    readonly kind?: PeopleKind;
    readonly position?: string;
    readonly primaryEmailAddress?: string;
    readonly updatedAt?: string;
    readonly version?: number;
  };
};
export type PeopleViewQuery = {
  response: PeopleViewQuery$data;
  variables: PeopleViewQuery$variables;
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
  "name": "additionalEmailAddresses",
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
  "name": "position",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "version",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PeopleViewQuery",
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
              (v10/*: any*/)
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
    "name": "PeopleViewQuery",
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
              (v10/*: any*/)
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
<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewQuery.graphql.ts
    "cacheID": "c392876240212ba16428ae5edb843d47",
=======
    "cacheID": "c9fe694bbdb3ff98d0758602390870a4",
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageQuery.graphql.ts
    "id": null,
    "metadata": {},
    "name": "PeopleViewQuery",
    "operationKind": "query",
<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewQuery.graphql.ts
    "text": "query PeopleViewQuery(\n  $peopleId: ID!\n) {\n  node(id: $peopleId) {\n    __typename\n    ... on People {\n      id\n      fullName\n      primaryEmailAddress\n      additionalEmailAddresses\n      kind\n      createdAt\n      updatedAt\n      version\n    }\n    id\n  }\n}\n"
=======
    "text": "query PeopleOverviewPageQuery(\n  $peopleId: ID!\n) {\n  node(id: $peopleId) {\n    __typename\n    ... on People {\n      id\n      fullName\n      primaryEmailAddress\n      additionalEmailAddresses\n      kind\n      position\n      createdAt\n      updatedAt\n      version\n    }\n    id\n  }\n}\n"
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageQuery.graphql.ts
  }
};
})();

<<<<<<< HEAD:apps/console/src/pages/organizations/people/__generated__/PeopleViewQuery.graphql.ts
(node as any).hash = "4fc97d6cd7fc4590b7be7b5a79ae7ab0";
=======
(node as any).hash = "6a04371010956846b504ea9e44c54cc4";
>>>>>>> 1c7bd5f (Add people position field):apps/console/src/pages/__generated__/PeopleOverviewPageQuery.graphql.ts

export default node;
