/**
 * @generated SignedSource<<b798b75c1d930af818109fd819d5cef3>>
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
  contractEndDate?: any | null | undefined;
  contractStartDate?: any | null | undefined;
  fullName?: string | null | undefined;
  id: string;
  kind?: PeopleKind | null | undefined;
  position?: string | null | undefined;
  primaryEmailAddress?: string | null | undefined;
};
export type PeopleGraphUpdateMutation$variables = {
  input: UpdatePeopleInput;
};
export type PeopleGraphUpdateMutation$data = {
  readonly updatePeople: {
    readonly people: {
      readonly additionalEmailAddresses: ReadonlyArray<string>;
      readonly fullName: string;
      readonly id: string;
      readonly kind: PeopleKind;
      readonly primaryEmailAddress: string;
    };
  };
};
export type PeopleGraphUpdateMutation = {
  response: PeopleGraphUpdateMutation$data;
  variables: PeopleGraphUpdateMutation$variables;
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
            "name": "kind",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "additionalEmailAddresses",
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
    "name": "PeopleGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PeopleGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "c61ca0a6b20067ce5e0c708fc55509d8",
    "id": null,
    "metadata": {},
    "name": "PeopleGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleGraphUpdateMutation(\n  $input: UpdatePeopleInput!\n) {\n  updatePeople(input: $input) {\n    people {\n      id\n      fullName\n      primaryEmailAddress\n      kind\n      additionalEmailAddresses\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "48c2618963704a8cd3b1f0a68817cdad";

export default node;
