/**
 * @generated SignedSource<<904f3388abab4e5e84dddf0528530a46>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type UpdatePeopleInput = {
  additionalEmailAddresses?: ReadonlyArray<any> | null | undefined;
  contractEndDate?: any | null | undefined;
  contractStartDate?: any | null | undefined;
  fullName?: string | null | undefined;
  id: string;
  kind?: PeopleKind | null | undefined;
  position?: string | null | undefined;
  primaryEmailAddress?: any | null | undefined;
};
export type PeopleGraphUpdateMutation$variables = {
  input: UpdatePeopleInput;
};
export type PeopleGraphUpdateMutation$data = {
  readonly updatePeople: {
    readonly people: {
      readonly additionalEmailAddresses: ReadonlyArray<any>;
      readonly contractEndDate: any | null | undefined;
      readonly contractStartDate: any | null | undefined;
      readonly fullName: string;
      readonly id: string;
      readonly kind: PeopleKind;
      readonly position: string | null | undefined;
      readonly primaryEmailAddress: any;
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
            "name": "position",
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
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "contractStartDate",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "contractEndDate",
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
    "cacheID": "23c51259f3ff117441f9c6e11de5ba5a",
    "id": null,
    "metadata": {},
    "name": "PeopleGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleGraphUpdateMutation(\n  $input: UpdatePeopleInput!\n) {\n  updatePeople(input: $input) {\n    people {\n      id\n      fullName\n      primaryEmailAddress\n      position\n      kind\n      additionalEmailAddresses\n      contractStartDate\n      contractEndDate\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "fc0946729d6a2729aa1b7c9b2dfbdf7a";

export default node;
