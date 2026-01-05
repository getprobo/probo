/**
 * @generated SignedSource<<d993bbef933d660af5aaac0a88d3c7f7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DataProtectionImpactAssessmentResidualRisk = "HIGH" | "LOW" | "MEDIUM";
export type CreateDataProtectionImpactAssessmentInput = {
  description?: string | null | undefined;
  mitigations?: string | null | undefined;
  necessityAndProportionality?: string | null | undefined;
  potentialRisk?: string | null | undefined;
  processingActivityId: string;
  residualRisk?: DataProtectionImpactAssessmentResidualRisk | null | undefined;
};
export type ProcessingActivityGraphCreateDPIAMutation$variables = {
  input: CreateDataProtectionImpactAssessmentInput;
};
export type ProcessingActivityGraphCreateDPIAMutation$data = {
  readonly createDataProtectionImpactAssessment: {
    readonly dataProtectionImpactAssessment: {
      readonly createdAt: string;
      readonly description: string | null | undefined;
      readonly id: string;
      readonly mitigations: string | null | undefined;
      readonly necessityAndProportionality: string | null | undefined;
      readonly potentialRisk: string | null | undefined;
      readonly residualRisk: DataProtectionImpactAssessmentResidualRisk | null | undefined;
      readonly updatedAt: string;
    };
  };
};
export type ProcessingActivityGraphCreateDPIAMutation = {
  response: ProcessingActivityGraphCreateDPIAMutation$data;
  variables: ProcessingActivityGraphCreateDPIAMutation$variables;
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
    "concreteType": "CreateDataProtectionImpactAssessmentPayload",
    "kind": "LinkedField",
    "name": "createDataProtectionImpactAssessment",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "DataProtectionImpactAssessment",
        "kind": "LinkedField",
        "name": "dataProtectionImpactAssessment",
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
            "name": "description",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "necessityAndProportionality",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "potentialRisk",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "mitigations",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "residualRisk",
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
    "name": "ProcessingActivityGraphCreateDPIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphCreateDPIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "71220501284ee2a9761f87d8909fcd00",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateDPIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateDPIAMutation(\n  $input: CreateDataProtectionImpactAssessmentInput!\n) {\n  createDataProtectionImpactAssessment(input: $input) {\n    dataProtectionImpactAssessment {\n      id\n      description\n      necessityAndProportionality\n      potentialRisk\n      mitigations\n      residualRisk\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e7c379b7972e44c5c854cb4fd486f557";

export default node;
