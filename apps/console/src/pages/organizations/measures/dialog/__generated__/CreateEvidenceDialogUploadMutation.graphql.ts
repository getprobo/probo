/**
 * @generated SignedSource<<9297f85591b8ab4723a207f6db6a27f6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type UploadMeasureEvidenceInput = {
  file: any;
  measureId: string;
};
export type CreateEvidenceDialogUploadMutation$variables = {
  connections: ReadonlyArray<string>;
  input: UploadMeasureEvidenceInput;
};
export type CreateEvidenceDialogUploadMutation$data = {
  readonly uploadMeasureEvidence: {
    readonly evidenceEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"MeasureEvidencesTabFragment_evidence">;
      };
    };
  };
};
export type CreateEvidenceDialogUploadMutation = {
  response: CreateEvidenceDialogUploadMutation$data;
  variables: CreateEvidenceDialogUploadMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "CreateEvidenceDialogUploadMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "UploadMeasureEvidencePayload",
        "kind": "LinkedField",
        "name": "uploadMeasureEvidence",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "EvidenceEdge",
            "kind": "LinkedField",
            "name": "evidenceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Evidence",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "MeasureEvidencesTabFragment_evidence"
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
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "CreateEvidenceDialogUploadMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "UploadMeasureEvidencePayload",
        "kind": "LinkedField",
        "name": "uploadMeasureEvidence",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "EvidenceEdge",
            "kind": "LinkedField",
            "name": "evidenceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Evidence",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "filename",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "size",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "type",
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
                    "name": "mimeType",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "appendEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "evidenceEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "5206b6576cacb45f5b42305bb1bb7df1",
    "id": null,
    "metadata": {},
    "name": "CreateEvidenceDialogUploadMutation",
    "operationKind": "mutation",
    "text": "mutation CreateEvidenceDialogUploadMutation(\n  $input: UploadMeasureEvidenceInput!\n) {\n  uploadMeasureEvidence(input: $input) {\n    evidenceEdge {\n      node {\n        id\n        ...MeasureEvidencesTabFragment_evidence\n      }\n    }\n  }\n}\n\nfragment MeasureEvidencesTabFragment_evidence on Evidence {\n  id\n  filename\n  size\n  type\n  createdAt\n  mimeType\n}\n"
  }
};
})();

(node as any).hash = "719cce21cdff705c55dda3de31997a53";

export default node;
