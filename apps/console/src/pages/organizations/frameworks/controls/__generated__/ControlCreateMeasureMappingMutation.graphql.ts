/**
 * @generated SignedSource<<369cb72e422d254a5e4956d17c47a8ae>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateControlMeasureMappingInput = {
  controlId: string;
  measureId: string;
};
export type ControlCreateMeasureMappingMutation$variables = {
  input: CreateControlMeasureMappingInput;
};
export type ControlCreateMeasureMappingMutation$data = {
  readonly createControlMeasureMapping: {
    readonly success: boolean;
  };
};
export type ControlCreateMeasureMappingMutation = {
  response: ControlCreateMeasureMappingMutation$data;
  variables: ControlCreateMeasureMappingMutation$variables;
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
    "concreteType": "CreateControlMeasureMappingPayload",
    "kind": "LinkedField",
    "name": "createControlMeasureMapping",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "success",
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
    "name": "ControlCreateMeasureMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ControlCreateMeasureMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "5ff3b52a5f40607556617a17e44dc3b1",
    "id": null,
    "metadata": {},
    "name": "ControlCreateMeasureMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ControlCreateMeasureMappingMutation(\n  $input: CreateControlMeasureMappingInput!\n) {\n  createControlMeasureMapping(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "2c0061b6e180971d0e0f8b8b5c44d566";

export default node;
