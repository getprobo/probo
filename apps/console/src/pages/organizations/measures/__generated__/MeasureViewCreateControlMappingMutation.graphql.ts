/**
 * @generated SignedSource<<441cfaaf785e0628a8c954249ef87bdf>>
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
export type MeasureViewCreateControlMappingMutation$variables = {
  input: CreateControlMeasureMappingInput;
};
export type MeasureViewCreateControlMappingMutation$data = {
  readonly createControlMeasureMapping: {
    readonly success: boolean;
  };
};
export type MeasureViewCreateControlMappingMutation = {
  response: MeasureViewCreateControlMappingMutation$data;
  variables: MeasureViewCreateControlMappingMutation$variables;
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
    "name": "MeasureViewCreateControlMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureViewCreateControlMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "ea74e0148d320dbfb79859bb6286427b",
    "id": null,
    "metadata": {},
    "name": "MeasureViewCreateControlMappingMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewCreateControlMappingMutation(\n  $input: CreateControlMeasureMappingInput!\n) {\n  createControlMeasureMapping(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "dfe964679c82f9c47d418122595989a8";

export default node;
