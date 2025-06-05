/**
 * @generated SignedSource<<2e60fc1d41d307f288d47b07d4651997>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateControlInput = {
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  referenceId?: string | null | undefined;
};
export type EditControlPageUpdateControlMutation$variables = {
  input: UpdateControlInput;
};
export type EditControlPageUpdateControlMutation$data = {
  readonly updateControl: {
    readonly control: {
      readonly description: string;
      readonly id: string;
      readonly name: string;
      readonly referenceId: string;
    };
  };
};
export type EditControlPageUpdateControlMutation = {
  response: EditControlPageUpdateControlMutation$data;
  variables: EditControlPageUpdateControlMutation$variables;
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
    "concreteType": "UpdateControlPayload",
    "kind": "LinkedField",
    "name": "updateControl",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Control",
        "kind": "LinkedField",
        "name": "control",
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
            "name": "referenceId",
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
    "name": "EditControlPageUpdateControlMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditControlPageUpdateControlMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "62ebcc3fcdb6bf1b1c585227e6555095",
    "id": null,
    "metadata": {},
    "name": "EditControlPageUpdateControlMutation",
    "operationKind": "mutation",
    "text": "mutation EditControlPageUpdateControlMutation(\n  $input: UpdateControlInput!\n) {\n  updateControl(input: $input) {\n    control {\n      id\n      name\n      description\n      referenceId\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2a3daa12db6c6ceac41ffb25be14760f";

export default node;
