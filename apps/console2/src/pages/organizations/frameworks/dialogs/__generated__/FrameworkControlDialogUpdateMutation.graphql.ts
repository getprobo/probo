/**
 * @generated SignedSource<<250267db99194edcb6fa20053d1d82df>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type UpdateControlInput = {
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  sectionTitle?: string | null | undefined;
};
export type FrameworkControlDialogUpdateMutation$variables = {
  input: UpdateControlInput;
};
export type FrameworkControlDialogUpdateMutation$data = {
  readonly updateControl: {
    readonly control: {
      readonly " $fragmentSpreads": FragmentRefs<"FrameworkControlDialogFragment">;
    };
  };
};
export type FrameworkControlDialogUpdateMutation = {
  response: FrameworkControlDialogUpdateMutation$data;
  variables: FrameworkControlDialogUpdateMutation$variables;
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
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkControlDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
                "args": null,
                "kind": "FragmentSpread",
                "name": "FrameworkControlDialogFragment"
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "FrameworkControlDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
    ]
  },
  "params": {
    "cacheID": "442b248ff35d4448c66052906ea5fff1",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlDialogUpdateMutation(\n  $input: UpdateControlInput!\n) {\n  updateControl(input: $input) {\n    control {\n      ...FrameworkControlDialogFragment\n      id\n    }\n  }\n}\n\nfragment FrameworkControlDialogFragment on Control {\n  id\n  name\n  description\n  sectionTitle\n}\n"
  }
};
})();

(node as any).hash = "0f070ad6868861c25e409d3777c247dd";

export default node;
