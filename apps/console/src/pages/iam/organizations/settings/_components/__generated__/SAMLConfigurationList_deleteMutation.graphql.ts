/**
 * @generated SignedSource<<4ce15b49974b7e512f334f47765caaa4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteSAMLConfigurationInput = {
  organizationId: string;
  samlConfigurationId: string;
};
export type SAMLConfigurationList_deleteMutation$variables = {
  input: DeleteSAMLConfigurationInput;
};
export type SAMLConfigurationList_deleteMutation$data = {
  readonly deleteSAMLConfiguration: {
    readonly deletedSamlConfigurationId: string;
  } | null | undefined;
};
export type SAMLConfigurationList_deleteMutation = {
  response: SAMLConfigurationList_deleteMutation$data;
  variables: SAMLConfigurationList_deleteMutation$variables;
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
    "concreteType": "DeleteSAMLConfigurationPayload",
    "kind": "LinkedField",
    "name": "deleteSAMLConfiguration",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedSamlConfigurationId",
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
    "name": "SAMLConfigurationList_deleteMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SAMLConfigurationList_deleteMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "cb46bb31e5d10f0ddad38795a72b1ab1",
    "id": null,
    "metadata": {},
    "name": "SAMLConfigurationList_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation SAMLConfigurationList_deleteMutation(\n  $input: DeleteSAMLConfigurationInput!\n) {\n  deleteSAMLConfiguration(input: $input) {\n    deletedSamlConfigurationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "2ebbe76aae843f00570f7ae1848ee995";

export default node;
