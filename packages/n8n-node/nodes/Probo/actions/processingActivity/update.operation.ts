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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Processing Activity ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the processing activity to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the processing activity',
	},
	{
		displayName: 'Purpose',
		name: 'purpose',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The purpose of the processing activity',
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Controller',
				value: 'CONTROLLER',
			},
			{
				name: 'Processor',
				value: 'PROCESSOR',
			},
		],
		default: '',
		description: 'The role for the processing activity',
	},
	{
		displayName: 'Lawful Basis',
		name: 'lawfulBasis',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Consent',
				value: 'CONSENT',
			},
			{
				name: 'Contractual Necessity',
				value: 'CONTRACTUAL_NECESSITY',
			},
			{
				name: 'Legal Obligation',
				value: 'LEGAL_OBLIGATION',
			},
			{
				name: 'Legitimate Interest',
				value: 'LEGITIMATE_INTEREST',
			},
			{
				name: 'Public Task',
				value: 'PUBLIC_TASK',
			},
			{
				name: 'Vital Interests',
				value: 'VITAL_INTERESTS',
			},
		],
		default: '',
		description: 'The lawful basis for the processing activity',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const purpose = this.getNodeParameter('purpose', itemIndex, '') as string;
	const role = this.getNodeParameter('role', itemIndex, '') as string;
	const lawfulBasis = this.getNodeParameter('lawfulBasis', itemIndex, '') as string;

	const query = `
		mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
			updateProcessingActivity(input: $input) {
				processingActivity {
					id
					name
					purpose
					role
					lawfulBasis
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string> = { id };
	if (name) input.name = name;
	if (purpose) input.purpose = purpose;
	if (role) input.role = role;
	if (lawfulBasis) input.lawfulBasis = lawfulBasis;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
