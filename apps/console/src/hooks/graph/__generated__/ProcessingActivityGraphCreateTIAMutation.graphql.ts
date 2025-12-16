/**
 * @generated SignedSource<<3da7aca4f891de8057947eda276a017f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateProcessingActivityTIAInput = {
  dataSubjects?: string | null | undefined;
  legalMechanism?: string | null | undefined;
  localLawRisk?: string | null | undefined;
  processingActivityId: string;
  supplementaryMeasures?: string | null | undefined;
  transfer?: string | null | undefined;
};
export type ProcessingActivityGraphCreateTIAMutation$variables = {
  input: CreateProcessingActivityTIAInput;
};
export type ProcessingActivityGraphCreateTIAMutation$data = {
  readonly createProcessingActivityTIA: {
    readonly processingActivityTia: {
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
    "concreteType": "CreateProcessingActivityTIAPayload",
    "kind": "LinkedField",
    "name": "createProcessingActivityTIA",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ProcessingActivityTIA",
        "kind": "LinkedField",
        "name": "processingActivityTia",
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
    "cacheID": "f04e4606a7ac8b57f6155cf5e2fa3eaa",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateTIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateTIAMutation(\n  $input: CreateProcessingActivityTIAInput!\n) {\n  createProcessingActivityTIA(input: $input) {\n    processingActivityTia {\n      id\n      dataSubjects\n      legalMechanism\n      transfer\n      localLawRisk\n      supplementaryMeasures\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e6315a3ae7db66c15850cfdae5927c03";

export default node;
