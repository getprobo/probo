// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';
import { organizationLocator, personLocator } from '../../locators';
import { correctiveActionPlanNodeFields } from './fields';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['correctiveActionPlan'],
		operation: ['create'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	{
		displayName: 'Finding ID',
		name: 'findingId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['correctiveActionPlan'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the finding the plan addresses',
		required: true,
	},
	{
		displayName: 'Title',
		name: 'title',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['correctiveActionPlan'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The title of the corrective action plan',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['correctiveActionPlan'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Description',
				name: 'description',
				type: 'string',
				typeOptions: {
					rows: 4,
				},
				default: '',
				description: 'The description of the corrective action plan',
			},
			{
				displayName: 'Due Date',
				name: 'dueDate',
				type: 'dateTime',
				default: '',
				description: 'When the corrective action plan is due',
			},
			personLocator({
				displayName: 'Owner',
				required: false,
			}),
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const findingId = this.getNodeParameter('findingId', itemIndex) as string;
	const title = this.getNodeParameter('title', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		description?: string;
		dueDate?: string;
	};
	const ownerId = this.getNodeParameter('additionalFields.ownerId', itemIndex, '', {
		extractValue: true,
	}) as string;

	const query = `
		mutation CreateCorrectiveActionPlan($input: CreateCorrectiveActionPlanInput!) {
			createCorrectiveActionPlan(input: $input) {
				correctiveActionPlanEdge {
					node {
						${correctiveActionPlanNodeFields}
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { findingId, title };
	if (additionalFields.description) input.description = additionalFields.description;
	if (additionalFields.dueDate) input.dueDate = additionalFields.dueDate;
	if (ownerId) input.ownerId = ownerId;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
