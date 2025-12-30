/**
 * @generated SignedSource<<d3c71290624f4f1394422723aba9d130>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateTransferImpactAssessmentInput = {
  dataSubjects?: string | null | undefined;
  legalMechanism?: string | null | undefined;
  localLawRisk?: string | null | undefined;
  processingActivityId: string;
  supplementaryMeasures?: string | null | undefined;
  transfer?: string | null | undefined;
};
export type ProcessingActivityGraphCreateTIAMutation$variables = {
  input: CreateTransferImpactAssessmentInput;
};
export type ProcessingActivityGraphCreateTIAMutation$data = {
  readonly createTransferImpactAssessment: {
    readonly transferImpactAssessment: {
      readonly createdAt: any;
      readonly dataSubjects: string | null | undefined;
      readonly id: string;
      readonly legalMechanism: string | null | undefined;
      readonly localLawRisk: string | null | undefined;
      readonly supplementaryMeasures: string | null | undefined;
      readonly transfer: string | null | undefined;
      readonly updatedAt: any;
    };
  };
};
export type ProcessingActivityGraphCreateTIAMutation = {
  response: ProcessingActivityGraphCreateTIAMutation$data;
  variables: ProcessingActivityGraphCreateTIAMutation$variables;
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
    "concreteType": "CreateTransferImpactAssessmentPayload",
    "kind": "LinkedField",
    "name": "createTransferImpactAssessment",
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
    "name": "ProcessingActivityGraphCreateTIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphCreateTIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "70e30255cf15b77fd95a57d6ddd50c97",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateTIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateTIAMutation(\n  $input: CreateTransferImpactAssessmentInput!\n) {\n  createTransferImpactAssessment(input: $input) {\n    transferImpactAssessment {\n      id\n      dataSubjects\n      legalMechanism\n      transfer\n      localLawRisk\n      supplementaryMeasures\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "43e9f83c7abb5d6f4d2a394c05940fbd";

export default node;
