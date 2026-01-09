/**
 * @generated SignedSource<<2893f55249b2548256bf21cf2eb434c3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteDataProtectionImpactAssessmentInput = {
  dataProtectionImpactAssessmentId: string;
};
export type ProcessingActivityGraphDeleteDPIAMutation$variables = {
  input: DeleteDataProtectionImpactAssessmentInput;
};
export type ProcessingActivityGraphDeleteDPIAMutation$data = {
  readonly deleteDataProtectionImpactAssessment: {
    readonly deletedDataProtectionImpactAssessmentId: string;
  };
};
export type ProcessingActivityGraphDeleteDPIAMutation = {
  response: ProcessingActivityGraphDeleteDPIAMutation$data;
  variables: ProcessingActivityGraphDeleteDPIAMutation$variables;
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
    "concreteType": "DeleteDataProtectionImpactAssessmentPayload",
    "kind": "LinkedField",
    "name": "deleteDataProtectionImpactAssessment",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedDataProtectionImpactAssessmentId",
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
    "name": "ProcessingActivityGraphDeleteDPIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphDeleteDPIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "6124ea356c08b9d7290082252a58c924",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphDeleteDPIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphDeleteDPIAMutation(\n  $input: DeleteDataProtectionImpactAssessmentInput!\n) {\n  deleteDataProtectionImpactAssessment(input: $input) {\n    deletedDataProtectionImpactAssessmentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "d68710e18cf1c83f60e57e5c25808cf6";

export default node;
