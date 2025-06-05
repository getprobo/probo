/**
 * @generated SignedSource<<b8d91f33ada88a0656ae9637ebae06a3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateControlInput = {
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  sectionTitle?: string | null | undefined;
};
export type EditControlViewUpdateControlMutation$variables = {
  input: UpdateControlInput;
};
export type EditControlViewUpdateControlMutation$data = {
  readonly updateControl: {
    readonly control: {
      readonly description: string;
      readonly id: string;
      readonly name: string;
      readonly sectionTitle: string;
    };
  };
};
export type EditControlViewUpdateControlMutation = {
  response: EditControlViewUpdateControlMutation$data;
  variables: EditControlViewUpdateControlMutation$variables;
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
    "concreteType": "UpdateControlPayload",
    "kind": "LinkedField",
    "name": "updateControl",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Control",
        "kind": "LinkedField",
        "name": "control",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "description",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "sectionTitle",
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
    "name": "EditControlViewUpdateControlMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditControlViewUpdateControlMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "266ebf922cda80bcc5d0384f37b2d3fa",
    "id": null,
    "metadata": {},
    "name": "EditControlViewUpdateControlMutation",
    "operationKind": "mutation",
    "text": "mutation EditControlViewUpdateControlMutation(\n  $input: UpdateControlInput!\n) {\n  updateControl(input: $input) {\n    control {\n      id\n      name\n      description\n      sectionTitle\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5f232d3fbab52a0bb165bfa58c09bc67";

export default node;
