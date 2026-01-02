/**
 * @generated SignedSource<<11122e36bb8e61bcf7ef6245a4b970da>>
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
export type EditControlDialogLinkMutation$variables = {
  input: LinkStateOfApplicabilityControlInput;
};
export type EditControlDialogLinkMutation$data = {
  readonly linkStateOfApplicabilityControl: {
    readonly stateOfApplicabilityControl: {
      readonly controlId: string;
      readonly exclusionJustification: string | null | undefined;
      readonly state: StateOfApplicabilityControlState;
      readonly stateOfApplicabilityId: string;
    };
  };
};
export type EditControlDialogLinkMutation = {
  response: EditControlDialogLinkMutation$data;
  variables: EditControlDialogLinkMutation$variables;
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
    "name": "EditControlDialogLinkMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditControlDialogLinkMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "a4d79d851ecf6a7c1327a8b693ad9504",
    "id": null,
    "metadata": {},
    "name": "EditControlDialogLinkMutation",
    "operationKind": "mutation",
    "text": "mutation EditControlDialogLinkMutation(\n  $input: LinkStateOfApplicabilityControlInput!\n) {\n  linkStateOfApplicabilityControl(input: $input) {\n    stateOfApplicabilityControl {\n      stateOfApplicabilityId\n      controlId\n      state\n      exclusionJustification\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5e9abef4e55b4c6ba247693a7455426a";

export default node;
