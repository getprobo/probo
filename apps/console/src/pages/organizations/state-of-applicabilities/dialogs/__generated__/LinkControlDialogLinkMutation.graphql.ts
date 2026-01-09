/**
 * @generated SignedSource<<7322f278bca210f57983a5cfea8672b5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type StateOfApplicabilityControlState = "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED";
export type LinkStateOfApplicabilityControlInput = {
  controlId: string;
  exclusionJustification?: string | null | undefined;
  state: StateOfApplicabilityControlState;
  stateOfApplicabilityId: string;
};
export type LinkControlDialogLinkMutation$variables = {
  input: LinkStateOfApplicabilityControlInput;
};
export type LinkControlDialogLinkMutation$data = {
  readonly linkStateOfApplicabilityControl: {
    readonly stateOfApplicabilityControl: {
      readonly controlId: string;
      readonly exclusionJustification: string | null | undefined;
      readonly state: StateOfApplicabilityControlState;
      readonly stateOfApplicabilityId: string;
    };
  };
};
export type LinkControlDialogLinkMutation = {
  response: LinkControlDialogLinkMutation$data;
  variables: LinkControlDialogLinkMutation$variables;
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
    "concreteType": "LinkStateOfApplicabilityControlPayload",
    "kind": "LinkedField",
    "name": "linkStateOfApplicabilityControl",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "StateOfApplicabilityControl",
        "kind": "LinkedField",
        "name": "stateOfApplicabilityControl",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "stateOfApplicabilityId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "controlId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "state",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "exclusionJustification",
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
    "name": "LinkControlDialogLinkMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "LinkControlDialogLinkMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "43ef3c1e873a68d60b7184bf6ad52df6",
    "id": null,
    "metadata": {},
    "name": "LinkControlDialogLinkMutation",
    "operationKind": "mutation",
    "text": "mutation LinkControlDialogLinkMutation(\n  $input: LinkStateOfApplicabilityControlInput!\n) {\n  linkStateOfApplicabilityControl(input: $input) {\n    stateOfApplicabilityControl {\n      stateOfApplicabilityId\n      controlId\n      state\n      exclusionJustification\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7aa3b295412c271f1e55ff2bc312bc03";

export default node;
