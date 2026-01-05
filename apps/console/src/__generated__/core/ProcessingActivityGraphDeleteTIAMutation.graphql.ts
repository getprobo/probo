/**
 * @generated SignedSource<<a9005a9448ac4c0cd8a5011d0a5a9665>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteTransferImpactAssessmentInput = {
  transferImpactAssessmentId: string;
};
export type ProcessingActivityGraphDeleteTIAMutation$variables = {
  input: DeleteTransferImpactAssessmentInput;
};
export type ProcessingActivityGraphDeleteTIAMutation$data = {
  readonly deleteTransferImpactAssessment: {
    readonly deletedTransferImpactAssessmentId: string;
  };
};
export type ProcessingActivityGraphDeleteTIAMutation = {
  response: ProcessingActivityGraphDeleteTIAMutation$data;
  variables: ProcessingActivityGraphDeleteTIAMutation$variables;
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
    "concreteType": "DeleteTransferImpactAssessmentPayload",
    "kind": "LinkedField",
    "name": "deleteTransferImpactAssessment",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedTransferImpactAssessmentId",
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
    "name": "ProcessingActivityGraphDeleteTIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphDeleteTIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "0b0fec997f13e35c70d2be48f3672546",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphDeleteTIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphDeleteTIAMutation(\n  $input: DeleteTransferImpactAssessmentInput!\n) {\n  deleteTransferImpactAssessment(input: $input) {\n    deletedTransferImpactAssessmentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "66a5fd8da4d9b70da0b39ee335cb9b6f";

export default node;
