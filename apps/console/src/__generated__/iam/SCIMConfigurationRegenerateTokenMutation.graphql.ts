/**
 * @generated SignedSource<<107740eb1068af82db9812255fd4fc3a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RegenerateSCIMTokenInput = {
  organizationId: string;
  scimConfigurationId: string;
};
export type SCIMConfigurationRegenerateTokenMutation$variables = {
  input: RegenerateSCIMTokenInput;
};
export type SCIMConfigurationRegenerateTokenMutation$data = {
  readonly regenerateSCIMToken: {
    readonly scimConfiguration: {
      readonly createdAt: string;
      readonly endpointUrl: string;
      readonly id: string;
      readonly updatedAt: string;
    };
    readonly token: string;
  } | null | undefined;
};
export type SCIMConfigurationRegenerateTokenMutation = {
  response: SCIMConfigurationRegenerateTokenMutation$data;
  variables: SCIMConfigurationRegenerateTokenMutation$variables;
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
    "concreteType": "RegenerateSCIMTokenPayload",
    "kind": "LinkedField",
    "name": "regenerateSCIMToken",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "SCIMConfiguration",
        "kind": "LinkedField",
        "name": "scimConfiguration",
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
            "name": "endpointUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "token",
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
    "name": "SCIMConfigurationRegenerateTokenMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SCIMConfigurationRegenerateTokenMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b32d9c1791b52aebd5592c3587db2577",
    "id": null,
    "metadata": {},
    "name": "SCIMConfigurationRegenerateTokenMutation",
    "operationKind": "mutation",
    "text": "mutation SCIMConfigurationRegenerateTokenMutation(\n  $input: RegenerateSCIMTokenInput!\n) {\n  regenerateSCIMToken(input: $input) {\n    scimConfiguration {\n      id\n      endpointUrl\n      createdAt\n      updatedAt\n    }\n    token\n  }\n}\n"
  }
};
})();

(node as any).hash = "e9cbe9f51cce2157568c86e11a5a7a9d";

export default node;
