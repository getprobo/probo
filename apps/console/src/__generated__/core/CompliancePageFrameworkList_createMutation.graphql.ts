/**
 * @generated SignedSource<<2980fa2d2a72851686e88a7b5402d720>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateComplianceFrameworkInput = {
  frameworkId: string;
  trustCenterId: string;
};
export type CompliancePageFrameworkList_createMutation$variables = {
  input: CreateComplianceFrameworkInput;
};
export type CompliancePageFrameworkList_createMutation$data = {
  readonly createComplianceFramework: {
    readonly complianceFrameworkEdge: {
      readonly node: {
        readonly id: string;
      };
    };
  };
};
export type CompliancePageFrameworkList_createMutation = {
  response: CompliancePageFrameworkList_createMutation$data;
  variables: CompliancePageFrameworkList_createMutation$variables;
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
    "concreteType": "CreateComplianceFrameworkPayload",
    "kind": "LinkedField",
    "name": "createComplianceFramework",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ComplianceFrameworkEdge",
        "kind": "LinkedField",
        "name": "complianceFrameworkEdge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ComplianceFramework",
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "id",
                "storageKey": null
              }
            ],
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
    "name": "CompliancePageFrameworkList_createMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageFrameworkList_createMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9caebfdeda321fb743a8d789c7e25e82",
    "id": null,
    "metadata": {},
    "name": "CompliancePageFrameworkList_createMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageFrameworkList_createMutation(\n  $input: CreateComplianceFrameworkInput!\n) {\n  createComplianceFramework(input: $input) {\n    complianceFrameworkEdge {\n      node {\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "41815aaef15ed1754b4d81c1d982a2c6";

export default node;
