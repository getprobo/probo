/**
 * @generated SignedSource<<55ce4f4205b3eb01954423a75e6e9369>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteSAMLConfigurationInput = {
  id: string;
};
export type SAMLConfigurationGraphDeleteMutation$variables = {
  input: DeleteSAMLConfigurationInput;
};
export type SAMLConfigurationGraphDeleteMutation$data = {
  readonly deleteSAMLConfiguration: {
    readonly deletedSAMLConfigurationId: string;
  };
};
export type SAMLConfigurationGraphDeleteMutation = {
  response: SAMLConfigurationGraphDeleteMutation$data;
  variables: SAMLConfigurationGraphDeleteMutation$variables;
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
        "name": "deletedSAMLConfigurationId",
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
    "name": "SAMLConfigurationGraphDeleteMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SAMLConfigurationGraphDeleteMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "be1192099e1b3d0765a358de10685870",
    "id": null,
    "metadata": {},
    "name": "SAMLConfigurationGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation SAMLConfigurationGraphDeleteMutation(\n  $input: DeleteSAMLConfigurationInput!\n) {\n  deleteSAMLConfiguration(input: $input) {\n    deletedSAMLConfigurationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "869072f879524c5c2acbc684f536bfe5";

export default node;
