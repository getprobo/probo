/**
 * @generated SignedSource<<2101e8d4c307eb6c37e83c625ca7801f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DisableSAMLInput = {
  id: string;
};
export type SAMLConfigurationGraphDisableMutation$variables = {
  input: DisableSAMLInput;
};
export type SAMLConfigurationGraphDisableMutation$data = {
  readonly disableSAML: {
    readonly samlConfiguration: {
      readonly enabled: boolean;
      readonly id: string;
    };
  };
};
export type SAMLConfigurationGraphDisableMutation = {
  response: SAMLConfigurationGraphDisableMutation$data;
  variables: SAMLConfigurationGraphDisableMutation$variables;
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
    "concreteType": "DisableSAMLPayload",
    "kind": "LinkedField",
    "name": "disableSAML",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "SAMLConfiguration",
        "kind": "LinkedField",
        "name": "samlConfiguration",
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
            "name": "enabled",
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
    "name": "SAMLConfigurationGraphDisableMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SAMLConfigurationGraphDisableMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "dbed05d465035d865dfd1f6473f88401",
    "id": null,
    "metadata": {},
    "name": "SAMLConfigurationGraphDisableMutation",
    "operationKind": "mutation",
    "text": "mutation SAMLConfigurationGraphDisableMutation(\n  $input: DisableSAMLInput!\n) {\n  disableSAML(input: $input) {\n    samlConfiguration {\n      id\n      enabled\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6492fd7729c72ca805ef8f99d1399081";

export default node;
