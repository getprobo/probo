/**
 * @generated SignedSource<<d60cc1832963c81fe19cfe8a71992273>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateFrameworkInput = {
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
};
export type FrameworkFormDialogUpdateMutation$variables = {
  input: UpdateFrameworkInput;
};
export type FrameworkFormDialogUpdateMutation$data = {
  readonly updateFramework: {
    readonly framework: {
      readonly description: string | null | undefined;
      readonly id: string;
      readonly name: string;
    };
  };
};
export type FrameworkFormDialogUpdateMutation = {
  response: FrameworkFormDialogUpdateMutation$data;
  variables: FrameworkFormDialogUpdateMutation$variables;
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
    "concreteType": "UpdateFrameworkPayload",
    "kind": "LinkedField",
    "name": "updateFramework",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Framework",
        "kind": "LinkedField",
        "name": "framework",
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
    "name": "FrameworkFormDialogUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "FrameworkFormDialogUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "54eaea672c9f7a7ddc3c6ba2b0655fc7",
    "id": null,
    "metadata": {},
    "name": "FrameworkFormDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkFormDialogUpdateMutation(\n  $input: UpdateFrameworkInput!\n) {\n  updateFramework(input: $input) {\n    framework {\n      id\n      name\n      description\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b8e9b44f6fa1cb437d7fae09133d97f0";

export default node;
