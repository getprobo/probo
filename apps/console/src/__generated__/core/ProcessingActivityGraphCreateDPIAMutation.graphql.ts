/**
 * @generated SignedSource<<85c085bec1ebdad122bbf166590c3788>>
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
      readonly canDelete: boolean;
      readonly canUpdate: boolean;
      readonly createdAt: string;
      readonly description: string | null | undefined;
      readonly id: string;
      readonly mitigations: string | null | undefined;
      readonly necessityAndProportionality: string | null | undefined;
      readonly potentialRisk: string | null | undefined;
      readonly processingActivity: {
        readonly dataProtectionImpactAssessment: {
          readonly canDelete: boolean;
          readonly canUpdate: boolean;
          readonly createdAt: string;
          readonly description: string | null | undefined;
          readonly id: string;
          readonly mitigations: string | null | undefined;
          readonly necessityAndProportionality: string | null | undefined;
          readonly potentialRisk: string | null | undefined;
          readonly residualRisk: DataProtectionImpactAssessmentResidualRisk | null | undefined;
          readonly updatedAt: string;
        } | null | undefined;
        readonly id: string;
      };
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
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "necessityAndProportionality",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "potentialRisk",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "mitigations",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "residualRisk",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v9 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:data-protection-impact-assessment:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:data-protection-impact-assessment:update\")"
},
v10 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:data-protection-impact-assessment:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:data-protection-impact-assessment:delete\")"
},
v11 = [
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
          (v1/*: any*/),
          (v2/*: any*/),
          (v3/*: any*/),
          (v4/*: any*/),
          (v5/*: any*/),
          (v6/*: any*/),
          (v7/*: any*/),
          (v8/*: any*/),
          (v9/*: any*/),
          (v10/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "ProcessingActivity",
            "kind": "LinkedField",
            "name": "processingActivity",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "DataProtectionImpactAssessment",
                "kind": "LinkedField",
                "name": "dataProtectionImpactAssessment",
                "plural": false,
                "selections": [
                  (v1/*: any*/),
                  (v2/*: any*/),
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  (v10/*: any*/)
                ],
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
    "name": "ProcessingActivityGraphCreateDPIAMutation",
    "selections": (v11/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphCreateDPIAMutation",
    "selections": (v11/*: any*/)
  },
  "params": {
    "cacheID": "ea459920c052d1188c193c0a7f7bf4bc",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateDPIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateDPIAMutation(\n  $input: CreateDataProtectionImpactAssessmentInput!\n) {\n  createDataProtectionImpactAssessment(input: $input) {\n    dataProtectionImpactAssessment {\n      id\n      description\n      necessityAndProportionality\n      potentialRisk\n      mitigations\n      residualRisk\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:data-protection-impact-assessment:update\")\n      canDelete: permission(action: \"core:data-protection-impact-assessment:delete\")\n      processingActivity {\n        id\n        dataProtectionImpactAssessment {\n          id\n          description\n          necessityAndProportionality\n          potentialRisk\n          mitigations\n          residualRisk\n          createdAt\n          updatedAt\n          canUpdate: permission(action: \"core:data-protection-impact-assessment:update\")\n          canDelete: permission(action: \"core:data-protection-impact-assessment:delete\")\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f9a989390460f5567e3661fd7e757f66";

export default node;
