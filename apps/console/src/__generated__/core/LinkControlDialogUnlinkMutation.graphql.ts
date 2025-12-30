/**
 * @generated SignedSource<<3bb4d04e228ac7dbda8ebfb8b7c2567f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteStateOfApplicabilityControlMappingInput = {
  controlId: string;
  stateOfApplicabilityId: string;
};
export type LinkControlDialogUnlinkMutation$variables = {
  input: DeleteStateOfApplicabilityControlMappingInput;
};
export type LinkControlDialogUnlinkMutation$data = {
  readonly deleteStateOfApplicabilityControlMapping: {
    readonly deletedControlId: string;
    readonly deletedStateOfApplicabilityControlId: string;
    readonly deletedStateOfApplicabilityId: string;
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
    "concreteType": "DeleteStateOfApplicabilityControlMappingPayload",
    "kind": "LinkedField",
    "name": "deleteStateOfApplicabilityControlMapping",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedStateOfApplicabilityId",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedControlId",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedStateOfApplicabilityControlId",
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
    "cacheID": "766d15d5071197ed2907d876caf37995",
    "id": null,
    "metadata": {},
    "name": "LinkControlDialogUnlinkMutation",
    "operationKind": "mutation",
    "text": "mutation LinkControlDialogUnlinkMutation(\n  $input: DeleteStateOfApplicabilityControlMappingInput!\n) {\n  deleteStateOfApplicabilityControlMapping(input: $input) {\n    deletedStateOfApplicabilityId\n    deletedControlId\n    deletedStateOfApplicabilityControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "a184d521b495bf3ff76796530e9eb0e1";

export default node;
