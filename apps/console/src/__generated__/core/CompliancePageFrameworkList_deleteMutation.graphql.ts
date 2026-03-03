/**
 * @generated SignedSource<<0f29d441e95842685f0311839d3b45ef>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteComplianceFrameworkInput = {
  id: string;
};
export type CompliancePageFrameworkList_deleteMutation$variables = {
  input: DeleteComplianceFrameworkInput;
};
export type CompliancePageFrameworkList_deleteMutation$data = {
  readonly deleteComplianceFramework: {
    readonly deletedComplianceFrameworkId: string;
  };
};
export type CompliancePageFrameworkList_deleteMutation = {
  response: CompliancePageFrameworkList_deleteMutation$data;
  variables: CompliancePageFrameworkList_deleteMutation$variables;
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
    "concreteType": "DeleteComplianceFrameworkPayload",
    "kind": "LinkedField",
    "name": "deleteComplianceFramework",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedComplianceFrameworkId",
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
    "name": "CompliancePageFrameworkList_deleteMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageFrameworkList_deleteMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "5792168e44352169568a1378b621df16",
    "id": null,
    "metadata": {},
    "name": "CompliancePageFrameworkList_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageFrameworkList_deleteMutation(\n  $input: DeleteComplianceFrameworkInput!\n) {\n  deleteComplianceFramework(input: $input) {\n    deletedComplianceFrameworkId\n  }\n}\n"
  }
};
})();

(node as any).hash = "fed45c59adf7d88aec4bbb7eb61840c3";

export default node;
