/**
 * @generated SignedSource<<b2635394b27e7beb548eef560262d82d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type UpdatePeopleInput = {
  additionalEmailAddresses?: ReadonlyArray<string> | null | undefined;
  expectedVersion: number;
  fullName?: string | null | undefined;
  id: string;
  kind?: PeopleKind | null | undefined;
  position?: string | null | undefined;
  primaryEmailAddress?: string | null | undefined;
};
export type PeopleViewPageUpdatePeopleMutation$variables = {
  input: UpdatePeopleInput;
};
export type PeopleViewPageUpdatePeopleMutation$data = {
  readonly updatePeople: {
    readonly people: {
      readonly additionalEmailAddresses: ReadonlyArray<string>;
      readonly fullName: string;
      readonly id: string;
      readonly kind: PeopleKind;
      readonly position: string;
      readonly primaryEmailAddress: string;
      readonly updatedAt: string;
      readonly version: number;
    };
  };
};
export type PeopleViewPageUpdatePeopleMutation = {
  response: PeopleViewPageUpdatePeopleMutation$data;
  variables: PeopleViewPageUpdatePeopleMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdatePeoplePayload",
    "kind": "LinkedField",
    "name": "updatePeople",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "People",
        "kind": "LinkedField",
        "name": "people",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "fullName",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "primaryEmailAddress",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "additionalEmailAddresses",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "kind",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "position",
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
            "name": "version",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PeopleViewPageUpdatePeopleMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PeopleViewPageUpdatePeopleMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "a054d1a1fa3ff51c0452773e9a1cedf7",
    "id": null,
    "metadata": {},
    "name": "PeopleViewPageUpdatePeopleMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleViewPageUpdatePeopleMutation(\n  $input: UpdatePeopleInput!\n) {\n  updatePeople(input: $input) {\n    people {\n      id\n      fullName\n      primaryEmailAddress\n      additionalEmailAddresses\n      kind\n      position\n      updatedAt\n      version\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "607684c9ed7f882cebf22f67caabb05d";

export default node;
