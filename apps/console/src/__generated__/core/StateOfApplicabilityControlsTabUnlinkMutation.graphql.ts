/**
 * @generated SignedSource<<a9742800c22e683eb17d46d52b7d6aa5>>
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
export type StateOfApplicabilityControlsTabUnlinkMutation$variables = {
  input: DeleteStateOfApplicabilityControlMappingInput;
};
export type StateOfApplicabilityControlsTabUnlinkMutation$data = {
  readonly deleteStateOfApplicabilityControlMapping: {
    readonly deletedControlId: string;
    readonly deletedStateOfApplicabilityControlId: string;
    readonly deletedStateOfApplicabilityId: string;
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
    "cacheID": "6bf3ac32bd032b215e3e4f48100353d7",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityControlsTabUnlinkMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityControlsTabUnlinkMutation(\n  $input: DeleteStateOfApplicabilityControlMappingInput!\n) {\n  deleteStateOfApplicabilityControlMapping(input: $input) {\n    deletedStateOfApplicabilityId\n    deletedControlId\n    deletedStateOfApplicabilityControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "b8c3e0ca44bde75468fcfab09151854d";

export default node;
