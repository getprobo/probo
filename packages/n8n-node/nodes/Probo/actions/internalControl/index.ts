// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as linkDocumentOp from './linkDocument.operation';
import * as linkThirdPartyOp from './linkThirdParty.operation';
import * as linkTreatmentPlanOp from './linkTreatmentPlan.operation';
import * as unlinkDocumentOp from './unlinkDocument.operation';
import * as unlinkThirdPartyOp from './unlinkThirdParty.operation';
import * as unlinkTreatmentPlanOp from './unlinkTreatmentPlan.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['internalControl'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new internal control',
				action: 'Create an internal control',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete an internal control',
				action: 'Delete an internal control',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get an internal control',
				action: 'Get an internal control',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many internal controls',
				action: 'Get many internal controls',
			},
			{
				name: 'Link Document',
				value: 'linkDocument',
				description: 'Link a document to an internal control',
				action: 'Link a document to an internal control',
			},
			{
				name: 'Link Third Party',
				value: 'linkThirdParty',
				description: 'Link a third party to an internal control',
				action: 'Link a third party to an internal control',
			},
			{
				name: 'Link Treatment Plan',
				value: 'linkTreatmentPlan',
				description: 'Link a treatment plan to an internal control',
				action: 'Link a treatment plan to an internal control',
			},
			{
				name: 'Unlink Document',
				value: 'unlinkDocument',
				description: 'Unlink a document from an internal control',
				action: 'Unlink a document from an internal control',
			},
			{
				name: 'Unlink Third Party',
				value: 'unlinkThirdParty',
				description: 'Unlink a third party from an internal control',
				action: 'Unlink a third party from an internal control',
			},
			{
				name: 'Unlink Treatment Plan',
				value: 'unlinkTreatmentPlan',
				description: 'Unlink a treatment plan from an internal control',
				action: 'Unlink a treatment plan from an internal control',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing internal control',
				action: 'Update an internal control',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...linkDocumentOp.description,
	...linkThirdPartyOp.description,
	...linkTreatmentPlanOp.description,
	...unlinkDocumentOp.description,
	...unlinkThirdPartyOp.description,
	...unlinkTreatmentPlanOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkDocumentOp as linkDocument,
	unlinkDocumentOp as unlinkDocument,
	linkThirdPartyOp as linkThirdParty,
	linkTreatmentPlanOp as linkTreatmentPlan,
	unlinkThirdPartyOp as unlinkThirdParty,
	unlinkTreatmentPlanOp as unlinkTreatmentPlan,
};
