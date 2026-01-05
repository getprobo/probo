/**
 * @generated SignedSource<<d4281ff7f3648b044caa41cf643f2838>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateTransferImpactAssessmentInput = {
  dataSubjects?: string | null | undefined;
  id: string;
  legalMechanism?: string | null | undefined;
  localLawRisk?: string | null | undefined;
  supplementaryMeasures?: string | null | undefined;
  transfer?: string | null | undefined;
};
export type ProcessingActivityGraphUpdateTIAMutation$variables = {
  input: UpdateTransferImpactAssessmentInput;
};
export type ProcessingActivityGraphUpdateTIAMutation$data = {
  readonly updateTransferImpactAssessment: {
    readonly transferImpactAssessment: {
      readonly createdAt: string;
      readonly dataSubjects: string | null | undefined;
      readonly id: string;
      readonly legalMechanism: string | null | undefined;
      readonly localLawRisk: string | null | undefined;
      readonly supplementaryMeasures: string | null | undefined;
      readonly transfer: string | null | undefined;
      readonly updatedAt: string;
    };
  };
};
export type ProcessingActivityGraphUpdateTIAMutation = {
  response: ProcessingActivityGraphUpdateTIAMutation$data;
  variables: ProcessingActivityGraphUpdateTIAMutation$variables;
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
    "concreteType": "UpdateTransferImpactAssessmentPayload",
    "kind": "LinkedField",
    "name": "updateTransferImpactAssessment",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TransferImpactAssessment",
        "kind": "LinkedField",
        "name": "transferImpactAssessment",
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
            "name": "dataSubjects",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "legalMechanism",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "transfer",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "localLawRisk",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "supplementaryMeasures",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
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
    "name": "ProcessingActivityGraphUpdateTIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphUpdateTIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "bf5b4de5fd040e63ff654b41532ba6fb",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphUpdateTIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphUpdateTIAMutation(\n  $input: UpdateTransferImpactAssessmentInput!\n) {\n  updateTransferImpactAssessment(input: $input) {\n    transferImpactAssessment {\n      id\n      dataSubjects\n      legalMechanism\n      transfer\n      localLawRisk\n      supplementaryMeasures\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2a94ca65e0c62ac9947027dfb60cc916";

export default node;
