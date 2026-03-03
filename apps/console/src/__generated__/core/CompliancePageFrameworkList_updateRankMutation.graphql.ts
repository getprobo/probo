/**
 * @generated SignedSource<<3c6052dff7f71548026fb23353e215b2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateComplianceFrameworkInput = {
  id: string;
  rank: number;
};
export type CompliancePageFrameworkList_updateRankMutation$variables = {
  input: UpdateComplianceFrameworkInput;
};
export type CompliancePageFrameworkList_updateRankMutation$data = {
  readonly updateComplianceFramework: {
    readonly complianceFramework: {
      readonly id: string;
      readonly rank: number;
    };
  };
};
export type CompliancePageFrameworkList_updateRankMutation = {
  response: CompliancePageFrameworkList_updateRankMutation$data;
  variables: CompliancePageFrameworkList_updateRankMutation$variables;
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
    "concreteType": "UpdateComplianceFrameworkPayload",
    "kind": "LinkedField",
    "name": "updateComplianceFramework",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ComplianceFramework",
        "kind": "LinkedField",
        "name": "complianceFramework",
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
            "name": "rank",
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
    "name": "CompliancePageFrameworkList_updateRankMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageFrameworkList_updateRankMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "429989dd7550f3c5466a8098f5167592",
    "id": null,
    "metadata": {},
    "name": "CompliancePageFrameworkList_updateRankMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageFrameworkList_updateRankMutation(\n  $input: UpdateComplianceFrameworkInput!\n) {\n  updateComplianceFramework(input: $input) {\n    complianceFramework {\n      id\n      rank\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "a60d166f55bbe09a53349b400355ffc0";

export default node;
