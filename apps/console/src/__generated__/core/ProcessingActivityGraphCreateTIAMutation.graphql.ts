/**
 * @generated SignedSource<<dc29d874e261d9f647f7988a70588d70>>
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
      readonly canDelete: boolean;
      readonly canUpdate: boolean;
      readonly createdAt: string;
      readonly dataSubjects: string | null | undefined;
      readonly id: string;
      readonly legalMechanism: string | null | undefined;
      readonly localLawRisk: string | null | undefined;
      readonly processingActivity: {
        readonly id: string;
        readonly transferImpactAssessment: {
          readonly canDelete: boolean;
          readonly canUpdate: boolean;
          readonly createdAt: string;
          readonly dataSubjects: string | null | undefined;
          readonly id: string;
          readonly legalMechanism: string | null | undefined;
          readonly localLawRisk: string | null | undefined;
          readonly supplementaryMeasures: string | null | undefined;
          readonly transfer: string | null | undefined;
          readonly updatedAt: string;
        } | null | undefined;
      };
      readonly supplementaryMeasures: string | null | undefined;
      readonly transfer: string | null | undefined;
      readonly updatedAt: string;
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
  "name": "dataSubjects",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "legalMechanism",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transfer",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "localLawRisk",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "supplementaryMeasures",
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
      "value": "core:transfer-impact-assessment:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:transfer-impact-assessment:update\")"
},
v10 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:transfer-impact-assessment:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:transfer-impact-assessment:delete\")"
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
                "concreteType": "TransferImpactAssessment",
                "kind": "LinkedField",
                "name": "transferImpactAssessment",
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
    "name": "ProcessingActivityGraphCreateTIAMutation",
    "selections": (v11/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphCreateTIAMutation",
    "selections": (v11/*: any*/)
  },
  "params": {
    "cacheID": "f8ebabea46a0088051fd6f379db8b729",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateTIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateTIAMutation(\n  $input: CreateTransferImpactAssessmentInput!\n) {\n  createTransferImpactAssessment(input: $input) {\n    transferImpactAssessment {\n      id\n      dataSubjects\n      legalMechanism\n      transfer\n      localLawRisk\n      supplementaryMeasures\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:transfer-impact-assessment:update\")\n      canDelete: permission(action: \"core:transfer-impact-assessment:delete\")\n      processingActivity {\n        id\n        transferImpactAssessment {\n          id\n          dataSubjects\n          legalMechanism\n          transfer\n          localLawRisk\n          supplementaryMeasures\n          createdAt\n          updatedAt\n          canUpdate: permission(action: \"core:transfer-impact-assessment:update\")\n          canDelete: permission(action: \"core:transfer-impact-assessment:delete\")\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b4bca88fcf13792cf8d03ba91de4261c";

export default node;
