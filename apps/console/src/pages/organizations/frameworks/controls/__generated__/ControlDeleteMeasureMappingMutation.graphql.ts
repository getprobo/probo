/**
 * @generated SignedSource<<6208e6173d03bbf497b38e0f8dece2a7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlMeasureMappingInput = {
  controlId: string;
  measureId: string;
};
export type ControlDeleteMeasureMappingMutation$variables = {
  input: DeleteControlMeasureMappingInput;
};
export type ControlDeleteMeasureMappingMutation$data = {
  readonly deleteControlMeasureMapping: {
    readonly success: boolean;
  };
};
export type ControlDeleteMeasureMappingMutation = {
  response: ControlDeleteMeasureMappingMutation$data;
  variables: ControlDeleteMeasureMappingMutation$variables;
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
    "concreteType": "DeleteControlMeasureMappingPayload",
    "kind": "LinkedField",
    "name": "deleteControlMeasureMapping",
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
    "name": "ControlDeleteMeasureMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ControlDeleteMeasureMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "446e0d6cc32e0c0dfc4f5e74e11bdb84",
    "id": null,
    "metadata": {},
    "name": "ControlDeleteMeasureMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ControlDeleteMeasureMappingMutation(\n  $input: DeleteControlMeasureMappingInput!\n) {\n  deleteControlMeasureMapping(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "f6a94c260a98dffe7d48e4aa9e690845";

export default node;
