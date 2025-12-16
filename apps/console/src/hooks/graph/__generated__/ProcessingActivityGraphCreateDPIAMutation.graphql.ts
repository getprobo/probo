/**
 * @generated SignedSource<<e12a9622852c453daa5e8d763054fe9b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProcessingActivityDPIAResidualRisk = "HIGH" | "LOW" | "MEDIUM";
export type CreateProcessingActivityDPIAInput = {
  description?: string | null | undefined;
  mitigations?: string | null | undefined;
  necessityAndProportionality?: string | null | undefined;
  potentialRisk?: string | null | undefined;
  processingActivityId: string;
  residualRisk?: ProcessingActivityDPIAResidualRisk | null | undefined;
};
export type ProcessingActivityGraphCreateDPIAMutation$variables = {
  input: CreateProcessingActivityDPIAInput;
};
export type ProcessingActivityGraphCreateDPIAMutation$data = {
  readonly createProcessingActivityDPIA: {
    readonly processingActivityDpia: {
      readonly createdAt: any;
      readonly description: string | null | undefined;
      readonly id: string;
      readonly mitigations: string | null | undefined;
      readonly necessityAndProportionality: string | null | undefined;
      readonly potentialRisk: string | null | undefined;
      readonly residualRisk: ProcessingActivityDPIAResidualRisk | null | undefined;
      readonly updatedAt: any;
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
    "concreteType": "CreateProcessingActivityDPIAPayload",
    "kind": "LinkedField",
    "name": "createProcessingActivityDPIA",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ProcessingActivityDPIA",
        "kind": "LinkedField",
        "name": "processingActivityDpia",
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
    "cacheID": "5bdea885aa22aedfe71b19bf6781be61",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateDPIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateDPIAMutation(\n  $input: CreateProcessingActivityDPIAInput!\n) {\n  createProcessingActivityDPIA(input: $input) {\n    processingActivityDpia {\n      id\n      description\n      necessityAndProportionality\n      potentialRisk\n      mitigations\n      residualRisk\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2b7346936fb433d268f816e687ddc2a6";

export default node;
