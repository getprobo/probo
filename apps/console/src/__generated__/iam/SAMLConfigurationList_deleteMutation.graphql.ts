/**
 * @generated SignedSource<<6466caf236dccf64aadc80e85172d8e3>>
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
  connections: ReadonlyArray<string>;
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
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedSamlConfigurationId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "SAMLConfigurationList_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteSAMLConfigurationPayload",
        "kind": "LinkedField",
        "name": "deleteSAMLConfiguration",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "SAMLConfigurationList_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteSAMLConfigurationPayload",
        "kind": "LinkedField",
        "name": "deleteSAMLConfiguration",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedSamlConfigurationId",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
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

(node as any).hash = "aba27afc1e97268478c5159cedf70026";

export default node;
