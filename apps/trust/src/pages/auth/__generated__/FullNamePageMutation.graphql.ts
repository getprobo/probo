/**
 * @generated SignedSource<<7861673355d6d647183c2a4425e9a5b9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateFullNameInput = {
  fullName: string;
};
export type FullNamePageMutation$variables = {
  input: UpdateFullNameInput;
};
export type FullNamePageMutation$data = {
  readonly updateFullName: {
    readonly success: boolean;
  } | null | undefined;
};
export type FullNamePageMutation = {
  response: FullNamePageMutation$data;
  variables: FullNamePageMutation$variables;
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
    "concreteType": "UpdateFullNamePayload",
    "kind": "LinkedField",
    "name": "updateFullName",
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
    "name": "FullNamePageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "FullNamePageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "502db67e5bab72aff71fec9a0be0eeaf",
    "id": null,
    "metadata": {},
    "name": "FullNamePageMutation",
    "operationKind": "mutation",
    "text": "mutation FullNamePageMutation(\n  $input: UpdateFullNameInput!\n) {\n  updateFullName(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "3d6352889babf98a37fe06a1bcaead15";

export default node;
