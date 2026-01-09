/**
 * @generated SignedSource<<087b629e6303e83a8b9fe4ab86182bb3>>
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
export type LinkControlDialogUnlinkMutation$variables = {
  input: UnlinkStateOfApplicabilityControlInput;
};
export type LinkControlDialogUnlinkMutation$data = {
  readonly unlinkStateOfApplicabilityControl: {
    readonly deletedControlId: string;
  };
};
export type LinkControlDialogUnlinkMutation = {
  response: LinkControlDialogUnlinkMutation$data;
  variables: LinkControlDialogUnlinkMutation$variables;
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
    "name": "LinkControlDialogUnlinkMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "LinkControlDialogUnlinkMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "6e406cd7121d96cf32fb3e19ceba6670",
    "id": null,
    "metadata": {},
    "name": "LinkControlDialogUnlinkMutation",
    "operationKind": "mutation",
    "text": "mutation LinkControlDialogUnlinkMutation(\n  $input: UnlinkStateOfApplicabilityControlInput!\n) {\n  unlinkStateOfApplicabilityControl(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "eee9cdd4861b6fa8b1c00cab24bf292d";

export default node;
