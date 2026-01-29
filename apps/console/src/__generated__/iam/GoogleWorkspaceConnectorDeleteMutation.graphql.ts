/**
 * @generated SignedSource<<b7e9a7b1979b5caf109266b80803fdbf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteSCIMConfigurationInput = {
  organizationId: string;
  scimConfigurationId: string;
};
export type GoogleWorkspaceConnectorDeleteMutation$variables = {
  input: DeleteSCIMConfigurationInput;
};
export type GoogleWorkspaceConnectorDeleteMutation$data = {
  readonly deleteSCIMConfiguration: {
    readonly deletedScimConfigurationId: string;
  } | null | undefined;
};
export type GoogleWorkspaceConnectorDeleteMutation = {
  response: GoogleWorkspaceConnectorDeleteMutation$data;
  variables: GoogleWorkspaceConnectorDeleteMutation$variables;
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
    "concreteType": "DeleteSCIMConfigurationPayload",
    "kind": "LinkedField",
    "name": "deleteSCIMConfiguration",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedScimConfigurationId",
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
    "name": "GoogleWorkspaceConnectorDeleteMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "GoogleWorkspaceConnectorDeleteMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "8fa5927c728a768a19fd163d162facac",
    "id": null,
    "metadata": {},
    "name": "GoogleWorkspaceConnectorDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation GoogleWorkspaceConnectorDeleteMutation(\n  $input: DeleteSCIMConfigurationInput!\n) {\n  deleteSCIMConfiguration(input: $input) {\n    deletedScimConfigurationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "fc1502ef215040eda187b2f782adb530";

export default node;
