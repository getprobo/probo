/**
 * @generated SignedSource<<17c0715666a53d67a24e1c5c860c182d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProcessingActivityDPIAResidualRisk = "HIGH" | "LOW" | "MEDIUM";
export type UpdateProcessingActivityDPIAInput = {
  description?: string | null | undefined;
  id: string;
  mitigations?: string | null | undefined;
  necessityAndProportionality?: string | null | undefined;
  potentialRisk?: string | null | undefined;
  residualRisk?: ProcessingActivityDPIAResidualRisk | null | undefined;
};
export type ProcessingActivityGraphUpdateDPIAMutation$variables = {
  input: UpdateProcessingActivityDPIAInput;
};
export type ProcessingActivityGraphUpdateDPIAMutation$data = {
  readonly updateProcessingActivityDPIA: {
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
    "concreteType": "UpdateProcessingActivityDPIAPayload",
    "kind": "LinkedField",
    "name": "updateProcessingActivityDPIA",
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
    "cacheID": "7b73761c8025d6f30a2fe6b9fdafff49",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphUpdateDPIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphUpdateDPIAMutation(\n  $input: UpdateProcessingActivityDPIAInput!\n) {\n  updateProcessingActivityDPIA(input: $input) {\n    processingActivityDpia {\n      id\n      description\n      necessityAndProportionality\n      potentialRisk\n      mitigations\n      residualRisk\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2a4a592df6f98848a7191288ceec2f18";

export default node;
