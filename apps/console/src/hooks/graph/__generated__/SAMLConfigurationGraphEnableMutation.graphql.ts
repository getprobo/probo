/**
 * @generated SignedSource<<8bcd8aad334a74db810b3d4c17fbb6ac>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type EnableSAMLInput = {
  id: string;
};
export type SAMLConfigurationGraphEnableMutation$variables = {
  input: EnableSAMLInput;
};
export type SAMLConfigurationGraphEnableMutation$data = {
  readonly enableSAML: {
    readonly samlConfiguration: {
      readonly enabled: boolean;
      readonly id: string;
    };
  };
};
export type SAMLConfigurationGraphEnableMutation = {
  response: SAMLConfigurationGraphEnableMutation$data;
  variables: SAMLConfigurationGraphEnableMutation$variables;
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
    "concreteType": "EnableSAMLPayload",
    "kind": "LinkedField",
    "name": "enableSAML",
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
    "name": "SAMLConfigurationGraphEnableMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SAMLConfigurationGraphEnableMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2bd2356d146a9fa5f5a5f5480d310bc9",
    "id": null,
    "metadata": {},
    "name": "SAMLConfigurationGraphEnableMutation",
    "operationKind": "mutation",
    "text": "mutation SAMLConfigurationGraphEnableMutation(\n  $input: EnableSAMLInput!\n) {\n  enableSAML(input: $input) {\n    samlConfiguration {\n      id\n      enabled\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1627602d91776f718f1913011bdce786";

export default node;
