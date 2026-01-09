/**
 * @generated SignedSource<<12ead3fad7a4f9ec7a79e1e23bcf9b17>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateStateOfApplicabilityInput = {
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
};
export type StateOfApplicabilityGraphUpdateMutation$variables = {
  input: UpdateStateOfApplicabilityInput;
};
export type StateOfApplicabilityGraphUpdateMutation$data = {
  readonly updateStateOfApplicability: {
    readonly stateOfApplicability: {
      readonly createdAt: any;
      readonly description: string | null | undefined;
      readonly id: string;
      readonly name: string;
      readonly snapshotId: string | null | undefined;
      readonly sourceId: string | null | undefined;
      readonly updatedAt: any;
    };
  };
};
export type StateOfApplicabilityGraphUpdateMutation = {
  response: StateOfApplicabilityGraphUpdateMutation$data;
  variables: StateOfApplicabilityGraphUpdateMutation$variables;
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
    "concreteType": "UpdateStateOfApplicabilityPayload",
    "kind": "LinkedField",
    "name": "updateStateOfApplicability",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "StateOfApplicability",
        "kind": "LinkedField",
        "name": "stateOfApplicability",
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
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "description",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "sourceId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "snapshotId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
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
    "name": "StateOfApplicabilityGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "StateOfApplicabilityGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "75b770c34f28b7596c54d42ef18e5a6d",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityGraphUpdateMutation(\n  $input: UpdateStateOfApplicabilityInput!\n) {\n  updateStateOfApplicability(input: $input) {\n    stateOfApplicability {\n      id\n      name\n      description\n      sourceId\n      snapshotId\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f86bd65bb8e47ac4f88d0f0cd1454080";

export default node;
