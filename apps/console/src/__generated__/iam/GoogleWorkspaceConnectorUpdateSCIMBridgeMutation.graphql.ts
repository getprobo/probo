/**
 * @generated SignedSource<<41b1b8127b7840bf52b967e212c286ec>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateSCIMBridgeInput = {
  excludedUserNames: ReadonlyArray<string>;
  organizationId: string;
  scimBridgeId: string;
};
export type GoogleWorkspaceConnectorUpdateSCIMBridgeMutation$variables = {
  input: UpdateSCIMBridgeInput;
};
export type GoogleWorkspaceConnectorUpdateSCIMBridgeMutation$data = {
  readonly updateSCIMBridge: {
    readonly scimBridge: {
      readonly excludedUserNames: ReadonlyArray<string>;
      readonly id: string;
    };
  } | null | undefined;
};
export type GoogleWorkspaceConnectorUpdateSCIMBridgeMutation = {
  response: GoogleWorkspaceConnectorUpdateSCIMBridgeMutation$data;
  variables: GoogleWorkspaceConnectorUpdateSCIMBridgeMutation$variables;
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
    "concreteType": "UpdateSCIMBridgePayload",
    "kind": "LinkedField",
    "name": "updateSCIMBridge",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "SCIMBridge",
        "kind": "LinkedField",
        "name": "scimBridge",
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
            "name": "excludedUserNames",
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
    "name": "GoogleWorkspaceConnectorUpdateSCIMBridgeMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "GoogleWorkspaceConnectorUpdateSCIMBridgeMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "5b2fdb1e51ff728727c10809a89ba0ea",
    "id": null,
    "metadata": {},
    "name": "GoogleWorkspaceConnectorUpdateSCIMBridgeMutation",
    "operationKind": "mutation",
    "text": "mutation GoogleWorkspaceConnectorUpdateSCIMBridgeMutation(\n  $input: UpdateSCIMBridgeInput!\n) {\n  updateSCIMBridge(input: $input) {\n    scimBridge {\n      id\n      excludedUserNames\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5b4046c71e69bbcc16e3fc533a5467ee";

export default node;
