/**
 * @generated SignedSource<<19ada29ea9b8f557b4fcc39f57bc3aec>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestAllAccessesInput = {
  email?: any | null | undefined;
  name?: string | null | undefined;
  trustCenterId: string;
};
export type RequestAccessDialogMutation$variables = {
  input: RequestAllAccessesInput;
};
export type RequestAccessDialogMutation$data = {
  readonly requestAllAccesses: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type RequestAccessDialogMutation = {
  response: RequestAccessDialogMutation$data;
  variables: RequestAccessDialogMutation$variables;
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
    "concreteType": "RequestAccessesPayload",
    "kind": "LinkedField",
    "name": "requestAllAccesses",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterAccess",
        "kind": "LinkedField",
        "name": "trustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
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
    "name": "RequestAccessDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RequestAccessDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "99eb2902ce5a921515d68db30f8a2189",
    "id": null,
    "metadata": {},
    "name": "RequestAccessDialogMutation",
    "operationKind": "mutation",
    "text": "mutation RequestAccessDialogMutation(\n  $input: RequestAllAccessesInput!\n) {\n  requestAllAccesses(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "cc4d5c8753438eff23833b4182dd485d";

export default node;
