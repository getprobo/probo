/**
 * @generated SignedSource<<9136dd4c154f9e4b2ba8b91316edbf4a>>
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
export type MeasureViewDeleteControlMappingMutation$variables = {
  input: DeleteControlMeasureMappingInput;
};
export type MeasureViewDeleteControlMappingMutation$data = {
  readonly deleteControlMeasureMapping: {
    readonly success: boolean;
  };
};
export type MeasureViewDeleteControlMappingMutation = {
  response: MeasureViewDeleteControlMappingMutation$data;
  variables: MeasureViewDeleteControlMappingMutation$variables;
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
    "name": "MeasureViewDeleteControlMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureViewDeleteControlMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "8e779137f0502f324968a4c414bff2e6",
    "id": null,
    "metadata": {},
    "name": "MeasureViewDeleteControlMappingMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewDeleteControlMappingMutation(\n  $input: DeleteControlMeasureMappingInput!\n) {\n  deleteControlMeasureMapping(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "7c74e8e787022c1c51fbfc38388445e4";

export default node;
