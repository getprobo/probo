/**
 * @generated SignedSource<<5819ea12b0c6baff2a21d28b6d22bf29>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteEvidenceInput = {
  evidenceId: string;
};
export type MeasureViewDeleteEvidenceMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteEvidenceInput;
};
export type MeasureViewDeleteEvidenceMutation$data = {
  readonly deleteEvidence: {
    readonly deletedEvidenceId: string;
  };
};
export type MeasureViewDeleteEvidenceMutation = {
  response: MeasureViewDeleteEvidenceMutation$data;
  variables: MeasureViewDeleteEvidenceMutation$variables;
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
  "name": "deletedEvidenceId",
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
    "name": "MeasureViewDeleteEvidenceMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteEvidencePayload",
        "kind": "LinkedField",
        "name": "deleteEvidence",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "MeasureViewDeleteEvidenceMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteEvidencePayload",
        "kind": "LinkedField",
        "name": "deleteEvidence",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedEvidenceId",
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
    "cacheID": "26fbd71aca06943318c6154a0079d6d7",
    "id": null,
    "metadata": {},
    "name": "MeasureViewDeleteEvidenceMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewDeleteEvidenceMutation(\n  $input: DeleteEvidenceInput!\n) {\n  deleteEvidence(input: $input) {\n    deletedEvidenceId\n  }\n}\n"
  }
};
})();

(node as any).hash = "1121dd991a7505864bcd00f06065f2eb";

export default node;
