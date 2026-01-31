/**
 * @generated SignedSource<<4b9c832eb09c31751c1da609cda983ba>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateSCIMConfigurationInput = {
  connectorId?: string | null | undefined;
  organizationId: string;
};
export type SCIMSettingsPageCreateSCIMConfigurationMutation$variables = {
  input: CreateSCIMConfigurationInput;
};
export type SCIMSettingsPageCreateSCIMConfigurationMutation$data = {
  readonly createSCIMConfiguration: {
    readonly scimBridge: {
      readonly id: string;
    } | null | undefined;
    readonly scimConfiguration: {
      readonly id: string;
    };
  } | null | undefined;
};
export type SCIMSettingsPageCreateSCIMConfigurationMutation = {
  response: SCIMSettingsPageCreateSCIMConfigurationMutation$data;
  variables: SCIMSettingsPageCreateSCIMConfigurationMutation$variables;
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
    "args": null,
    "kind": "ScalarField",
    "name": "id",
    "storageKey": null
  }
],
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "CreateSCIMConfigurationPayload",
    "kind": "LinkedField",
    "name": "createSCIMConfiguration",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "SCIMConfiguration",
        "kind": "LinkedField",
        "name": "scimConfiguration",
        "plural": false,
        "selections": (v1/*: any*/),
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "SCIMBridge",
        "kind": "LinkedField",
        "name": "scimBridge",
        "plural": false,
        "selections": (v1/*: any*/),
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
    "name": "SCIMSettingsPageCreateSCIMConfigurationMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SCIMSettingsPageCreateSCIMConfigurationMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "281c9e39b96e2e56c3f4379f8dd8dc30",
    "id": null,
    "metadata": {},
    "name": "SCIMSettingsPageCreateSCIMConfigurationMutation",
    "operationKind": "mutation",
    "text": "mutation SCIMSettingsPageCreateSCIMConfigurationMutation(\n  $input: CreateSCIMConfigurationInput!\n) {\n  createSCIMConfiguration(input: $input) {\n    scimConfiguration {\n      id\n    }\n    scimBridge {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "841dfdb5ce8a226c7e683090c75951f0";

export default node;
