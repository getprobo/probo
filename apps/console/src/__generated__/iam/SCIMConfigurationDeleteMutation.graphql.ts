/**
 * @generated SignedSource<<6c49f0f68abab263d483a28d103c04a0>>
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
export type SCIMConfigurationDeleteMutation$variables = {
  input: DeleteSCIMConfigurationInput;
};
export type SCIMConfigurationDeleteMutation$data = {
  readonly deleteSCIMConfiguration: {
    readonly deletedScimConfigurationId: string;
  } | null | undefined;
};
export type SCIMConfigurationDeleteMutation = {
  response: SCIMConfigurationDeleteMutation$data;
  variables: SCIMConfigurationDeleteMutation$variables;
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
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedScimConfigurationId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SCIMConfigurationDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "DeleteSCIMConfigurationPayload",
        "kind": "LinkedField",
        "name": "deleteSCIMConfiguration",
        "plural": false,
        "selections": [
          (v2/*: any*/)
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
    "name": "SCIMConfigurationDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "DeleteSCIMConfigurationPayload",
        "kind": "LinkedField",
        "name": "deleteSCIMConfiguration",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteRecord",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedScimConfigurationId"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1a275300c9ccf4c15d6042f132fd4527",
    "id": null,
    "metadata": {},
    "name": "SCIMConfigurationDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation SCIMConfigurationDeleteMutation(\n  $input: DeleteSCIMConfigurationInput!\n) {\n  deleteSCIMConfiguration(input: $input) {\n    deletedScimConfigurationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "6241b16c2db7d3b5d575c5ef702268cd";

export default node;
