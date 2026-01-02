/**
 * @generated SignedSource<<6c0c15f8a5b275576ebadb02fc21b02e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UnlinkStateOfApplicabilityControlInput = {
  controlId: string;
  stateOfApplicabilityId: string;
};
export type StateOfApplicabilityControlsTabUnlinkMutation$variables = {
  input: UnlinkStateOfApplicabilityControlInput;
};
export type StateOfApplicabilityControlsTabUnlinkMutation$data = {
  readonly unlinkStateOfApplicabilityControl: {
    readonly deletedControlId: string;
  };
};
export type StateOfApplicabilityControlsTabUnlinkMutation = {
  response: StateOfApplicabilityControlsTabUnlinkMutation$data;
  variables: StateOfApplicabilityControlsTabUnlinkMutation$variables;
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
    "concreteType": "UnlinkStateOfApplicabilityControlPayload",
    "kind": "LinkedField",
    "name": "unlinkStateOfApplicabilityControl",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedControlId",
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
    "name": "StateOfApplicabilityControlsTabUnlinkMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "StateOfApplicabilityControlsTabUnlinkMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "4798af92da17215cc5f07c3fc57b13f9",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityControlsTabUnlinkMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityControlsTabUnlinkMutation(\n  $input: UnlinkStateOfApplicabilityControlInput!\n) {\n  unlinkStateOfApplicabilityControl(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "7a9a6b717b733984da701cd7fbc9fa39";

export default node;
