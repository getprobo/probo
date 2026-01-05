/**
 * @generated SignedSource<<9bd2a9a518e61ab378f74e7fd2521a01>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DataProtectionImpactAssessmentResidualRisk = "HIGH" | "LOW" | "MEDIUM";
export type UpdateDataProtectionImpactAssessmentInput = {
  description?: string | null | undefined;
  id: string;
  mitigations?: string | null | undefined;
  necessityAndProportionality?: string | null | undefined;
  potentialRisk?: string | null | undefined;
  residualRisk?: DataProtectionImpactAssessmentResidualRisk | null | undefined;
};
export type ProcessingActivityGraphUpdateDPIAMutation$variables = {
  input: UpdateDataProtectionImpactAssessmentInput;
};
export type ProcessingActivityGraphUpdateDPIAMutation$data = {
  readonly updateDataProtectionImpactAssessment: {
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
export type ProcessingActivityGraphUpdateDPIAMutation = {
  response: ProcessingActivityGraphUpdateDPIAMutation$data;
  variables: ProcessingActivityGraphUpdateDPIAMutation$variables;
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
    "concreteType": "UpdateDataProtectionImpactAssessmentPayload",
    "kind": "LinkedField",
    "name": "updateDataProtectionImpactAssessment",
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
    "name": "ProcessingActivityGraphUpdateDPIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphUpdateDPIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "8209c5674c73064b76e46b67c017b5fa",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphUpdateDPIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphUpdateDPIAMutation(\n  $input: UpdateDataProtectionImpactAssessmentInput!\n) {\n  updateDataProtectionImpactAssessment(input: $input) {\n    dataProtectionImpactAssessment {\n      id\n      description\n      necessityAndProportionality\n      potentialRisk\n      mitigations\n      residualRisk\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9b03511f95fd87aafd075b66fd82184c";

export default node;
